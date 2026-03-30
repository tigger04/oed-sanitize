// ABOUTME: OED spelling engine. Loads US→UK and -ise→-ize word lists into
// separate maps, performs case-preserving whole-word replacement, and tracks
// spelling and -ize correction counts independently.
package spelling

import (
	"strings"
	"unicode"
)

// OEDEngine holds separate lookup maps for US→UK spelling and -ise→-ize
// corrections, tracking replacement counts independently.
type OEDEngine struct {
	spelling map[string]string
	ize      map[string]string
	SpellingChanges int
	IzeChanges      int
}

// NewOEDEngine creates an engine from two word list data strings:
// the first for US→UK spelling, the second for -ise→-ize corrections.
func NewOEDEngine(spellingData, izeData string) (*OEDEngine, error) {
	e := &OEDEngine{
		spelling: make(map[string]string),
		ize:      make(map[string]string),
	}
	if err := e.loadWordList(e.spelling, spellingData); err != nil {
		return nil, err
	}
	if err := e.loadWordList(e.ize, izeData); err != nil {
		return nil, err
	}
	return e, nil
}

// loadWordList parses lines of "wrong=correct" into the given map.
func (e *OEDEngine) loadWordList(dest map[string]string, data string) error {
	for _, line := range strings.Split(data, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if key != "" && val != "" {
			dest[strings.ToLower(key)] = strings.ToLower(val)
		}
	}
	return nil
}

// ProcessLine replaces words in a single line, preserving case.
func (e *OEDEngine) ProcessLine(line string) string {
	runes := []rune(line)
	var result strings.Builder
	result.Grow(len(line))

	i := 0
	for i < len(runes) {
		if isWordChar(runes[i]) {
			// Extract the whole word
			j := i
			for j < len(runes) && isWordChar(runes[j]) {
				j++
			}
			word := string(runes[i:j])
			replaced := e.replaceWord(word)
			result.WriteString(replaced)
			i = j
		} else {
			result.WriteRune(runes[i])
			i++
		}
	}
	return result.String()
}

// isWordChar returns true for letters and apostrophes (word-internal).
func isWordChar(r rune) bool {
	return unicode.IsLetter(r) || r == '\''
}

// replaceWord looks up a word in both maps and applies case-preserving
// replacement, incrementing the appropriate counter.
func (e *OEDEngine) replaceWord(word string) string {
	lower := strings.ToLower(word)
	if replacement, ok := e.spelling[lower]; ok {
		e.SpellingChanges++
		return applyCase(word, replacement)
	}
	if replacement, ok := e.ize[lower]; ok {
		e.IzeChanges++
		return applyCase(word, replacement)
	}
	return word
}

// applyCase transfers the case pattern of orig onto replacement.
func applyCase(orig, replacement string) string {
	if orig == strings.ToLower(orig) {
		return replacement
	}
	if orig == strings.ToUpper(orig) {
		return strings.ToUpper(replacement)
	}
	// Title Case: first letter uppercase, rest lowercase
	origRunes := []rune(orig)
	if len(origRunes) > 0 && unicode.IsUpper(origRunes[0]) {
		runes := []rune(replacement)
		if len(runes) > 0 {
			runes[0] = unicode.ToUpper(runes[0])
		}
		return string(runes)
	}
	// Mixed case fallback: return lowercase
	return replacement
}
