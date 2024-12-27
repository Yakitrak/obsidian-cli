// Package fuzzyfinder provides terminal user interfaces for fuzzy-finding.
//
// Note that, all functions are not goroutine-safe.
package fuzzyfinder

import (
	"context"
	"flag"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/gdamore/tcell/v2"
	"github.com/ktr0731/go-ansisgr"
	"github.com/ktr0731/go-fuzzyfinder/matching"
	runewidth "github.com/mattn/go-runewidth"
	"github.com/pkg/errors"
)

var (
	// ErrAbort is returned from Find* functions if there are no selections.
	ErrAbort   = errors.New("abort")
	errEntered = errors.New("entered")
)

// Finds the minimum value among the arguments
func min(vars ...int) int {
	min := vars[0]

	for _, i := range vars {
		if min > i {
			min = i
		}
	}

	return min
}

type state struct {
	items      []string           // All item names.
	allMatched []matching.Matched // All items.
	matched    []matching.Matched // Matched items against the input.

	// x is the current index of the prompt line.
	x int
	// cursorX is the position of prompt line.
	// Note that cursorX is the actual width of input runes.
	cursorX int

	// The current index of filtered items (matched).
	// The initial value is 0.
	y int
	// cursorY is the position of item line.
	// Note that the max size of cursorY depends on max height.
	cursorY int

	input []rune

	// selections holds whether a key is selected or not. Each key is
	// an index of an item (Matched.Idx). Each value represents the position
	// which it is selected.
	selection map[int]int
	// selectionIdx holds the next index, which is used to a selection's value.
	selectionIdx int
}

type finder struct {
	term      terminal
	stateMu   sync.RWMutex
	state     state
	drawTimer *time.Timer
	eventCh   chan struct{}
	opt       *opt

	termEventsChan <-chan tcell.Event
}

func newFinder() *finder {
	return &finder{}
}

func (f *finder) initFinder(items []string, matched []matching.Matched, opt opt) error {
	if f.term == nil {
		screen, err := tcell.NewScreen()
		if err != nil {
			return errors.Wrap(err, "failed to new screen")
		}
		f.term = &termImpl{
			screen: screen,
		}
		if err := f.term.Init(); err != nil {
			return errors.Wrap(err, "failed to initialize screen")
		}

		eventsChan := make(chan tcell.Event)
		go f.term.ChannelEvents(eventsChan, nil)
		f.termEventsChan = eventsChan
	}

	f.opt = &opt
	f.state = state{}

	if opt.multi {
		f.state.selection = map[int]int{}
	}

	f.state.items = items
	f.state.matched = matched
	f.state.allMatched = matched

	if opt.beginAtTop {
		f.state.cursorY = len(f.state.matched) - 1
		f.state.y = len(f.state.matched) - 1
	}

	if !isInTesting() {
		f.drawTimer = time.AfterFunc(0, func() {
			f.stateMu.Lock()
			f._draw()
			f._drawPreview()
			f.stateMu.Unlock()
			f.term.Show()
		})
		f.drawTimer.Stop()
	}
	f.eventCh = make(chan struct{}, 30) // A large value

	if opt.query != "" {
		f.state.input = []rune(opt.query)
		f.state.cursorX = runewidth.StringWidth(opt.query)
		f.state.x = len(opt.query)
		f.filter()
	}

	return nil
}

func (f *finder) updateItems(items []string, matched []matching.Matched) {
	f.stateMu.Lock()
	f.state.items = items
	f.state.matched = matched
	f.state.allMatched = matched
	f.stateMu.Unlock()
	f.eventCh <- struct{}{}
}

// _draw is used from draw with a timer.
func (f *finder) _draw() {
	width, height := f.term.Size()
	f.term.Clear()

	maxWidth := width
	if f.opt.previewFunc != nil {
		maxWidth = width/2 - 1
	}

	maxHeight := height

	// prompt line
	var promptLinePad int

	for _, r := range f.opt.promptString {
		style := tcell.StyleDefault.
			Foreground(tcell.ColorBlue).
			Background(tcell.ColorDefault)

		f.term.SetContent(promptLinePad, maxHeight-1, r, nil, style)
		promptLinePad++
	}
	var r rune
	var w int
	for _, r = range f.state.input {
		style := tcell.StyleDefault.
			Foreground(tcell.ColorDefault).
			Background(tcell.ColorDefault).
			Bold(true)

		// Add a space between '>' and runes.
		f.term.SetContent(promptLinePad+w, maxHeight-1, r, nil, style)
		w += runewidth.RuneWidth(r)
	}
	f.term.ShowCursor(promptLinePad+f.state.cursorX, maxHeight-1)

	maxHeight--

	// Header line
	if len(f.opt.header) > 0 {
		w = 0
		for _, r := range runewidth.Truncate(f.opt.header, maxWidth-2, "..") {
			style := tcell.StyleDefault.
				Foreground(tcell.ColorGreen).
				Background(tcell.ColorDefault)
			f.term.SetContent(2+w, maxHeight-1, r, nil, style)
			w += runewidth.RuneWidth(r)
		}
		maxHeight--
	}

	// Number line
	for i, r := range fmt.Sprintf("%d/%d", len(f.state.matched), len(f.state.items)) {
		style := tcell.StyleDefault.
			Foreground(tcell.ColorYellow).
			Background(tcell.ColorDefault)

		f.term.SetContent(2+i, maxHeight-1, r, nil, style)
	}
	maxHeight--

	// Item lines
	itemAreaHeight := maxHeight - 1
	matched := f.state.matched
	offset := f.state.cursorY
	y := f.state.y
	// From the first (the most bottom) item in the item lines to the end.
	matched = matched[y-offset:]

	for i, m := range matched {
		if i > itemAreaHeight {
			break
		}
		if i == f.state.cursorY {
			style := tcell.StyleDefault.
				Foreground(tcell.ColorRed).
				Background(tcell.ColorBlack)

			f.term.SetContent(0, maxHeight-1-i, '>', nil, style)
			f.term.SetContent(1, maxHeight-1-i, ' ', nil, style)
		}

		if f.opt.multi {
			if _, ok := f.state.selection[m.Idx]; ok {
				style := tcell.StyleDefault.
					Foreground(tcell.ColorRed).
					Background(tcell.ColorBlack)

				f.term.SetContent(1, maxHeight-1-i, '>', nil, style)
			}
		}

		var posIdx int
		w := 2
		for j, r := range []rune(f.state.items[m.Idx]) {
			style := tcell.StyleDefault.
				Foreground(tcell.ColorDefault).
				Background(tcell.ColorDefault)
			// Highlight selected strings.
			hasHighlighted := false
			if posIdx < len(f.state.input) {
				from, to := m.Pos[0], m.Pos[1]
				if !(from == -1 && to == -1) && (from <= j && j <= to) {
					if unicode.ToLower(f.state.input[posIdx]) == unicode.ToLower(r) {
						style = tcell.StyleDefault.
							Foreground(tcell.ColorGreen).
							Background(tcell.ColorDefault)
						hasHighlighted = true
						posIdx++
					}
				}
			}
			if i == f.state.cursorY {
				if hasHighlighted {
					style = tcell.StyleDefault.
						Foreground(tcell.ColorDarkCyan).
						Bold(true).
						Background(tcell.ColorBlack)
				} else {
					style = tcell.StyleDefault.
						Foreground(tcell.ColorYellow).
						Bold(true).
						Background(tcell.ColorBlack)
				}
			}

			rw := runewidth.RuneWidth(r)
			// Shorten item cells.
			if w+rw+2 > maxWidth {
				f.term.SetContent(w, maxHeight-1-i, '.', nil, style)
				f.term.SetContent(w+1, maxHeight-1-i, '.', nil, style)
				break
			} else {
				f.term.SetContent(w, maxHeight-1-i, r, nil, style)
				w += rw
			}
		}
	}
}

func (f *finder) _drawPreview() {
	if f.opt.previewFunc == nil {
		return
	}

	width, height := f.term.Size()
	var idx int
	if len(f.state.matched) == 0 {
		idx = -1
	} else {
		idx = f.state.matched[f.state.y].Idx
	}

	iter := ansisgr.NewIterator(f.opt.previewFunc(idx, width, height))

	// top line
	for i := width / 2; i < width; i++ {
		var r rune
		switch {
		case i == width/2:
			r = '┌'
		case i == width-1:
			r = '┐'
		default:
			r = '─'
		}

		style := tcell.StyleDefault.
			Foreground(tcell.ColorBlack).
			Background(tcell.ColorDefault)

		f.term.SetContent(i, 0, r, nil, style)
	}
	// bottom line
	for i := width / 2; i < width; i++ {
		var r rune
		switch {
		case i == width/2:
			r = '└'
		case i == width-1:
			r = '┘'
		default:
			r = '─'
		}

		style := tcell.StyleDefault.
			Foreground(tcell.ColorBlack).
			Background(tcell.ColorDefault)

		f.term.SetContent(i, height-1, r, nil, style)
	}
	// Start with h=1 to exclude each corner rune.
	const vline = '│'
	var wvline = runewidth.RuneWidth(vline)
	for h := 1; h < height-1; h++ {
		// donePreviewLine indicates the preview string of the current line identified by h is already drawn.
		var donePreviewLine bool
		w := width / 2
		for i := width / 2; i < width; i++ {
			switch {
			// Left vertical line.
			case i == width/2:
				style := tcell.StyleDefault.
					Foreground(tcell.ColorBlack).
					Background(tcell.ColorDefault)
				f.term.SetContent(i, h, vline, nil, style)
				w += wvline
			// Right vertical line.
			case i == width-1:
				style := tcell.StyleDefault.
					Foreground(tcell.ColorBlack).
					Background(tcell.ColorDefault)
				f.term.SetContent(i, h, vline, nil, style)
				w += wvline
			// Spaces between left and right vertical lines.
			case w == width/2+wvline, w == width-1-wvline:
				style := tcell.StyleDefault.
					Foreground(tcell.ColorDefault).
					Background(tcell.ColorDefault)

				f.term.SetContent(w, h, ' ', nil, style)
				w++
			default: // Preview text
				if donePreviewLine {
					continue
				}

				r, rstyle, ok := iter.Next()
				if !ok || r == '\n' {
					// Consumed all preview characters.
					donePreviewLine = true
					continue
				}

				rw := runewidth.RuneWidth(r)
				if w+rw > width-1-2 {
					donePreviewLine = true

					// Discard the rest of the current line.
					consumeIterator(iter, '\n')

					style := tcell.StyleDefault.
						Foreground(tcell.ColorDefault).
						Background(tcell.ColorDefault)

					f.term.SetContent(w, h, '.', nil, style)
					f.term.SetContent(w+1, h, '.', nil, style)

					w += 2
					continue
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
				f.term.SetContent(w, h, r, nil, style)
				w += rw
			}
		}
	}
}

func (f *finder) draw(d time.Duration) {
	f.stateMu.RLock()
	defer f.stateMu.RUnlock()

	if isInTesting() {
		// Don't use goroutine scheduling.
		f._draw()
		f._drawPreview()
		f.term.Show()
	} else {
		f.drawTimer.Reset(d)
	}
}

// readKey reads a key input.
// It returns ErrAbort if esc, CTRL-C or CTRL-D keys are inputted,
// errEntered in case of enter key, and a context error when the passed
// context is cancelled.
func (f *finder) readKey(ctx context.Context) error {
	f.stateMu.RLock()
	prevInputLen := len(f.state.input)
	f.stateMu.RUnlock()
	defer func() {
		f.stateMu.RLock()
		currentInputLen := len(f.state.input)
		f.stateMu.RUnlock()
		if prevInputLen != currentInputLen {
			f.eventCh <- struct{}{}
		}
	}()

	var e tcell.Event

	select {
	case ee := <-f.termEventsChan:
		e = ee
	case <-ctx.Done():
		return ctx.Err()
	}

	f.stateMu.Lock()
	defer f.stateMu.Unlock()

	_, screenHeight := f.term.Size()
	matchedLinesCount := len(f.state.matched)

	// Max number of lines to scroll by using PgUp and PgDn
	var pageScrollBy = screenHeight - 3

	switch e := e.(type) {
	case *tcell.EventKey:
		switch e.Key() {
		case tcell.KeyEsc, tcell.KeyCtrlC, tcell.KeyCtrlD:
			return ErrAbort
		case tcell.KeyBackspace, tcell.KeyBackspace2:
			if len(f.state.input) == 0 {
				return nil
			}
			if f.state.x == 0 {
				return nil
			}
			x := f.state.x
			f.state.cursorX -= runewidth.RuneWidth(f.state.input[x-1])
			f.state.x--
			f.state.input = append(f.state.input[:x-1], f.state.input[x:]...)
		case tcell.KeyDelete:
			if f.state.x == len(f.state.input) {
				return nil
			}
			x := f.state.x

			f.state.input = append(f.state.input[:x], f.state.input[x+1:]...)
		case tcell.KeyEnter:
			return errEntered
		case tcell.KeyLeft, tcell.KeyCtrlB:
			if f.state.x > 0 {
				f.state.cursorX -= runewidth.RuneWidth(f.state.input[f.state.x-1])
				f.state.x--
			}
		case tcell.KeyRight, tcell.KeyCtrlF:
			if f.state.x < len(f.state.input) {
				f.state.cursorX += runewidth.RuneWidth(f.state.input[f.state.x])
				f.state.x++
			}
		case tcell.KeyCtrlA, tcell.KeyHome:
			f.state.cursorX = 0
			f.state.x = 0
		case tcell.KeyCtrlE, tcell.KeyEnd:
			f.state.cursorX = runewidth.StringWidth(string(f.state.input))
			f.state.x = len(f.state.input)
		case tcell.KeyCtrlW:
			in := f.state.input[:f.state.x]
			inStr := string(in)
			pos := strings.LastIndex(strings.TrimRightFunc(inStr, unicode.IsSpace), " ")
			if pos == -1 {
				f.state.input = []rune{}
				f.state.cursorX = 0
				f.state.x = 0
				return nil
			}
			pos = utf8.RuneCountInString(inStr[:pos])
			newIn := f.state.input[:pos+1]
			f.state.input = newIn
			f.state.cursorX = runewidth.StringWidth(string(newIn))
			f.state.x = len(newIn)
		case tcell.KeyCtrlU:
			f.state.input = f.state.input[f.state.x:]
			f.state.cursorX = 0
			f.state.x = 0
		case tcell.KeyUp, tcell.KeyCtrlK, tcell.KeyCtrlP:
			if f.state.y+1 < matchedLinesCount {
				f.state.y++
			}
			if f.state.cursorY+1 < min(matchedLinesCount, screenHeight-2) {
				f.state.cursorY++
			}
		case tcell.KeyDown, tcell.KeyCtrlJ, tcell.KeyCtrlN:
			if f.state.y > 0 {
				f.state.y--
			}
			if f.state.cursorY-1 >= 0 {
				f.state.cursorY--
			}
		case tcell.KeyPgUp:
			f.state.y += min(pageScrollBy, matchedLinesCount-1-f.state.y)
			maxCursorY := min(screenHeight-3, matchedLinesCount-1)
			f.state.cursorY += min(pageScrollBy, maxCursorY-f.state.cursorY)
		case tcell.KeyPgDn:
			f.state.y -= min(pageScrollBy, f.state.y)
			f.state.cursorY -= min(pageScrollBy, f.state.cursorY)
		case tcell.KeyTab:
			if !f.opt.multi {
				return nil
			}
			idx := f.state.matched[f.state.y].Idx
			if _, ok := f.state.selection[idx]; ok {
				delete(f.state.selection, idx)
			} else {
				f.state.selection[idx] = f.state.selectionIdx
				f.state.selectionIdx++
			}
			if f.state.y > 0 {
				f.state.y--
			}
			if f.state.cursorY > 0 {
				f.state.cursorY--
			}
		default:
			if e.Rune() != 0 {
				width, _ := f.term.Size()
				maxLineWidth := width - 2 - 1
				if len(f.state.input)+1 > maxLineWidth {
					// Discard inputted rune.
					return nil
				}

				x := f.state.x
				f.state.input = append(f.state.input[:x], append([]rune{e.Rune()}, f.state.input[x:]...)...)
				f.state.cursorX += runewidth.RuneWidth(e.Rune())
				f.state.x++
			}
		}
	case *tcell.EventResize:
		f.term.Clear()

		width, height := f.term.Size()
		itemAreaHeight := height - 2 - 1
		if itemAreaHeight >= 0 && f.state.cursorY > itemAreaHeight {
			f.state.cursorY = itemAreaHeight
		}

		maxLineWidth := width - 2 - 1
		if maxLineWidth < 0 {
			f.state.input = nil
			f.state.cursorX = 0
			f.state.x = 0
		} else if len(f.state.input)+1 > maxLineWidth {
			// Discard inputted rune.
			f.state.input = f.state.input[:maxLineWidth]
			f.state.cursorX = runewidth.StringWidth(string(f.state.input))
			f.state.x = maxLineWidth
		}
	}
	return nil
}

func (f *finder) filter() {
	f.stateMu.RLock()
	if len(f.state.input) == 0 {
		f.stateMu.RUnlock()
		f.stateMu.Lock()
		defer f.stateMu.Unlock()
		f.state.matched = f.state.allMatched
		return
	}

	// TODO: If input is not delete operation, it is able to
	// reduce total iteration.
	// FindAll may take a lot of time, so it is desired to use RLock to avoid goroutine blocking.
	matchedItems := matching.FindAll(string(f.state.input), f.state.items, matching.WithMode(matching.Mode(f.opt.mode)))
	f.stateMu.RUnlock()

	f.stateMu.Lock()
	defer f.stateMu.Unlock()
	f.state.matched = matchedItems
	if len(f.state.matched) == 0 {
		f.state.cursorY = 0
		f.state.y = 0
		return
	}

	switch {
	case f.state.cursorY >= len(f.state.matched):
		f.state.cursorY = len(f.state.matched) - 1
		f.state.y = len(f.state.matched) - 1
	case f.state.y >= len(f.state.matched):
		f.state.y = len(f.state.matched) - 1
	}
}

func (f *finder) find(slice interface{}, itemFunc func(i int) string, opts []Option) ([]int, error) {
	if itemFunc == nil {
		return nil, errors.New("itemFunc must not be nil")
	}

	opt := defaultOption
	for _, o := range opts {
		o(&opt)
	}

	rv := reflect.ValueOf(slice)
	if opt.hotReload && (rv.Kind() != reflect.Ptr || reflect.Indirect(rv).Kind() != reflect.Slice) {
		return nil, errors.Errorf("the first argument must be a pointer to a slice, but got %T", slice)
	} else if !opt.hotReload && rv.Kind() != reflect.Slice {
		return nil, errors.Errorf("the first argument must be a slice, but got %T", slice)
	}

	makeItems := func(sliceLen int) ([]string, []matching.Matched) {
		items := make([]string, sliceLen)
		matched := make([]matching.Matched, sliceLen)
		for i := 0; i < sliceLen; i++ {
			items[i] = itemFunc(i)
			matched[i] = matching.Matched{Idx: i} //nolint:exhaustivestruct
		}
		return items, matched
	}

	var (
		items   []string
		matched []matching.Matched
	)

	var parentContext context.Context
	if opt.context != nil {
		parentContext = opt.context
	} else {
		parentContext = context.Background()
	}

	ctx, cancel := context.WithCancel(parentContext)
	defer cancel()

	inited := make(chan struct{})
	if opt.hotReload && rv.Kind() == reflect.Ptr {
		opt.hotReloadLock.Lock()
		rvv := reflect.Indirect(rv)
		items, matched = makeItems(rvv.Len())
		opt.hotReloadLock.Unlock()

		go func() {
			<-inited

			var prev int
			for {
				select {
				case <-ctx.Done():
					return
				case <-time.After(30 * time.Millisecond):
					opt.hotReloadLock.Lock()
					curr := rvv.Len()
					if prev != curr {
						items, matched = makeItems(curr)
						f.updateItems(items, matched)
					}
					opt.hotReloadLock.Unlock()
					prev = curr
				}
			}
		}()
	} else {
		items, matched = makeItems(rv.Len())
	}

	if err := f.initFinder(items, matched, opt); err != nil {
		return nil, errors.Wrap(err, "failed to initialize the fuzzy finder")
	}

	if !isInTesting() {
		defer f.term.Fini()
	}

	close(inited)

	if opt.selectOne && len(f.state.matched) == 1 {
		return []int{f.state.matched[0].Idx}, nil
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case <-f.eventCh:
				f.filter()
				f.draw(0)
			}
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			f.draw(10 * time.Millisecond)

			err := f.readKey(ctx)
			// hack for earning time to filter exec
			if isInTesting() {
				time.Sleep(50 * time.Millisecond)
			}
			switch {
			case errors.Is(err, ErrAbort):
				return nil, ErrAbort
			case errors.Is(err, errEntered):
				f.stateMu.RLock()
				defer f.stateMu.RUnlock()

				if len(f.state.matched) == 0 {
					return nil, ErrAbort
				}
				if f.opt.multi {
					if len(f.state.selection) == 0 {
						return []int{f.state.matched[f.state.y].Idx}, nil
					}
					poss, idxs := make([]int, 0, len(f.state.selection)), make([]int, 0, len(f.state.selection))
					for idx, pos := range f.state.selection {
						idxs = append(idxs, idx)
						poss = append(poss, pos)
					}
					sort.Slice(idxs, func(i, j int) bool {
						return poss[i] < poss[j]
					})
					return idxs, nil
				}
				return []int{f.state.matched[f.state.y].Idx}, nil
			case err != nil:
				return nil, errors.Wrap(err, "failed to read a key")
			}
		}
	}
}

// Find displays a UI that provides fuzzy finding against the provided slice.
// The argument slice must be of a slice type. If not, Find returns
// an error. itemFunc is called by the length of slice. previewFunc is called
// when the cursor which points to the currently selected item is changed.
// If itemFunc is nil, Find returns an error.
//
// itemFunc receives an argument i, which is the index of the item currently
// selected.
//
// Find returns ErrAbort if a call to Find is finished with no selection.
func Find(slice interface{}, itemFunc func(i int) string, opts ...Option) (int, error) {
	f := newFinder()
	return f.Find(slice, itemFunc, opts...)
}

func (f *finder) Find(slice interface{}, itemFunc func(i int) string, opts ...Option) (int, error) {
	res, err := f.find(slice, itemFunc, opts)

	if err != nil {
		return 0, err
	}
	return res[0], err
}

// FindMulti is nearly the same as Find. The only difference from Find is that
// the user can select multiple items at once, by using the tab key.
func FindMulti(slice interface{}, itemFunc func(i int) string, opts ...Option) ([]int, error) {
	f := newFinder()
	return f.FindMulti(slice, itemFunc, opts...)
}

func (f *finder) FindMulti(slice interface{}, itemFunc func(i int) string, opts ...Option) ([]int, error) {
	opts = append(opts, withMulti())
	res, err := f.find(slice, itemFunc, opts)
	return res, err
}

func isInTesting() bool {
	return flag.Lookup("test.v") != nil
}

func consumeIterator(iter *ansisgr.Iterator, r rune) {
	for {
		r, _, ok := iter.Next()
		if !ok || r == '\n' {
			return
		}
	}
}
