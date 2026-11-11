package review

import "charm.land/lipgloss/v2"

// Colors decorate roles; labels and emphasis remain meaningful without color.
type palette struct {
	dark, noColor bool
}

func (p palette) colors() (accent, secondary, success string) {
	if p.dark {
		return "#BD93F9", "#A0A0A0", "#73D39C"
	}
	return "#6F42C1", "#626262", "#237A43"
}

func (p palette) accent() lipgloss.Style {
	s := lipgloss.NewStyle().Bold(true)
	accent, _, _ := p.colors()
	if !p.noColor {
		s = s.Foreground(lipgloss.Color(accent))
	}
	return s
}

func (p palette) secondary() lipgloss.Style {
	s := lipgloss.NewStyle()
	_, secondary, _ := p.colors()
	if !p.noColor {
		s = s.Foreground(lipgloss.Color(secondary))
	}
	return s
}

func (p palette) success() lipgloss.Style {
	s := lipgloss.NewStyle()
	_, _, success := p.colors()
	if !p.noColor {
		s = s.Foreground(lipgloss.Color(success))
	}
	return s
}
