package huid

import (
	"testing"
	"time"
)

func TestParseAndValid(t *testing.T) {
	for _, value := range []string{
		"20260829-235855",
		"20260829-235855-rexim",
		"20260829-235855-01-A-b",
		"20260829-235855-",
	} {
		id, err := Parse(value)
		if err != nil || !Valid(value) || id.String() != value {
			t.Errorf("Parse(%q) = %q, %v; Valid = %v", value, id, err, Valid(value))
		}
	}

	for _, value := range []string{
		"", "2026082-235855", "20260829-23585", "20260829_235855",
		"20260829-235855!", "20260829-235855-café", "20260829-235855-x y",
	} {
		if _, err := Parse(value); err == nil || Valid(value) {
			t.Errorf("%q should be invalid", value)
		}
	}
}

func TestNewUsesUTCAndValidatesSuffix(t *testing.T) {
	local := time.Date(2026, time.August, 30, 1, 58, 55, 0, time.FixedZone("UTC+2", 2*60*60))
	for _, tc := range []struct {
		suffix string
		want   string
	}{
		{"", "20260829-235855"},
		{"rexim-01", "20260829-235855-rexim-01"},
	} {
		id, err := New(local, tc.suffix)
		if err != nil || id.String() != tc.want {
			t.Errorf("New(%q) = %q, %v; want %q", tc.suffix, id, err, tc.want)
		}
	}
	if _, err := New(local, "bad_suffix"); err == nil {
		t.Fatal("New accepted an invalid suffix")
	}
}

func TestChopConsumesValidPrefix(t *testing.T) {
	for _, tc := range []struct {
		input string
		want  string
		rest  string
		ok    bool
	}{
		{"20260829-235855-rexim!", "20260829-235855-rexim", "!", true},
		{"20260829-235855: note", "20260829-235855", ": note", true},
		{"20260829-235855-!", "20260829-235855-", "!", true},
		{"20260829-23585", "", "20260829-23585", false},
	} {
		id, rest, ok := Chop(tc.input)
		if id.String() != tc.want || rest != tc.rest || ok != tc.ok {
			t.Errorf("Chop(%q) = %q, %q, %v; want %q, %q, %v", tc.input, id, rest, ok, tc.want, tc.rest, tc.ok)
		}
	}
}
