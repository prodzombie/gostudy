package huid

import (
	"fmt"
	"time"
)

// HUID is a human-readable identifier in tatr's YYYYMMDD-HHMMSS format,
// optionally followed by a dash and an ASCII alphanumeric or dash suffix.
// Its zero value is invalid.
type HUID struct {
	value string
}

// New creates a HUID from a UTC timestamp and an optional suffix.
// An empty suffix produces the short form without a trailing dash.
func New(now time.Time, suffix string) (HUID, error) {
	for i := 0; i < len(suffix); i++ {
		if !validSuffixByte(suffix[i]) {
			return HUID{}, fmt.Errorf("invalid HUID suffix %q", suffix)
		}
	}

	value := now.UTC().Format("20060102-150405")
	if suffix != "" {
		value += "-" + suffix
	}
	return Parse(value)
}

// Parse returns a HUID only when the entire string has tatr's HUID format.
func Parse(value string) (HUID, error) {
	if !Valid(value) {
		return HUID{}, fmt.Errorf("invalid HUID %q", value)
	}
	return HUID{value: value}, nil
}

// Valid reports whether value is a complete HUID.
func Valid(value string) bool {
	_, rest, ok := Chop(value)
	return ok && rest == ""
}

// Chop consumes a HUID at the beginning of input and returns the remainder.
// It leaves input untouched on failure. A suffix extends through consecutive
// ASCII letters, digits, and dashes, matching tatr's prefix parser.
func Chop(input string) (HUID, string, bool) {
	if len(input) < 15 || !validBase(input[:15]) {
		return HUID{}, input, false
	}

	end := 15
	if len(input) > end && input[end] == '-' {
		end++
		for end < len(input) && validSuffixByte(input[end]) {
			end++
		}
	}
	return HUID{value: input[:end]}, input[end:], true
}

// String returns the HUID in its original text form.
func (id HUID) String() string {
	return id.value
}

func validBase(value string) bool {
	if len(value) != 15 || value[8] != '-' {
		return false
	}
	for i := 0; i < len(value); i++ {
		if i != 8 && (value[i] < '0' || value[i] > '9') {
			return false
		}
	}
	return true
}

func validSuffixByte(value byte) bool {
	return value >= '0' && value <= '9' ||
		value >= 'A' && value <= 'Z' ||
		value >= 'a' && value <= 'z' ||
		value == '-'
}
