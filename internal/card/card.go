package card

import (
	"github.com/emiliopalmerini/gostudy/internal/huid"
)

type Card struct {
	Answer   string
	ID       huid.HUID
	Question string
	Tags     []*Tag
}

type Tag struct {
	Value string
}
