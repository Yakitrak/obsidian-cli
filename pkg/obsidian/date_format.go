package obsidian

import (
	"fmt"
	"strings"
	"time"
)

// ExpandDatePattern expands an Obsidian-style date format string using the provided time.
//
// It supports:
//   - Obsidian/Moment.js-style tokens (curated subset): YYYY, YY, MMMM, MMM, MM, M, DD, D,
//     dddd, ddd, HH, H, hh, h, mm, m, ss, s, A, a, ZZ, Z, z
//   - Literal blocks wrapped in square brackets, e.g. YYYY-[ToDo]-MM
//   - Legacy "token brace" patterns used in this repo, e.g. {YYYY-MM-DD-HHmmss}
//   - The "zettel" style: YYYYMMDDHHmmss (with or without braces)
//
// Unknown characters are treated literally.
func ExpandDatePattern(pattern string, now time.Time) string {
	s, _ := FormatDatePattern(pattern, now)
	return s
}

// FormatDatePattern is like ExpandDatePattern, but returns an error if the pattern is empty.
func FormatDatePattern(pattern string, now time.Time) (string, error) {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return "", fmt.Errorf("pattern is empty")
	}

	cleaned, literals := stripBracesAndCaptureLiterals(pattern)
	goFormat := convertObsidianFormatToGo(cleaned)
	formatted := now.Format(goFormat)

	for placeholder, literal := range literals {
		formatted = strings.ReplaceAll(formatted, placeholder, literal)
	}

	return formatted, nil
}

func stripBracesAndCaptureLiterals(pattern string) (cleaned string, literals map[string]string) {
	var b strings.Builder
	literals = map[string]string{}

	inBracketLiteral := false
	var literal strings.Builder
	literalIndex := 0

	for _, r := range pattern {
		switch r {
		case '[':
			if inBracketLiteral {
				literal.WriteRune(r)
				continue
			}
			inBracketLiteral = true
			literal.Reset()
		case ']':
			if !inBracketLiteral {
				b.WriteRune(r)
				continue
			}
			inBracketLiteral = false
			placeholder := fmt.Sprintf("\x00LIT%d\x00", literalIndex)
			literalIndex++
			literals[placeholder] = literal.String()
			b.WriteString(placeholder)
		case '{', '}':
			if inBracketLiteral {
				literal.WriteRune(r)
			}
			// Treat braces as formatting sugar (legacy patterns) and ignore them.
		default:
			if inBracketLiteral {
				literal.WriteRune(r)
				continue
			}
			b.WriteRune(r)
		}
	}

	// If a literal block was opened but never closed, treat it literally (keep the '[').
	if inBracketLiteral {
		b.WriteRune('[')
		b.WriteString(literal.String())
	}

	return b.String(), literals
}

func convertObsidianFormatToGo(format string) string {
	// Curated Obsidian/Moment tokens mapped to Go time format.
	tokenMap := map[string]string{
		// Year
		"YYYY": "2006",
		"YY":   "06",

		// Month
		"MMMM": "January",
		"MMM":  "Jan",
		"MM":   "01",
		"M":    "1",

		// Day of month
		"DD": "02",
		"D":  "2",

		// Weekday
		"dddd": "Monday",
		"ddd":  "Mon",

		// Hour
		"HH": "15",
		"H":  "15",
		"hh": "03",
		"h":  "3",

		// Minute
		"mm": "04",
		"m":  "4",

		// Second
		"ss": "05",
		"s":  "5",

		// AM/PM
		"A": "PM",
		"a": "pm",

		// Timezone
		"ZZ": "-0700",
		"Z":  "-07:00",
		"z":  "MST",
	}

	orderedTokens := []string{
		"YYYY", "MMMM", "dddd",
		"MMM", "ddd",
		"YY", "MM", "DD", "HH", "hh", "mm", "ss", "ZZ",
		"M", "D", "H", "h", "m", "s", "A", "a", "Z", "z",
	}

	var out strings.Builder
	i := 0
	for i < len(format) {
		matched := false
		for _, tok := range orderedTokens {
			if i+len(tok) <= len(format) && format[i:i+len(tok)] == tok {
				out.WriteString(tokenMap[tok])
				i += len(tok)
				matched = true
				break
			}
		}
		if matched {
			continue
		}
		out.WriteByte(format[i])
		i++
	}
	return out.String()
}

