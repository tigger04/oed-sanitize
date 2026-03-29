// ABOUTME: CLI entry point for the sanitize tool. Parses subcommands and flags,
// reads stdin line-by-line, applies transformations, writes to stdout.
package main

import (
	"bufio"
	"fmt"
	"os"

	"github.com/tigger04/british-english-oed-fix/data"
	"github.com/tigger04/british-english-oed-fix/pkg/spelling"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: sanitize <subcommand> [flags]")
		fmt.Fprintln(os.Stderr, "subcommands: oed, symbols")
		os.Exit(2)
	}

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
		default:
			fmt.Fprintf(os.Stderr, "unknown argument: %s\n", arg)
			os.Exit(2)
		}
	}

	if !doOED && !doSymbols {
		fmt.Fprintln(os.Stderr, "usage: sanitize <subcommand> [flags]")
		fmt.Fprintln(os.Stderr, "subcommands: oed, symbols")
		os.Exit(2)
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

	for scanner.Scan() {
		line := scanner.Text()
		// Fixed order: spelling first, then symbols (per architecture.md)
		if oedEngine != nil {
			line = oedEngine.ProcessLine(line)
		}
		if symEngine != nil {
			line = symEngine.ProcessLine(line)
		}
		fmt.Fprintln(writer, line)
	}

	writer.Flush()

	if err := scanner.Err(); err != nil {
		fmt.Fprintf(os.Stderr, "error reading stdin: %v\n", err)
		os.Exit(1)
	}

	if !quiet {
		totalChanges := 0
		var parts []string
		if oedEngine != nil && oedEngine.Changes > 0 {
			parts = append(parts, fmt.Sprintf("%d spelling corrections", oedEngine.Changes))
			totalChanges += oedEngine.Changes
		}
		if symEngine != nil && symEngine.Changes > 0 {
			parts = append(parts, fmt.Sprintf("%d symbol replacements", symEngine.Changes))
			totalChanges += symEngine.Changes
		}
		if totalChanges > 0 {
			fmt.Fprintln(os.Stderr, joinParts(parts))
		}
	}
}

// joinParts joins summary parts with ", ".
func joinParts(parts []string) string {
	result := ""
	for i, p := range parts {
		if i > 0 {
			result += ", "
		}
		result += p
	}
	return result
}
