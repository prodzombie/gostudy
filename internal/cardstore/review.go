package cardstore

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/prodzombie/gostudy/internal/card"
	"github.com/prodzombie/gostudy/internal/huid"
	"github.com/prodzombie/gostudy/internal/schedule"
)

var scheduleKeys = []string{"ease_factor", "interval", "repetitions", "last_reviewed"}

// Load reads a card's question, answer, and persisted schedule.
// The directory layout supplies identity and tag; unknown frontmatter is preserved.
func (s Store) Load(id string) (card.Card, schedule.Schedule, error) {
	path, err := s.Path(id)
	if err != nil {
		return card.Card{}, schedule.Schedule{}, err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return card.Card{}, schedule.Schedule{}, err
	}
	parsed, err := huid.Parse(id)
	if err != nil {
		return card.Card{}, schedule.Schedule{}, err
	}
	lines, end, err := frontmatter(string(content))
	if err != nil {
		return card.Card{}, schedule.Schedule{}, fmt.Errorf("card %s: %w", id, err)
	}
	progress, err := parseSchedule(lines[1:end], parsed)
	if err != nil {
		return card.Card{}, schedule.Schedule{}, fmt.Errorf("card %s: %w", id, err)
	}
	question, answer, err := cardSections(strings.Join(lines[end+1:], ""))
	if err != nil {
		return card.Card{}, schedule.Schedule{}, fmt.Errorf("card %s: %w", id, err)
	}
	if question == "" || answer == "" {
		return card.Card{}, schedule.Schedule{}, fmt.Errorf("card %s: question and answer must not be blank", id)
	}
	return card.Card{ID: parsed, Question: question, Answer: answer, Tags: &card.Tag{Value: filepath.Base(filepath.Dir(filepath.Dir(path)))}}, progress, nil
}

func frontmatter(content string) ([]string, int, error) {
	lines := strings.SplitAfter(content, "\n")
	if len(lines) == 0 || strings.TrimRight(lines[0], "\r\n") != "---" {
		return nil, 0, fmt.Errorf("missing YAML frontmatter")
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimRight(lines[i], "\r\n") == "---" {
			return lines, i, nil
		}
	}
	return nil, 0, fmt.Errorf("unclosed YAML frontmatter")
}

// scheduleProperty recognizes only the flat properties owned by the scheduler.
func scheduleProperty(line string) (string, string, bool) {
	for _, key := range scheduleKeys {
		if strings.HasPrefix(line, key+":") {
			return key, strings.TrimSpace(strings.TrimPrefix(line, key+":")), true
		}
	}
	return "", "", false
}

func parseSchedule(lines []string, id huid.HUID) (schedule.Schedule, error) {
	progress := schedule.Schedule{Card: id}
	values := make(map[string]string)
	for _, line := range lines {
		key, value, ok := scheduleProperty(line)
		if !ok {
			continue
		}
		if _, exists := values[key]; exists {
			return progress, fmt.Errorf("duplicate schedule property %s", key)
		}
		// Generated values are plain numbers or quoted timestamps. Accept a trailing comment too.
		if i := strings.Index(value, " #"); i >= 0 {
			value = strings.TrimSpace(value[:i])
		}
		if strings.HasPrefix(value, "\"") {
			var decoded string
			if err := json.Unmarshal([]byte(value), &decoded); err != nil {
				return progress, fmt.Errorf("invalid %s: %w", key, err)
			}
			value = decoded
		} else if len(value) >= 2 && value[0] == '\'' && value[len(value)-1] == '\'' {
			value = strings.ReplaceAll(value[1:len(value)-1], "''", "'")
		}
		values[key] = value
	}
	if len(values) == 0 {
		return progress, nil
	}
	if len(values) != len(scheduleKeys) {
		return progress, fmt.Errorf("incomplete schedule properties")
	}
	var err error
	progress.EaseFactor, err = strconv.ParseFloat(values["ease_factor"], 64)
	if err != nil {
		return progress, fmt.Errorf("invalid ease_factor: %w", err)
	}
	progress.Interval, err = strconv.Atoi(values["interval"])
	if err != nil {
		return progress, fmt.Errorf("invalid interval: %w", err)
	}
	progress.Repetitions, err = strconv.Atoi(values["repetitions"])
	if err != nil {
		return progress, fmt.Errorf("invalid repetitions: %w", err)
	}
	reviewed, err := time.Parse(time.RFC3339Nano, values["last_reviewed"])
	if err != nil {
		return progress, fmt.Errorf("invalid last_reviewed: %w", err)
	}
	progress.LastReviewed = &reviewed
	return progress, validateSchedule(progress)
}

func validateSchedule(progress schedule.Schedule) error {
	if math.IsNaN(progress.EaseFactor) || math.IsInf(progress.EaseFactor, 0) || progress.EaseFactor < 1.3 || progress.Interval < 1 || progress.Repetitions < 0 || progress.LastReviewed == nil {
		return fmt.Errorf("invalid review schedule")
	}
	return nil
}

// SaveSchedule updates only scheduling properties, preserving all other content.
// Replacement is atomic so an interrupted write cannot truncate CARD.md.
func (s Store) SaveSchedule(id string, progress schedule.Schedule) error {
	if err := validateSchedule(progress); err != nil {
		return err
	}
	path, err := s.Path(id)
	if err != nil {
		return err
	}
	content, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	lines, end, err := frontmatter(string(content))
	if err != nil {
		return err
	}
	// Reject corrupt existing schedules rather than silently discarding them.
	parsed, err := huid.Parse(id)
	if err != nil {
		return err
	}
	if _, err := parseSchedule(lines[1:end], parsed); err != nil {
		return err
	}
	newline := "\n"
	if strings.HasSuffix(lines[0], "\r\n") {
		newline = "\r\n"
	}
	var updated strings.Builder
	updated.WriteString(lines[0])
	for _, line := range lines[1:end] {
		if _, _, owned := scheduleProperty(line); !owned {
			updated.WriteString(line)
		}
	}
	for _, property := range []string{
		"ease_factor: " + strconv.FormatFloat(progress.EaseFactor, 'g', -1, 64),
		"interval: " + strconv.Itoa(progress.Interval),
		"repetitions: " + strconv.Itoa(progress.Repetitions),
		"last_reviewed: " + yamlString(progress.LastReviewed.UTC().Format(time.RFC3339Nano)),
	} {
		updated.WriteString(property + newline)
	}
	updated.WriteString(strings.Join(lines[end:], ""))
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".review-*")
	if err != nil {
		return err
	}
	defer os.Remove(temp.Name())
	if err := temp.Chmod(info.Mode().Perm()); err != nil {
		temp.Close()
		return err
	}
	if _, err := temp.WriteString(updated.String()); err != nil {
		temp.Close()
		return err
	}
	if err := temp.Close(); err != nil {
		return err
	}
	return os.Rename(temp.Name(), path)
}

// cardSections recognizes the generated section headings outside fenced code blocks.
func cardSections(body string) (string, string, error) {
	lines := strings.SplitAfter(body, "\n")
	start := -1
	var fence byte
	fenceLength := 0
	for i, line := range lines {
		text := strings.TrimRight(line, "\r\n")
		trimmed := strings.TrimLeft(text, " ")
		if len(text)-len(trimmed) <= 3 && len(trimmed) >= 3 && (trimmed[0] == '`' || trimmed[0] == '~') {
			count := 0
			for count < len(trimmed) && trimmed[count] == trimmed[0] {
				count++
			}
			if count >= 3 {
				if fence == 0 {
					fence, fenceLength = trimmed[0], count
				} else if fence == trimmed[0] && count >= fenceLength && strings.TrimSpace(trimmed[count:]) == "" {
					fence = 0
				}
				continue
			}
		}
		if fence != 0 {
			continue
		}
		if start < 0 && text == "## Question" {
			start = i + 1
			continue
		}
		if start >= 0 && text == "## Answer" {
			question := strings.TrimSpace(strings.Join(lines[start:i], ""))
			answer := strings.TrimSpace(strings.Join(lines[i+1:], ""))
			return question, answer, nil
		}
	}
	return "", "", fmt.Errorf("missing Question or Answer section")
}

// Question reads the question section, including blank cards awaiting editing.
func (s Store) Question(entry Entry) (string, error) {
	id := entry.ID
	if !huid.Valid(id) || !card.ValidTag(entry.Tag) {
		return "", fmt.Errorf("invalid card entry")
	}
	content, err := os.ReadFile(filepath.Join(s.root, "cards", entry.Tag, id, "CARD.md"))
	if err != nil {
		return "", err
	}
	lines, end, err := frontmatter(string(content))
	if err != nil {
		return "", fmt.Errorf("card %s: %w", id, err)
	}
	question, _, err := cardSections(strings.Join(lines[end+1:], ""))
	if err != nil {
		return "", fmt.Errorf("card %s: %w", id, err)
	}
	return question, nil
}
