# WTF: What's This For?

- A fast CLI tool that converts English text to [Oxford spelling](https://en.wikipedia.org/wiki/Oxford_spelling) (OED preferred British English with `-ize` suffixes).
- Sanitizes typographic symbols that cause problems in code and plain-text workflows.

```bash
# Fix spellings: US → UK (OED), and non-OED British -ise → -ize
echo "I need to organise the center" | sanitize oed
# Output: I need to organize the centre

# Fix typographic symbols: smart quotes, em dashes, ellipsis, etc.
echo 'He said "hello"…' | sanitize symbols
# Output: He said "hello"...

# Both at once (subcommands in any order)
echo 'I need to "organise" the center…' | sanitize oed symbols
# Output: I need to "organize" the centre...
```

## This is not a spell checker!

- It will not fix typos, misspellings, or grammar
- It will not highlight errors or suggest corrections
- It takes an English language string on stdin, and outputs the same string with many American or non-OED British spelling converted to OED spelling. That is all
- My dictionary is not exhaustive. I welcome contributions to improve it

## Subcommands

| Subcommand | What it does |
|-----------|--------------|
| `oed` | Converts US spellings to UK (center→centre) and non-OED -ise to -ize (organise→organize) |
| `symbols` | Converts typographic characters to ASCII equivalents (smart quotes, em/en dashes, ellipsis, bullets, arrows) |

Subcommands can be combined in any order. Running `sanitize` with no subcommands defaults to applying both `oed` and `symbols`.

## Code block protection

Content inside code blocks is never modified. This means technical documents, READMEs, and literate programs can be safely piped through `sanitize` without breaking code examples.

Supported code block formats:
- **Markdown fenced blocks** — `` ``` `` or `` ```language ``
- **Inline backtick spans** — `` `code here` ``
- **Org-mode source blocks** — `#+BEGIN_SRC` / `#+END_SRC` (case-insensitive)

## Flags

| Flag | Effect |
|------|--------|
| `-q` | Quiet mode — suppresses the change summary on stderr |
| `-h`, `--help` | Print usage |
| `--version` | Print version |

## Installation

```bash
brew install tigger04/tap/sanitize
```

Or build from source:

```bash
git clone https://github.com/tigger04/oed-sanitize.git
cd oed-sanitize
make install
```

## WTF: Why the ... ?

A question I get asked often:
> Why do you mix British and American spelling?

I don't. I follow Oxford spelling (`en-GB-oxendict`), which favours `-ize` endings. This is the OED standard, still preferred in technical writing, scientific papers, and by international organizations like the UN.

[Oxford spelling](https://en.wikipedia.org/wiki/Oxford_spelling) was the norm in most British newspapers until the early 2000s, when they collectively reverted to Cambridge spelling (`-ise`). The shift had more to do with asserting cultural distance from American English than with any linguistic rationale.

I speak Hiberno-English. I follow the OED because it offers the least ambiguity to the widest English-reading audience, short of switching to US English entirely - which isn't my dialect. I feel no obligation to follow the fashion of a neighbouring country that has abandoned its own dictionary's recommendation.

## Project structure

```
.
├── cmd/sanitize/main.go      # CLI entry point
├── pkg/spelling/              # Spelling conversion logic
├── data/
│   ├── us-to-uk.txt           # US → UK word pairs (721 entries)
│   └── ise-to-ize.txt         # -ise → -ize word pairs (1,213 entries)
├── docs/
│   ├── vision.md              # Project vision and goals
│   ├── architecture.md        # Technical architecture
│   ├── implementation-plan.md # Phased build plan
│   └── testing.md             # Testing strategy
├── tests/
│   └── regression/            # Automated tests
├── Makefile                   # Build, test, install, release
└── README.md
```

## Contributing

Word list contributions are welcome. The dictionaries live in `data/`:
- `us-to-uk.txt` — US to UK spelling pairs
- `ise-to-ize.txt` — non-OED -ise to OED -ize pairs

Format: `wrong_spelling=correct_spelling`, one per line. Lines starting with `#` are comments.

## Licence

MIT. Copyright Tadg Paul.
