// ABOUTME: Symbol sanitization engine. Replaces typographic characters (smart
// quotes, dashes, ellipsis, arrows, bullets) with ASCII equivalents.
package spelling

import (
	"strings"
	"unicode"
)

// bulletChars contains Unicode bullet characters commonly used by Word/Outlook.
var bulletChars = map[rune]bool{
	'\u2022': true, // •
	'\u25E6': true, // ◦
	'\u25AA': true, // ▪
	'\u25B8': true, // ▸
	'\u25BA': true, // ►
	'\u2023': true, // ‣
	'\u2043': true, // ⁃
}

// charReplacements maps typographic characters to their ASCII equivalents.
var charReplacements = []struct {
	old string
	new string
}{
	{"\u201c", `"`},  // left double quote
	{"\u201d", `"`},  // right double quote
	{"\u2018", "'"},  // left single quote
	{"\u2019", "'"},  // right single quote
	{"\u2014", "-"},  // em dash
	{"\u2013", "-"},  // en dash
	{"\u2026", "..."}, // ellipsis
	{"\u2192", "->"},  // arrow
}

// SymbolEngine tracks replacement counts for symbol sanitization.
type SymbolEngine struct {
	Changes int
}

// NewSymbolEngine creates a new symbol sanitization engine.
func NewSymbolEngine() *SymbolEngine {
	return &SymbolEngine{}
}

// ProcessLine applies symbol replacements to a single line.
func (e *SymbolEngine) ProcessLine(line string) string {
	// Handle bullet-prefixed lines first
	if result, handled := e.processBulletLine(line); handled {
		return result
	}

	// Apply character replacements
	for _, r := range charReplacements {
		count := strings.Count(line, r.old)
		if count > 0 {
			line = strings.ReplaceAll(line, r.old, r.new)
			e.Changes += count
		}
	}

	return line
}

// processBulletLine handles lines that start with optional whitespace + bullet.
// Returns the processed line and true if a bullet was found at line start.
func (e *SymbolEngine) processBulletLine(line string) (string, bool) {
	runes := []rune(line)

	var indent strings.Builder
	i := 0

	// Consume leading whitespace, converting tabs to 2 spaces
	for i < len(runes) && (runes[i] == ' ' || runes[i] == '\t') {
		if runes[i] == '\t' {
			indent.WriteString("  ")
		} else {
			indent.WriteRune(runes[i])
		}
		i++
	}

	// Check for bullet character
	if i >= len(runes) || !bulletChars[runes[i]] {
		return "", false
	}

	// Found a bullet at line start
	e.Changes++
	i++ // skip the bullet

	// Consume whitespace after bullet
	for i < len(runes) && (runes[i] == ' ' || runes[i] == '\t') {
		i++
	}

	// Build result: preserved indent + "- " + rest of line
	var result strings.Builder
	result.WriteString(indent.String())
	result.WriteString("- ")
	result.WriteString(string(runes[i:]))

	// Also apply character replacements to the rest of the line
	rest := result.String()
	for _, r := range charReplacements {
		count := strings.Count(rest, r.old)
		if count > 0 {
			rest = strings.ReplaceAll(rest, r.old, r.new)
			e.Changes += count
		}
	}

	return rest, true
}

// isBulletChar returns true if the rune is a recognised bullet character.
func isBulletChar(r rune) bool {
	return bulletChars[r]
}

// isWhitespace returns true for space and tab characters.
func isWhitespace(r rune) bool {
	return unicode.IsSpace(r)
}
