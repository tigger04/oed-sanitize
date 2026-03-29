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
