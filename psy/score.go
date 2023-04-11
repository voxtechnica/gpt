package psy

import (
	"errors"
	"strconv"
	"strings"
)

// Selection is the type of score selection.
type Selection string

const (
	First Selection = "first"
	Last  Selection = "last"
	All   Selection = "all"
	None  Selection = "none"
)

// String returns the string representation of the Selection.
func (s Selection) String() string {
	return string(s)
}

// IsValid returns true if the Selection is valid.
func (s Selection) IsValid() bool {
	return s == First || s == Last || s == All || s == None
}

// SelectScore selects the desired score(s) from a block of text.
// Valid selections are "first", "last", "all", or "none".
// An invalid selection defaults to "none".
func SelectScores(s string, sel Selection) []float32 {
	if s == "" || sel == None || !sel.IsValid() {
		return nil
	}
	fields := strings.Fields(s)
	if len(fields) == 0 {
		return nil
	}
	if sel == Last {
		// reverse the fields:
		for i, j := 0, len(fields)-1; i < j; i, j = i+1, j-1 {
			fields[i], fields[j] = fields[j], fields[i]
		}
	}
	if sel == First || sel == Last {
		// Find the first score:
		for _, field := range fields {
			if score, err := ParseScore(field); err == nil {
				return []float32{score}
			}
		}
		return nil
	}
	// Find all scores:
	var scores []float32
	for _, field := range fields {
		if score, err := ParseScore(field); err == nil {
			scores = append(scores, score)
		}
	}
	return scores
}

// ParseScore parses a string as a floating-point number.
func ParseScore(s string) (float32, error) {
	// Check if the word starts with a plus/minus sign or numeric digit:
	if len(s) == 0 || (s[0] != '-' && s[0] != '+' && (s[0] < '0' || s[0] > '9')) {
		return 0, errors.New("score: not a number")
	}
	// Remove trailing punctuation:
	for len(s) > 0 && (s[len(s)-1] < '0' || s[len(s)-1] > '9') {
		s = s[:len(s)-1]
	}
	// Parse the number:
	score, err := strconv.ParseFloat(s, 32)
	return float32(score), err
}
