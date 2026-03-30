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
	if !strings.Contains(errOut, "2") {
		t.Errorf("expected stderr to mention count of 2 corrections, got %q", errOut)
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
