package card

import (
	"github.com/emiliopalmerini/gostudy/internal/huid"
)

// Card pairs a study question with its answer and tags.
type Card struct {
	Answer   string
	ID       huid.HUID
	Question string
	Tags     *Tag
}

// Tag names a category that can be attached to a card.
type Tag struct {
	Value string
}

// ValidTag reports whether a tag is a safe path component and contains a non-digit.
func ValidTag(tag string) bool {
	if tag == "" || tag == "." || tag == ".." {
		return false
	}
	hasNonDigit := false
	for _, c := range tag {
		if !('a' <= c && c <= 'z' || 'A' <= c && c <= 'Z' || '0' <= c && c <= '9' || c == '_' || c == '-') {
			return false
		}
		if c < '0' || c > '9' {
			hasNonDigit = true
		}
	}
	return hasNonDigit
}
