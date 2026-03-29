// ABOUTME: Unit tests for the symbol sanitization engine.
// Tests character replacements, bullet handling, and ASCII passthrough.
package spelling

import (
	"testing"
)

// RT-019: smart quotes (left/right double and single) → straight equivalents
func TestSymbols_SmartQuotes_ConvertToStraight_RT019(t *testing.T) {
	engine := NewSymbolEngine()

	// Double smart quotes
	got := engine.ProcessLine("He said \u201chello\u201d")
	want := `He said "hello"`
	if got != want {
		t.Errorf("double quotes: got %q, want %q", got, want)
	}

	// Reset for single quote test
	engine = NewSymbolEngine()

	// Single smart quotes and apostrophe
	got = engine.ProcessLine("It\u2019s a \u2018test\u2019")
	want = "It's a 'test'"
	if got != want {
		t.Errorf("single quotes: got %q, want %q", got, want)
	}
}

// RT-020: em dash and en dash → hyphen
func TestSymbols_Dashes_ConvertToHyphen_RT020(t *testing.T) {
	engine := NewSymbolEngine()

	got := engine.ProcessLine("word\u2014word and 1\u20132")
	want := "word-word and 1-2"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// RT-021: ellipsis character → three dots
func TestSymbols_Ellipsis_ConvertToThreeDots_RT021(t *testing.T) {
	engine := NewSymbolEngine()

	got := engine.ProcessLine("wait\u2026")
	want := "wait..."
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// RT-022: arrow character → "->"
func TestSymbols_Arrow_ConvertToASCII_RT022(t *testing.T) {
	engine := NewSymbolEngine()

	got := engine.ProcessLine("go \u2192 there")
	want := "go -> there"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// RT-023: top-level bullet → "- text"
func TestSymbols_TopLevelBullet_ConvertToHyphen_RT023(t *testing.T) {
	engine := NewSymbolEngine()

	got := engine.ProcessLine("\u2022\ttext")
	want := "- text"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// RT-024: tab-indented bullet → two-space-indented "- text"
func TestSymbols_TabIndentedBullet_ConvertToSpaceIndent_RT024(t *testing.T) {
	engine := NewSymbolEngine()

	got := engine.ProcessLine("\t\u2022\ttext")
	want := "  - text"
	if got != want {
		t.Errorf("single tab: got %q, want %q", got, want)
	}

	engine = NewSymbolEngine()

	got = engine.ProcessLine("\t\t\u2022 text")
	want = "    - text"
	if got != want {
		t.Errorf("double tab: got %q, want %q", got, want)
	}
}

// RT-025: space-indented bullet → spaces preserved + "- text"
func TestSymbols_SpaceIndentedBullet_PreservesSpaces_RT025(t *testing.T) {
	engine := NewSymbolEngine()

	got := engine.ProcessLine("  \u2022\ttext")
	want := "  - text"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// RT-026: mid-line bullet character is unchanged
func TestSymbols_MidLineBullet_Unchanged_RT026(t *testing.T) {
	engine := NewSymbolEngine()

	input := "some \u2022 text"
	got := engine.ProcessLine(input)
	if got != input {
		t.Errorf("got %q, want %q", got, input)
	}
}

// RT-027: plain ASCII text with standard quotes, hyphens, and dots is unchanged
func TestSymbols_PlainASCII_PassesThrough_RT027(t *testing.T) {
	engine := NewSymbolEngine()

	input := `He said "hello" - it's fine...`
	got := engine.ProcessLine(input)
	if got != input {
		t.Errorf("got %q, want %q", got, input)
	}
}
