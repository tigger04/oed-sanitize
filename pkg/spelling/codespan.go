// ABOUTME: Inline code span splitter. Splits a line into segments tagged as
// Code (inside backticks) or Text (outside backticks) for selective processing.
package spelling

// SegmentKind distinguishes code spans from normal text.
type SegmentKind int

const (
	// Text is content outside backtick spans, eligible for processing.
	Text SegmentKind = iota
	// Code is content inside backtick spans, exempt from processing.
	Code
)

// Segment is a contiguous piece of a line with a kind tag.
type Segment struct {
	Content string
	Kind    SegmentKind
}

// SplitCodeSpans splits a line into Text and Code segments based on backtick
// delimiters. Backtick characters are included in the Code segments.
// Unclosed backticks are treated as literal text.
func SplitCodeSpans(line string) []Segment {
	runes := []rune(line)
	var segments []Segment
	i := 0

	for i < len(runes) {
		if runes[i] == '`' {
			// Find the closing backtick
			j := i + 1
			for j < len(runes) && runes[j] != '`' {
				j++
			}
			if j < len(runes) {
				// Found closing backtick — emit Code segment including delimiters
				segments = append(segments, Segment{
					Content: string(runes[i : j+1]),
					Kind:    Code,
				})
				i = j + 1
			} else {
				// Unclosed backtick — treat rest as Text
				segments = append(segments, Segment{
					Content: string(runes[i:]),
					Kind:    Text,
				})
				i = len(runes)
			}
		} else {
			// Collect text until next backtick or end of line
			j := i
			for j < len(runes) && runes[j] != '`' {
				j++
			}
			segments = append(segments, Segment{
				Content: string(runes[i:j]),
				Kind:    Text,
			})
			i = j
		}
	}

	return segments
}
