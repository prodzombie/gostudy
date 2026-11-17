package review

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"charm.land/bubbles/v2/viewport"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"

	"github.com/prodzombie/gostudy/internal/cardstore"
	"github.com/prodzombie/gostudy/internal/schedule"
)

// screen owns interactive presentation; schedules are persisted before advancing.
type screen struct {
	store                cardstore.Store
	due                  []dueCard
	now                  func() time.Time
	viewport             viewport.Model
	width, height, index int
	revealed, help       bool
	status               string
	err                  error
	palette              palette
	practice             bool
}

// newViewport scrolls the card. f is removed from the default page-down keys
// because it answers "forgot".
func newViewport() viewport.Model {
	v := viewport.New()
	v.KeyMap.PageDown.SetKeys("pgdown", "space")
	return v
}

func runScreen(store cardstore.Store, due []dueCard, in io.Reader, out io.Writer, now func() time.Time, options Options) error {
	m := &screen{practice: options.Practice, store: store, due: due, now: now, viewport: newViewport(), palette: palette{dark: true, noColor: os.Getenv("NO_COLOR") != ""}}
	if _, err := tea.NewProgram(m, tea.WithInput(in), tea.WithOutput(out)).Run(); err != nil {
		return err
	}
	if m.err != nil {
		return m.err
	}
	state := "Session stopped"
	if m.index == len(due) {
		state = "Session complete"
	}
	if m.status != "" {
		if _, err := fmt.Fprintln(out, m.status); err != nil {
			return err
		}
	}
	verb := "Reviewed"
	if m.practice {
		verb = "Practiced"
	}
	_, err := fmt.Fprintf(out, "%s. %s %d of %d cards. %d remaining in this session.\n", state, verb, m.index, len(due), len(due)-m.index)
	return err
}

func (m *screen) Init() tea.Cmd {
	if m.palette.noColor {
		return nil
	}
	return tea.RequestBackgroundColor
}

func (m *screen) header() string {
	phase := "Question"
	if m.revealed {
		phase = "Answer"
	}
	prefix := ""
	if m.practice {
		prefix = "Practice · "
	}
	return prefix + m.palette.secondary().Render(fmt.Sprintf("Card %d of %d · %s", m.index+1, len(m.due), m.due[m.index].card.Tags.Value)) + " · " + m.palette.accent().Render(phase) + "\n" + m.palette.success().Render(m.status)
}

func (m *screen) footer() string {
	reading := ""
	if m.viewport.TotalLineCount() > m.viewport.Height() {
		reading = fmt.Sprintf("Scroll %.0f%%", m.viewport.ScrollPercent()*100)
	}
	controls := m.palette.accent().Render("Enter: reveal")
	if m.revealed {
		controls = m.palette.accent().Render("r: recalled · f: forgot")
	}
	if m.practice && m.revealed {
		controls = m.palette.accent().Render("Enter: next card")
	}
	help := "↑/↓ scroll · PgUp/PgDn page · ? help · q quit"
	if m.help {
		help += "\nHome/End: top/bottom · j/k: scroll · mouse wheel: scroll\nCtrl+C: quit"
	}
	return m.palette.secondary().Render(reading) + "\n" + m.palette.secondary().Render(help) + "\n" + controls
}

func (m *screen) layout() {
	width := max(1, m.width-1)
	m.viewport.SetWidth(width)
	m.viewport.SetHeight(max(1, m.height-lipgloss.Height(lipgloss.NewStyle().Width(width).Render(m.header()))-lipgloss.Height(lipgloss.NewStyle().Width(width).Render(m.footer()))-2))
}

func (m *screen) render(jumpToAnswer bool) error {
	if m.width == 0 {
		return nil
	}
	m.layout()
	r, err := themedMarkdownRenderer(max(1, min(80, m.width-1)), m.palette)
	if err != nil {
		return err
	}
	question, err := r.Render(m.due[m.index].card.Question)
	if err != nil {
		return err
	}
	content := m.palette.accent().Render("Question") + "\n\n" + strings.Trim(question, "\n")
	answerLine := strings.Count(content, "\n") + 2
	if m.revealed {
		answer, err := r.Render(m.due[m.index].card.Answer)
		if err != nil {
			return err
		}
		content += "\n\n" + m.palette.accent().Render("Answer") + "\n\n" + strings.Trim(answer, "\n")
	}
	offset := m.viewport.YOffset()
	m.viewport.SetContent(content)
	m.viewport.SetYOffset(offset)
	if jumpToAnswer {
		m.viewport.SetYOffset(answerLine)
	}
	return nil
}

func (m *screen) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.BackgroundColorMsg:
		m.palette.dark = msg.IsDark()
		m.err = m.render(false)
	case tea.WindowSizeMsg:
		m.width, m.height = msg.Width, msg.Height
		m.err = m.render(false)
	case tea.KeyPressMsg:
		key := msg.String()
		if (m.width < 24 || m.height < 12) && key != "q" && key != "ctrl+c" {
			return m, nil
		}
		switch key {
		case "q", "ctrl+c":
			return m, tea.Quit
		case "home", "g":
			m.viewport.GotoTop()
		case "end", "G":
			m.viewport.GotoBottom()
		case "?":
			m.help = !m.help
			m.layout()
		case "enter":
			if !m.revealed {
				m.revealed = true
				m.err = m.render(true)
			} else if m.practice {
				return m, m.next()
			}
		case "r", "f":
			if m.revealed && !m.practice {
				progress := schedule.Review(m.due[m.index].progress, key == "r", m.now())
				if err := m.store.SaveSchedule(m.due[m.index].card.ID.String(), progress); err != nil {
					m.err = err
					return m, tea.Quit
				}
				unit := "days"
				if progress.Interval == 1 {
					unit = "day"
				}
				grade := "forgot"
				if key == "r" {
					grade = "recalled"
				}
				m.status = fmt.Sprintf("Previous card: %s · due in %d %s", grade, progress.Interval, unit)
				return m, m.next()
			}
		default:
			var cmd tea.Cmd
			m.viewport, cmd = m.viewport.Update(msg)
			return m, cmd
		}
	default:
		var cmd tea.Cmd
		m.viewport, cmd = m.viewport.Update(msg)
		return m, cmd
	}
	if m.err != nil {
		return m, tea.Quit
	}
	return m, nil
}

// next advances to the following card's question, quitting after the last one.
func (m *screen) next() tea.Cmd {
	m.index++
	if m.index == len(m.due) {
		return tea.Quit
	}
	m.revealed = false
	m.viewport.GotoTop()
	if m.err = m.render(false); m.err != nil {
		return tea.Quit
	}
	return nil
}

func (m *screen) View() tea.View {
	v := tea.NewView("")
	v.AltScreen = true
	v.MouseMode = tea.MouseModeCellMotion
	if m.index == len(m.due) {
		return v
	}
	if m.width < 24 || m.height < 12 {
		v.SetContent("Enlarge terminal (24×12 minimum). q: quit")
		return v
	}
	style := lipgloss.NewStyle().Width(m.width - 1)
	v.SetContent(style.Render(m.header()) + "\n\n" + m.viewport.View() + "\n" + style.Render(m.footer()))
	return v
}
