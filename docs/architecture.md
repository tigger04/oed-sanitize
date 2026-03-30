<!-- Version: 1.0 | Last updated: 2026-03-29 -->

# Architecture

## Overview

`sanitize` is a Go CLI tool that reads text from stdin, applies transformations, and writes to stdout. It uses subcommands to select which transformations to apply.

## Language choice

Go was chosen for:
- Near-instant startup (~5ms) — critical for interactive use (clipboard paste)
- Single static binary — no runtime dependencies, trivial installation
- `embed.FS` — word lists compiled into the binary, no file resolution at runtime
- Straightforward string handling via stdlib (`strings`, `bufio`, `regexp`)
- Cross-compilation for Homebrew distribution

## CLI design

```
sanitize <subcommand> [<subcommand>...] [flags]

Subcommands:
  oed       Convert US→UK and -ise→-ize spellings
  symbols   Convert typographic characters to ASCII

Flags:
  -q        Suppress change summary on stderr
  -h        Print usage
  --version Print version

No subcommand → defaults to both oed + symbols (notice on stderr, suppressed by -q).
Subcommands can appear in any order.
Multiple subcommands apply all transformations in a fixed internal order.
```

### Transformation order

When both subcommands are active, transformations apply in this order:
1. **Spelling** (`oed`) — first, so that word boundaries are stable
2. **Symbols** (`symbols`) — after, since symbol replacement doesn't affect word content

This order is fixed regardless of subcommand order on the command line.

## Data flow

```
stdin → bufio.Scanner (line-by-line)
     → [if oed] spelling replacements (map lookup, case-preserving)
     → [if symbols] character replacements
     → stdout

stderr ← change summary (unless -q)
```

### Line-by-line processing

Text is processed line by line via `bufio.Scanner`. This keeps memory usage constant regardless of input size and preserves line structure exactly.

## Spelling engine (`oed` subcommand)

### Word lists

Two embedded word lists in `data/`:

| File | Entries | Purpose |
|------|---------|---------|
| `us-to-uk.txt` | 721 | US → UK spelling (center→centre, analyze→analyse) |
| `ise-to-ize.txt` | 1,213 | Non-OED British -ise → OED -ize (organise→organize) |

Both use `wrong=correct` format, one pair per line. Comments (`#`) and blank lines are ignored.

### Lookup strategy

At startup, both word lists are parsed into a single `map[string]string` (lowercase key → lowercase value). At ~2,000 entries this is trivially fast and uses negligible memory.

For each word in the input text:
1. Extract the word (contiguous letters/apostrophes)
2. Lowercase it
3. Look up in the map
4. If found, replace — preserving the original case pattern (all-lower, all-upper, title-case)
5. If not found, pass through unchanged

### Case preservation

Three patterns are recognized:

| Input pattern | Example | Replacement |
|---------------|---------|-------------|
| all lowercase | `center` | `centre` |
| ALL UPPERCASE | `CENTER` | `CENTRE` |
| Title Case | `Center` | `Centre` |
| Mixed/other | `cEnTeR` | `centre` (falls back to lowercase) |

### Word boundaries

Words are identified by splitting on non-letter characters. This avoids the brittleness of regex word boundaries and handles punctuation-adjacent words correctly (e.g. `"center"` → `"centre"`).

## Symbol engine (`symbols` subcommand)

Simple character/string replacements:

| Input | Output | Description |
|-------|--------|-------------|
| `\u201c` `\u201d` | `"` | Smart double quotes → straight |
| `\u2018` `\u2019` | `'` | Smart single quotes → straight |
| `\u2014` | `-` | Em dash → hyphen |
| `\u2013` | `-` | En dash → hyphen |
| `\u2026` | `...` | Ellipsis → three dots |
| `\u2192` | `->` | Arrow → ASCII arrow |
| `\u2022` (at line start) | `- ` | Bullet → hyphen list item |

## Embedding

Word lists are embedded at compile time via Go's `embed.FS`:

```go
//go:embed data/us-to-uk.txt
var usToUkData string

//go:embed data/ise-to-ize.txt
var iseToIzeData string
```

This means:
- No file I/O at runtime
- No path resolution issues
- The binary is fully self-contained
- Word lists are still human-editable .txt files in the repo

## Change summary

By default, `sanitize` writes a summary to stderr:

```
3 spelling corrections, 2 symbol replacements
```

Suppressed with `-q` for pipeline/scripting use.

## Project layout

```
oed-sanitize/
├── cmd/sanitize/main.go        # CLI entry point, flag parsing, subcommand dispatch
├── pkg/
│   └── spelling/
│       ├── oed.go              # Word list loading, map building, case-preserving replacement
│       ├── symbols.go          # Typographic character replacement
│       └── engine.go           # Orchestration: applies selected transformations per line
├── data/
│   ├── us-to-uk.txt            # US → UK word pairs
│   └── ise-to-ize.txt          # -ise → -ize word pairs
├── tests/
│   ├── regression/             # Automated regression tests
│   └── one_off/                # One-off tests (with .gitkeep)
├── docs/                       # Project documentation
├── Makefile                    # build, test, install, release
├── CLAUDE.md                   # Claude Code project instructions
└── README.md
```
