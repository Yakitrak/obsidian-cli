// go-ansisgr provides an SGR (Select Graphic Rendition, a part of ANSI Escape Sequence) parser.
package ansisgr

import (
	"unicode"
)

const (
	attrReset = 0
	attrBold  = 1 << iota
	attrDim
	attrItalic
	attrUnderline
	attrBlink
	attrReverse
	attrInvisible
	attrStrikethrough
)

const (
	colorValid = 1 << (24 + iota)
	colorIs16
	colorIs256
	colorIsRGB
)

// Style represents a set of attributes, foreground color, and background color.
type Style struct {
	attr       int
	foreground int
	background int
}

// Mode represents a color syntax.
type Mode int

const (
	// ModeNone is used to when color is not specified.
	ModeNone Mode = iota
	// Mode16 represents 16 colors.
	Mode16
	// Mode256 represents 256 colors.
	Mode256
	// ModeRGB represents RGB colors (a.k.a. True Color).
	ModeRGB
)

// Color represents color value specified by SGR.
type Color struct{ v int }

// Mode returns the color's Mode.
func (c Color) Mode() Mode {
	switch {
	case c.v&colorIs16 == colorIs16:
		return Mode16
	case c.v&colorIs256 == colorIs256:
		return Mode256
	case c.v&colorIsRGB == colorIsRGB:
		return ModeRGB
	}

	return ModeNone
}

// RGB returns red, green, and blue color values. It assumes that c.Mode() == ModeRGB.
func (c Color) RGB() (int, int, int) {
	return 0xff0000 & c.v >> 16, 0x00ff00 & c.v >> 8, 0x0000ff & c.v
}

// Value returns the color value represented by int.
// For example, '\x1b[31m' is 31, '\x1b[38;5;116m' is 116, and '\x1b[38;2;10;20;30' is 660510 (0x0a141e).
func (c Color) Value() int { return c.v & 0xffffff }

// Foreground returns the foreground color. the second return value indicates whether the color is valid or not.
func (s *Style) Foreground() (Color, bool) {
	return Color{v: s.foreground}, s.foreground&colorValid == colorValid
}

// Background returns the background color. the second return value indicates whether the color is valid or not.
func (s *Style) Background() (Color, bool) {
	return Color{v: s.background}, s.background&colorValid == colorValid
}

// Bold indicates whether bold is enabled.
func (s *Style) Bold() bool { return s.attr&attrBold == attrBold }

// Dim indicates whether dim (faint) is enabled.
func (s *Style) Dim() bool { return s.attr&attrDim == attrDim }

// Italic indicates whether italic is enabled.
func (s *Style) Italic() bool { return s.attr&attrItalic == attrItalic }

// Underline indicates whether underline is enabled.
func (s *Style) Underline() bool { return s.attr&attrUnderline == attrUnderline }

// Blink indicates whether blink is enabled.
func (s *Style) Blink() bool { return s.attr&attrBlink == attrBlink }

// Reverse indicates whether reverse (hidden) is enabled.
func (s *Style) Reverse() bool { return s.attr&attrReverse == attrReverse }

// Invisible indicates whether invisible is enabled.
func (s *Style) Invisible() bool { return s.attr&attrInvisible == attrInvisible }

// Strikethrough indicates whether strikethrough is enabled.
func (s *Style) Strikethrough() bool { return s.attr&attrStrikethrough == attrStrikethrough }

// Iterator is an iterator over a parsed string.
type Iterator struct {
	runes []rune
	i     int
	style Style
}

// NewIterator returns a new Iterator.
func NewIterator(s string) *Iterator {
	return &Iterator{i: -1, runes: []rune(s)}
}

// Next returns a next rune and its style. the third return value indicates there is a next value.
func (a *Iterator) Next() (rune, Style, bool) {
	for a.next() {
		r := a.runes[a.i]
		if r != 0x1b {
			return r, a.style, true
		}

		if ok := a.next(); ok && a.runes[a.i] != '[' {
			continue
		}

		args := a.consumeSequence()
	LOOP:
		for len(args) != 0 {
			if args[0] == 38 {
				offset, consumed := consumeAs256OrRGB(args)
				args = args[offset:]
				if consumed.valid {
					a.style.foreground = consumed.v
					continue
				}
			} else if args[0] == 48 {
				offset, consumed := consumeAs256OrRGB(args)
				args = args[offset:]
				if consumed.valid {
					a.style.background = consumed.v
					continue
				}
			}

			for _, arg := range args {
				switch arg {
				case 38, 48:
					if len(args) != 1 {
						continue LOOP
					}

					// Ignore when the arg is the last element.
				case 0:
					a.style.attr = attrReset
					a.style.foreground = 0 | colorIs16
					a.style.background = 0 | colorIs16
				case 39:
					a.style.foreground = 0 | colorIs16
				case 49:
					a.style.background = 0 | colorIs16
				case 1:
					a.style.attr |= attrBold
				case 2:
					a.style.attr |= attrDim
				case 3:
					a.style.attr |= attrItalic
				case 4:
					a.style.attr |= attrUnderline
				case 5:
					a.style.attr |= attrBlink
				case 7:
					a.style.attr |= attrReverse
				case 8:
					a.style.attr |= attrInvisible
				case 9:
					a.style.attr |= attrStrikethrough
				case 22:
					if a.style.attr&attrBold == attrBold {
						a.style.attr ^= attrBold
					}
					if a.style.attr&attrDim == attrDim {
						a.style.attr ^= attrDim
					}
				case 23:
					a.style.attr ^= attrItalic
				case 24:
					a.style.attr ^= attrUnderline
				case 25:
					a.style.attr ^= attrBlink
				case 27:
					a.style.attr ^= attrReverse
				case 28:
					a.style.attr ^= attrInvisible
				case 29:
					a.style.attr ^= attrStrikethrough
				default:
					switch {
					case (arg >= 30 && arg <= 37):
						a.style.foreground = arg | colorIs16 | colorValid
					case arg >= 40 && arg <= 47:
						a.style.background = arg | colorIs16 | colorValid
					case arg >= 90 && arg <= 97:
						a.style.foreground = arg | colorIs16 | colorValid
					case arg >= 100 && arg <= 107:
						a.style.background = arg | colorIs16 | colorValid
					}
				}

				args = args[1:]
			}
		}
	}

	return 0, Style{}, false
}

func (a *Iterator) next() bool {
	a.i++
	return a.i < len(a.runes)
}

func (a *Iterator) consumeSequence() []int {
	var args []int
	var val int

	for a.next() {
		switch a.runes[a.i] {
		case 'm':
			return append(args, val)
		case ';':
			args = append(args, val)
			val = 0
		default:
			if !unicode.IsDigit(a.runes[a.i]) {
				return nil // Broken sequence, ignore
			}

			val = val*10 + int(a.runes[a.i]-'0')
		}
	}

	return nil
}

type consumedValue struct {
	v     int
	valid bool
}

func consumeAs256OrRGB(args []int) (offset int, v consumedValue) {
	if len(args) < 2 || (args[0] != 38 && args[0] != 48) {
		return 0, v
	}

	switch args[1] {
	case 5: // 256
		if l := len(args); l < 3 {
			return l, v
		}

		if args[2] > 255 {
			return 3, v
		}

		return 3, consumedValue{
			v:     args[2] | colorValid | colorIs256,
			valid: true,
		}
	case 2: // RGB
		if l := len(args); l < 5 {
			return l, v
		}

		var val int
		for i := 0; i < 3; i++ {
			if args[i+2] > 255 {
				return i + 2 + 1, v
			}

			val |= args[i+2] << (8 * (2 - i))
		}

		return 5, consumedValue{
			v:     val | colorValid | colorIsRGB,
			valid: true,
		}
	}

	return 2, v
}
