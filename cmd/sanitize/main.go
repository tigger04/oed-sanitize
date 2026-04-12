// ABOUTME: CLI entry point for the sanitize tool. Parses subcommands and flags,
// reads stdin line-by-line, applies transformations, writes to stdout.
package main

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/tigger04/oed-sanitize/data"
	"github.com/tigger04/oed-sanitize/pkg/spelling"
)

// version is set at build time via -ldflags.
var version = "dev"

const usageText = `usage: sanitize <subcommand> [<subcommand>...] [flags]

Subcommands:
  oed       Convert US→UK and -ise→-ize spellings
  symbols   Convert typographic characters to ASCII

Flags:
  -q          Suppress change summary on stderr
  -h, --help  Print this help message
  --version   Print version`

func main() {
	var doOED bool
	var doSymbols bool
	var quiet bool

	for _, arg := range os.Args[1:] {
		switch arg {
		case "oed":
			doOED = true
		case "symbols":
			doSymbols = true
		case "-q":
			quiet = true
		case "-h", "--help":
			fmt.Println(usageText)
			os.Exit(0)
		case "--version":
			fmt.Printf("sanitize %s\n", version)
			os.Exit(0)
		default:
			fmt.Fprintf(os.Stderr, "unknown argument: %s\n", arg)
			os.Exit(2)
		}
	}

	// Default to both when no subcommand specified
	defaulting := false
	if !doOED && !doSymbols {
		doOED = true
		doSymbols = true
		defaulting = true
	}

	if defaulting && !quiet {
		fmt.Fprintln(os.Stderr, "sanitize: defaulting to oed + symbols")
	}

	var oedEngine *spelling.OEDEngine
	var symEngine *spelling.SymbolEngine

	if doOED {
		var err error
		oedEngine, err = spelling.NewOEDEngine(data.UsToUkData, data.IseToIzeData)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
	}

	if doSymbols {
		symEngine = spelling.NewSymbolEngine()
	}

	scanner := bufio.NewScanner(os.Stdin)
	writer := bufio.NewWriter(os.Stdout)

	inCodeBlock := false
	for scanner.Scan() {
		line := scanner.Text()

		// Check for fenced code block delimiters (Markdown ``` and org-mode #+BEGIN_SRC/#+END_SRC)
		if isCodeBlockDelimiter(line, inCodeBlock) {
			inCodeBlock = !inCodeBlock
			fmt.Fprintln(writer, line)
			continue
		}

		// Lines inside fenced/org code blocks pass through unchanged
		if inCodeBlock {
			fmt.Fprintln(writer, line)
			continue
		}

		// Process inline code spans: split, process only Text segments, reassemble
		line = processWithCodeSpans(line, oedEngine, symEngine)
		fmt.Fprintln(writer, line)
	}

	writer.Flush()

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error reading stdin: %v\n", err)
		os.Exit(1)
	}

	if !quiet {
		if oedEngine != nil && oedEngine.SpellingChanges > 0 {
			fmt.Fprintln(os.Stderr, pluralize(oedEngine.SpellingChanges, "US spelling correction"))
		}
		if oedEngine != nil && oedEngine.IzeChanges > 0 {
			fmt.Fprintln(os.Stderr, pluralize(oedEngine.IzeChanges, "-ize correction"))
		}
		if symEngine != nil && symEngine.Changes > 0 {
			fmt.Fprintln(os.Stderr, pluralize(symEngine.Changes, "symbol replacement"))
		}
	}
}

// pluralize returns "N label" for count==1, "N labels" otherwise.
func pluralize(count int, label string) string {
	if count == 1 {
		return fmt.Sprintf("%d %s", count, label)
	}
	return fmt.Sprintf("%d %ss", count, label)
}

// isCodeBlockDelimiter returns true if the line opens or closes a fenced code
// block. Supports Markdown triple-backtick (with optional language identifier)
// and org-mode #+BEGIN_SRC / #+END_SRC (case-insensitive).
func isCodeBlockDelimiter(line string, inBlock bool) bool {
	trimmed := strings.TrimSpace(line)
	if strings.HasPrefix(trimmed, "```") {
		return true
	}
	upper := strings.ToUpper(trimmed)
	if !inBlock && (upper == "#+BEGIN_SRC" || strings.HasPrefix(upper, "#+BEGIN_SRC ")) {
		return true
	}
	if inBlock && upper == "#+END_SRC" {
		return true
	}
	return false
}

// processWithCodeSpans splits a line into code and text segments, applies
// engines only to text segments, and reassembles the line.
func processWithCodeSpans(line string, oedEngine *spelling.OEDEngine, symEngine *spelling.SymbolEngine) string {
	segments := spelling.SplitCodeSpans(line)

	// Fast path: if there are no code spans, process the whole line directly
	hasCode := false
	for _, seg := range segments {
		if seg.Kind == spelling.Code {
			hasCode = true
			break
		}
	}
	if !hasCode {
		if oedEngine != nil {
			line = oedEngine.ProcessLine(line)
		}
		if symEngine != nil {
			line = symEngine.ProcessLine(line)
		}
		return line
	}

	// Process only Text segments, leave Code segments unchanged
	var result strings.Builder
	for _, seg := range segments {
		if seg.Kind == spelling.Text {
			s := seg.Content
			if oedEngine != nil {
				s = oedEngine.ProcessLine(s)
			}
			if symEngine != nil {
				s = symEngine.ProcessLine(s)
			}
			result.WriteString(s)
		} else {
			result.WriteString(seg.Content)
		}
	}
	return result.String()
}
