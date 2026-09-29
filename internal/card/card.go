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
