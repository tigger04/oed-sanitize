// ABOUTME: Integration tests for the sanitize CLI binary.
// Tests stdin/stdout pipeline behaviour, stderr summary, and embedded word lists.
package regression

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// binaryPath returns the path to the compiled sanitize binary.
// Tests must be run after `make build`.
func binaryPath(t *testing.T) string {
	t.Helper()
	// Look for binary relative to the repo root
	// When running via `go test`, the working dir is the test file's directory
	candidates := []string{
		filepath.Join("..", "..", "sanitize"),
		filepath.Join("..", "..", "bin", "sanitize"),
	}
	for _, p := range candidates {
		abs, err := filepath.Abs(p)
		if err != nil {
			continue
		}
		if _, err := os.Stat(abs); err == nil {
			return abs
		}
	}
	t.Fatal("sanitize binary not found — run `make build` first")
	return ""
}

// RT-011: pipe round-trip with corrections
func TestCLI_PipeRoundTrip_CorrectOutput_RT011(t *testing.T) {
	bin := binaryPath(t)

	cmd := exec.Command(bin, "oed")
	cmd.Stdin = strings.NewReader("organise the center")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	got := strings.TrimRight(string(out), "\n")
	want := "organize the centre"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// RT-012: multi-line input preserves line structure
func TestCLI_MultiLineInput_PreservesStructure_RT012(t *testing.T) {
	bin := binaryPath(t)

	input := "organise this\ncenter that\n"
	cmd := exec.Command(bin, "oed")
	cmd.Stdin = strings.NewReader(input)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	got := string(out)
	want := "organize this\ncentre that\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// RT-013: empty input yields empty output
func TestCLI_EmptyInput_EmptyOutput_RT013(t *testing.T) {
	bin := binaryPath(t)

	cmd := exec.Command(bin, "oed")
	cmd.Stdin = strings.NewReader("")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	if len(out) != 0 {
		t.Errorf("expected empty output, got %q", string(out))
	}
}

// RT-014: stderr contains change count with default flags
func TestCLI_DefaultFlags_StderrHasCount_RT014(t *testing.T) {
	bin := binaryPath(t)

	cmd := exec.Command(bin, "oed")
	cmd.Stdin = strings.NewReader("organise the center")

	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("command failed: %v", err)
	}

	errOut := stderr.String()
	if errOut == "" {
		t.Error("expected stderr output, got empty")
	}
	// "organise the center" triggers 1 US spelling + 1 -ize correction
	if !strings.Contains(errOut, "US spelling") {
		t.Errorf("expected stderr to mention US spelling corrections, got %q", errOut)
	}
	if !strings.Contains(errOut, "-ize") {
		t.Errorf("expected stderr to mention -ize corrections, got %q", errOut)
	}
}

// RT-015: stderr is empty with -q flag
func TestCLI_QuietFlag_StderrEmpty_RT015(t *testing.T) {
	bin := binaryPath(t)

	cmd := exec.Command(bin, "oed", "-q")
	cmd.Stdin = strings.NewReader("organise the center")

	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("command failed: %v", err)
	}

	if stderr.String() != "" {
		t.Errorf("expected empty stderr with -q, got %q", stderr.String())
	}
}

// RT-016: binary runs correctly from a directory containing no .txt files
func TestCLI_NoTxtFilesInDir_StillWorks_RT016(t *testing.T) {
	bin := binaryPath(t)

	tmpDir := t.TempDir()

	cmd := exec.Command(bin, "oed")
	cmd.Dir = tmpDir
	cmd.Stdin = strings.NewReader("organise")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("command failed from tmpdir: %v", err)
	}

	got := strings.TrimRight(string(out), "\n")
	if got != "organize" {
		t.Errorf("got %q, want %q", got, "organize")
	}
}

// RT-028: sanitize symbols pipe round-trip with change count on stderr
func TestCLI_SymbolsPipeRoundTrip_CorrectOutputAndStderr_RT028(t *testing.T) {
	bin := binaryPath(t)

	cmd := exec.Command(bin, "symbols")
	cmd.Stdin = strings.NewReader("He said \u201chello\u201d\u2026")

	var stderr strings.Builder
	cmd.Stderr = &stderr

	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	got := strings.TrimRight(string(out), "\n")
	want := `He said "hello"...`
	if got != want {
		t.Errorf("stdout: got %q, want %q", got, want)
	}

	errOut := stderr.String()
	if errOut == "" {
		t.Error("expected stderr output, got empty")
	}
}

// RT-029: sanitize symbols -q suppresses stderr
func TestCLI_SymbolsQuietFlag_StderrEmpty_RT029(t *testing.T) {
	bin := binaryPath(t)

	cmd := exec.Command(bin, "symbols", "-q")
	cmd.Stdin = strings.NewReader("He said \u201chello\u201d")

	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("command failed: %v", err)
	}

	if stderr.String() != "" {
		t.Errorf("expected empty stderr with -q, got %q", stderr.String())
	}
}

// RT-030: no subcommand defaults to both (exits 0, not error)
// Updated for issue #5: no subcommand now defaults to oed + symbols
func TestCLI_NoSubcommand_ExitsZero_RT030(t *testing.T) {
	bin := binaryPath(t)

	cmd := exec.Command(bin, "-q")
	cmd.Stdin = strings.NewReader("")

	err := cmd.Run()
	if err != nil {
		t.Fatalf("expected exit 0, got error: %v", err)
	}
}

// RT-031: oed symbols and symbols oed produce identical output
func TestCLI_SubcommandOrder_Irrelevant_RT031(t *testing.T) {
	bin := binaryPath(t)

	input := "I need to organise the center\u2014it\u2019s important\u2026"

	cmd1 := exec.Command(bin, "oed", "symbols", "-q")
	cmd1.Stdin = strings.NewReader(input)
	out1, err := cmd1.Output()
	if err != nil {
		t.Fatalf("oed symbols failed: %v", err)
	}

	cmd2 := exec.Command(bin, "symbols", "oed", "-q")
	cmd2.Stdin = strings.NewReader(input)
	out2, err := cmd2.Output()
	if err != nil {
		t.Fatalf("symbols oed failed: %v", err)
	}

	if string(out1) != string(out2) {
		t.Errorf("outputs differ:\n  oed symbols: %q\n  symbols oed: %q", string(out1), string(out2))
	}
}

// RT-032: duplicate subcommand deduplication
func TestCLI_DuplicateSubcommand_Deduplicated_RT032(t *testing.T) {
	bin := binaryPath(t)

	cmd := exec.Command(bin, "oed", "oed", "-q")
	cmd.Stdin = strings.NewReader("organise")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	got := strings.TrimRight(string(out), "\n")
	if got != "organize" {
		t.Errorf("got %q, want %q", got, "organize")
	}
}

// RT-033: unknown argument exits non-zero
func TestCLI_UnknownArg_ExitsNonZero_RT033(t *testing.T) {
	bin := binaryPath(t)

	cmd := exec.Command(bin, "foo")
	cmd.Stdin = strings.NewReader("")

	var stderr strings.Builder
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatal("expected non-zero exit, got success")
	}

	if stderr.String() == "" {
		t.Error("expected error message on stderr, got empty")
	}
}

// RT-034: -h exits 0 with usage
func TestCLI_HelpFlag_ExitsZeroWithUsage_RT034(t *testing.T) {
	bin := binaryPath(t)

	cmd := exec.Command(bin, "-h")

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("expected exit 0, got error: %v", err)
	}

	combined := stdout.String() + stderr.String()
	if !strings.Contains(combined, "usage") && !strings.Contains(combined, "Usage") {
		t.Errorf("expected usage text, got stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

// RT-035: --version exits 0 with version string
func TestCLI_VersionFlag_ExitsZeroWithVersion_RT035(t *testing.T) {
	bin := binaryPath(t)

	cmd := exec.Command(bin, "--version")

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		t.Fatalf("expected exit 0, got error: %v", err)
	}

	combined := stdout.String() + stderr.String()
	if !strings.Contains(combined, "sanitize") {
		t.Errorf("expected version string containing 'sanitize', got stdout=%q stderr=%q", stdout.String(), stderr.String())
	}
}

// RT-036: no subcommand defaults to both oed + symbols
func TestCLI_NoSubcommand_DefaultsBoth_RT036(t *testing.T) {
	bin := binaryPath(t)

	cmd := exec.Command(bin, "-q")
	cmd.Stdin = strings.NewReader("organise the center\u2026")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	got := strings.TrimRight(string(out), "\n")
	want := "organize the centre..."
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// RT-037: defaulting notice on stderr when no subcommand given
func TestCLI_NoSubcommand_StderrHasDefaultingNotice_RT037(t *testing.T) {
	bin := binaryPath(t)

	cmd := exec.Command(bin)
	cmd.Stdin = strings.NewReader("organise")

	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("command failed: %v", err)
	}

	lines := strings.Split(stderr.String(), "\n")
	if len(lines) == 0 || !strings.Contains(lines[0], "defaulting") {
		t.Errorf("expected first stderr line to contain 'defaulting', got %q", stderr.String())
	}
}

// RT-038: no defaulting notice when subcommands are explicit
func TestCLI_ExplicitSubcommand_NoDefaultingNotice_RT038(t *testing.T) {
	bin := binaryPath(t)

	cmd := exec.Command(bin, "oed")
	cmd.Stdin = strings.NewReader("organise")

	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("command failed: %v", err)
	}

	if strings.Contains(stderr.String(), "defaulting") {
		t.Errorf("expected no defaulting notice with explicit subcommand, got %q", stderr.String())
	}
}

// RT-039: -q suppresses defaulting notice and summary when no subcommand
func TestCLI_NoSubcommandQuiet_StderrEmpty_RT039(t *testing.T) {
	bin := binaryPath(t)

	cmd := exec.Command(bin, "-q")
	cmd.Stdin = strings.NewReader("organise the center\u2026")

	var stderr strings.Builder
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("command failed: %v", err)
	}

	if stderr.String() != "" {
		t.Errorf("expected empty stderr with -q, got %q", stderr.String())
	}
}

// --- Issue #8: summary format tests ---

// summaryLines extracts non-empty summary lines from stderr, skipping the
// "defaulting to oed + symbols" notice if present.
func summaryLines(stderr string) []string {
	var lines []string
	for _, l := range strings.Split(stderr, "\n") {
		l = strings.TrimSpace(l)
		if l == "" || strings.Contains(l, "defaulting") {
			continue
		}
		lines = append(lines, l)
	}
	return lines
}

// RT-043: single US spelling change uses singular label
func TestCLI_SingleUSSpelling_SingularLabel_RT043(t *testing.T) {
	bin := binaryPath(t)
	cmd := exec.Command(bin, "oed")
	cmd.Stdin = strings.NewReader("center")
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("command failed: %v", err)
	}
	lines := summaryLines(stderr.String())
	if len(lines) != 1 {
		t.Fatalf("expected 1 summary line, got %d: %q", len(lines), stderr.String())
	}
	if lines[0] != "1 US spelling correction" {
		t.Errorf("got %q, want %q", lines[0], "1 US spelling correction")
	}
}

// RT-044: multiple US spelling changes uses plural label
func TestCLI_MultipleUSSpelling_PluralLabel_RT044(t *testing.T) {
	bin := binaryPath(t)
	cmd := exec.Command(bin, "oed")
	cmd.Stdin = strings.NewReader("center color")
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("command failed: %v", err)
	}
	lines := summaryLines(stderr.String())
	if len(lines) != 1 {
		t.Fatalf("expected 1 summary line, got %d: %q", len(lines), stderr.String())
	}
	if lines[0] != "2 US spelling corrections" {
		t.Errorf("got %q, want %q", lines[0], "2 US spelling corrections")
	}
}

// RT-045: single symbol change uses singular label
func TestCLI_SingleSymbol_SingularLabel_RT045(t *testing.T) {
	bin := binaryPath(t)
	cmd := exec.Command(bin, "symbols")
	cmd.Stdin = strings.NewReader("\u2014")
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("command failed: %v", err)
	}
	lines := summaryLines(stderr.String())
	if len(lines) != 1 {
		t.Fatalf("expected 1 summary line, got %d: %q", len(lines), stderr.String())
	}
	if lines[0] != "1 symbol replacement" {
		t.Errorf("got %q, want %q", lines[0], "1 symbol replacement")
	}
}

// RT-046: single -ize change uses singular label
func TestCLI_SingleIze_SingularLabel_RT046(t *testing.T) {
	bin := binaryPath(t)
	cmd := exec.Command(bin, "oed")
	cmd.Stdin = strings.NewReader("organise")
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("command failed: %v", err)
	}
	lines := summaryLines(stderr.String())
	if len(lines) != 1 {
		t.Fatalf("expected 1 summary line, got %d: %q", len(lines), stderr.String())
	}
	if lines[0] != "1 -ize correction" {
		t.Errorf("got %q, want %q", lines[0], "1 -ize correction")
	}
}

// RT-047: US spelling only — no -ize line on stderr
func TestCLI_USSpellingOnly_NoIzeLine_RT047(t *testing.T) {
	bin := binaryPath(t)
	cmd := exec.Command(bin, "oed")
	cmd.Stdin = strings.NewReader("center")
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("command failed: %v", err)
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "US spelling") {
		t.Errorf("expected US spelling line, got %q", errOut)
	}
	if strings.Contains(errOut, "-ize") {
		t.Errorf("expected no -ize line, got %q", errOut)
	}
}

// RT-048: -ize only — no US spelling line on stderr
func TestCLI_IzeOnly_NoUSSpellingLine_RT048(t *testing.T) {
	bin := binaryPath(t)
	cmd := exec.Command(bin, "oed")
	cmd.Stdin = strings.NewReader("organise")
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("command failed: %v", err)
	}
	errOut := stderr.String()
	if !strings.Contains(errOut, "-ize") {
		t.Errorf("expected -ize line, got %q", errOut)
	}
	if strings.Contains(errOut, "US spelling") {
		t.Errorf("expected no US spelling line, got %q", errOut)
	}
}

// RT-049: both US and -ize triggered — two separate lines
func TestCLI_BothUSAndIze_TwoLines_RT049(t *testing.T) {
	bin := binaryPath(t)
	cmd := exec.Command(bin, "oed")
	cmd.Stdin = strings.NewReader("center organise")
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("command failed: %v", err)
	}
	lines := summaryLines(stderr.String())
	if len(lines) != 2 {
		t.Fatalf("expected 2 summary lines, got %d: %q", len(lines), stderr.String())
	}
	hasUS := false
	hasIze := false
	for _, l := range lines {
		if strings.Contains(l, "US spelling") {
			hasUS = true
		}
		if strings.Contains(l, "-ize") {
			hasIze = true
		}
	}
	if !hasUS {
		t.Errorf("missing US spelling line in %q", stderr.String())
	}
	if !hasIze {
		t.Errorf("missing -ize line in %q", stderr.String())
	}
}

// RT-050: all three categories — three separate lines
func TestCLI_AllThreeCategories_ThreeLines_RT050(t *testing.T) {
	bin := binaryPath(t)
	cmd := exec.Command(bin)
	cmd.Stdin = strings.NewReader("center organise \u2014")
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("command failed: %v", err)
	}
	lines := summaryLines(stderr.String())
	if len(lines) != 3 {
		t.Fatalf("expected 3 summary lines, got %d: %q", len(lines), stderr.String())
	}
}

// RT-051: one category only — exactly one summary line
func TestCLI_OneCategory_OneLine_RT051(t *testing.T) {
	bin := binaryPath(t)
	cmd := exec.Command(bin, "symbols")
	cmd.Stdin = strings.NewReader("\u2014")
	var stderr strings.Builder
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		t.Fatalf("command failed: %v", err)
	}
	lines := summaryLines(stderr.String())
	if len(lines) != 1 {
		t.Fatalf("expected 1 summary line, got %d: %q", len(lines), stderr.String())
	}
}

// --- Issue #10: code block protection tests ---

// runSanitize is a helper that pipes input through the sanitize binary with -q
// and returns stdout. Uses both oed + symbols subcommands.
func runSanitize(t *testing.T, input string) string {
	t.Helper()
	bin := binaryPath(t)
	cmd := exec.Command(bin, "-q")
	cmd.Stdin = strings.NewReader(input)
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}
	return string(out)
}

// RT-10.1: OED-correctable word inside fenced block passes through unchanged
// User action: pipe a Markdown doc with a fenced code block containing "center"
// User observes: "center" is unchanged in the output
func TestCLI_FencedBlock_OEDSkipped_RT10_1(t *testing.T) {
	input := "```\ncenter\n```\n"
	got := runSanitize(t, input)
	want := "```\ncenter\n```\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// RT-10.2: Typographic symbol inside fenced block passes through unchanged
// User action: pipe a Markdown doc with an em dash inside a fenced code block
// User observes: the em dash is preserved, not converted to ASCII hyphen
func TestCLI_FencedBlock_SymbolsSkipped_RT10_2(t *testing.T) {
	input := "```\nvalue \u2014 other\n```\n"
	got := runSanitize(t, input)
	want := "```\nvalue \u2014 other\n```\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// RT-10.3: Fenced block with language specifier still protects content
// User action: pipe a Markdown doc with ```makefile block containing "center"
// User observes: "center" inside the block is unchanged
func TestCLI_FencedBlockLangSpec_Protected_RT10_3(t *testing.T) {
	input := "```makefile\n.PHONY: center organize\n```\n"
	got := runSanitize(t, input)
	want := "```makefile\n.PHONY: center organize\n```\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// RT-10.4: Multi-line fenced block protects all enclosed lines
// User action: pipe a Markdown doc with several lines inside a fenced block
// User observes: every line inside the block is unchanged
func TestCLI_FencedBlockMultiLine_AllProtected_RT10_4(t *testing.T) {
	input := "```\ncenter\norganise\ncolor \u2014 flavor\n```\n"
	got := runSanitize(t, input)
	want := "```\ncenter\norganise\ncolor \u2014 flavor\n```\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// RT-10.5: OED-correctable word inside inline backticks passes through unchanged
// User action: pipe a line with `center` in backticks
// User observes: "center" inside backticks is unchanged
func TestCLI_InlineCode_OEDSkipped_RT10_5(t *testing.T) {
	input := "the `center` variable\n"
	got := runSanitize(t, input)
	want := "the `center` variable\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// RT-10.6: Typographic symbol inside inline backticks passes through unchanged
// User action: pipe a line with an em dash inside backticks
// User observes: the em dash inside backticks is preserved
func TestCLI_InlineCode_SymbolsSkipped_RT10_6(t *testing.T) {
	input := "use `a \u2014 b` for ranges\n"
	got := runSanitize(t, input)
	want := "use `a \u2014 b` for ranges\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// RT-10.7: Multiple inline code spans on one line are each protected
// User action: pipe a line with two separate backtick spans
// User observes: both spans are unchanged
func TestCLI_InlineCodeMultiSpan_AllProtected_RT10_7(t *testing.T) {
	input := "use `center` and `color` here\n"
	got := runSanitize(t, input)
	want := "use `center` and `color` here\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// RT-10.8: Text outside inline code span is corrected, text inside is not
// User action: pipe a line with both inline code and normal text containing correctable words
// User observes: "center" outside backticks becomes "centre", "color" inside backticks stays "color"
func TestCLI_InlineCodeMixed_OutsideCorrected_RT10_8(t *testing.T) {
	input := "the center of `color` is important\n"
	got := runSanitize(t, input)
	want := "the centre of `color` is important\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// RT-10.9: Lines before and after a fenced code block are corrected normally
// User action: pipe a doc with text, then fenced block, then more text
// User observes: surrounding text is corrected, block content is not
func TestCLI_FencedBlockAdjacent_SurroundingCorrected_RT10_9(t *testing.T) {
	input := "center\n```\ncenter\n```\ncenter\n"
	got := runSanitize(t, input)
	want := "centre\n```\ncenter\n```\ncentre\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// RT-10.10: OED-correctable word inside org-mode source block passes through unchanged
// User action: pipe an org-mode file with #+BEGIN_SRC block containing "center"
// User observes: "center" inside the block is unchanged
func TestCLI_OrgSrcBlock_OEDSkipped_RT10_10(t *testing.T) {
	input := "#+BEGIN_SRC\ncenter\n#+END_SRC\n"
	got := runSanitize(t, input)
	want := "#+BEGIN_SRC\ncenter\n#+END_SRC\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// RT-10.11: Org-mode source block with language specifier still protects content
// User action: pipe an org file with #+BEGIN_SRC bash block
// User observes: content inside is unchanged
func TestCLI_OrgSrcBlockLangSpec_Protected_RT10_11(t *testing.T) {
	input := "#+BEGIN_SRC bash\norganise center\n#+END_SRC\n"
	got := runSanitize(t, input)
	want := "#+BEGIN_SRC bash\norganise center\n#+END_SRC\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

// RT-10.12: Case-insensitive matching — mixed-case delimiters are recognised
// User action: pipe an org file with lowercase #+begin_src delimiters
// User observes: content inside is unchanged
func TestCLI_OrgSrcBlockCaseInsensitive_RT10_12(t *testing.T) {
	input := "#+begin_src python\ncolor = \"center\"\n#+end_src\n"
	got := runSanitize(t, input)
	want := "#+begin_src python\ncolor = \"center\"\n#+end_src\n"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
