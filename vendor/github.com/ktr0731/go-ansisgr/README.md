# go-ansisgr

[![PkgGoDev](https://pkg.go.dev/badge/github.com/ktr0731/go-ansisgr)](https://pkg.go.dev/github.com/ktr0731/go-ansisgr)
[![GitHub Actions](https://github.com/ktr0731/go-ansisgr/workflows/main/badge.svg)](https://github.com/ktr0731/go-ansisgr/actions)
[![codecov](https://codecov.io/gh/ktr0731/go-ansisgr/branch/master/graph/badge.svg?token=6IHRfCBs7K)](https://codecov.io/gh/ktr0731/go-ansisgr)  

`go-ansisgr` provides a SGR (Select Graphic Rendition, a part of ANSI Escape Sequence) parser.

- 16 colors, 256 colors and RGB colors support
- All attributes support

## Installation
``` bash
go get github.com/ktr0731/go-fuzzyfinder
```

## Usage
`ansisgr.NewIterator` is the only entry-point API. This function returns an iteratorw which consumes the passed string.

``` go
in := "a\x1b[1;31mb"
iter := ansisgr.NewIterator(in)

for {
	r, style, ok := iter.Next()
	if !ok {
		break
	}

        // do something.
}
```

`r` is a rune, and `style` is the foreground/background color and attributes `r` has.

``` go
if color, ok := style.Foreground(); ok {
	// Foreground color is specified.
}
if color, ok := style.Background(); ok {
	// Background color is specified.
}
```

The attribute method reports whether `r` has the attribute.

``` go
style.Bold()
style.Italic()
```

## go-ansisgr with gdamore/tcell
`go-ansisgr` is useful when you construct a rich terminal user interface by using `gdamore/tcell` or others. Although `gdamore/tcell` and other libraries provides SGR functionality from their API, they doesn't support "raw" strings which contain ANSI Escape Sequence. Therefore, `go-ansisgr` translates these strings and makes it easy to use the TUI library's functionality.  
For example, the following code displays colored string powered by `gdamore/tcell`.

```go
func main() {
	screen, err := tcell.NewScreen()
	if err != nil {
		log.Fatal(err)
	}
	defer screen.Fini()

	if err := screen.Init(); err != nil {
		log.Fatal(err)
	}

	in := "\x1b[38;2;100;200;200mhello, \x1b[0;1;30;48;5;245mworld!"
	iter := ansisgr.NewIterator(in)
	for i := 0; ; i++ {
		r, rstyle, ok := iter.Next()
		if !ok {
			break
		}

		style := tcell.StyleDefault
		if color, ok := rstyle.Foreground(); ok {
			switch color.Mode() {
			case ansisgr.Mode16:
				style = style.Foreground(tcell.PaletteColor(color.Value() - 30))
			case ansisgr.Mode256:
				style = style.Foreground(tcell.PaletteColor(color.Value()))
			case ansisgr.ModeRGB:
				r, g, b := color.RGB()
				style = style.Foreground(tcell.NewRGBColor(int32(r), int32(g), int32(b)))
			}
		}
		if color, valid := rstyle.Background(); valid {
			switch color.Mode() {
			case ansisgr.Mode16:
				style = style.Background(tcell.PaletteColor(color.Value() - 40))
			case ansisgr.Mode256:
				style = style.Background(tcell.PaletteColor(color.Value()))
			case ansisgr.ModeRGB:
				r, g, b := color.RGB()
				style = style.Background(tcell.NewRGBColor(int32(r), int32(g), int32(b)))
			}
		}

		style = style.
			Bold(rstyle.Bold()).
			Dim(rstyle.Dim()).
			Italic(rstyle.Italic()).
			Underline(rstyle.Underline()).
			Blink(rstyle.Blink()).
			Reverse(rstyle.Reverse()).
			StrikeThrough(rstyle.Strikethrough())

		screen.SetContent(i, 0, r, nil, style)
	}

	screen.Show()

	time.Sleep(3 * time.Second)
}
```
