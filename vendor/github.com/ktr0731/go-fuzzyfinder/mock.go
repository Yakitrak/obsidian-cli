package fuzzyfinder

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/gdamore/tcell/v2"
	runewidth "github.com/mattn/go-runewidth"
	"github.com/nsf/termbox-go"
)

type cell struct {
	ch     rune
	bg, fg termbox.Attribute
}

type simScreen tcell.SimulationScreen

// TerminalMock is a mocked terminal for testing.
// Most users should use it by calling UseMockedTerminal.
type TerminalMock struct {
	simScreen
	sizeMu        sync.RWMutex
	width, height int

	eventsMu sync.Mutex
	events   []termbox.Event

	cellsMu sync.RWMutex
	cells   []*cell

	resultMu sync.RWMutex
	result   string

	sleepDuration time.Duration
	v2            bool
}

// SetSize changes the pseudo-size of the window.
// Note that SetSize resets added cells.
func (m *TerminalMock) SetSize(w, h int) {
	if m.v2 {
		m.simScreen.SetSize(w, h)
		return
	}
	m.sizeMu.Lock()
	defer m.sizeMu.Unlock()
	m.cellsMu.Lock()
	defer m.cellsMu.Unlock()
	m.width = w
	m.height = h
	m.cells = make([]*cell, w*h)
}

// Deprecated: Use SetEventsV2
// SetEvents sets all events, which are fetched by pollEvent.
// A user of this must set the EscKey event at the end.
func (m *TerminalMock) SetEvents(events ...termbox.Event) {
	m.eventsMu.Lock()
	defer m.eventsMu.Unlock()
	m.events = events
}

// SetEventsV2 sets all events, which are fetched by pollEvent.
// A user of this must set the EscKey event at the end.
func (m *TerminalMock) SetEventsV2(events ...tcell.Event) {
	for _, event := range events {
		switch event := event.(type) {
		case *tcell.EventKey:
			ek := event
			m.simScreen.InjectKey(ek.Key(), ek.Rune(), ek.Modifiers())
		case *tcell.EventResize:
			er := event
			w, h := er.Size()
			m.simScreen.SetSize(w, h)
		}
	}
}

// GetResult returns a flushed string that is displayed to the actual terminal.
// It contains all escape sequences such that ANSI escape code.
func (m *TerminalMock) GetResult() string {
	if !m.v2 {
		m.resultMu.RLock()
		defer m.resultMu.RUnlock()
		return m.result
	}

	var s string

	// set cursor for snapshot test
	setCursor := func() {
		cursorX, cursorY, _ := m.simScreen.GetCursor()
		mainc, _, _, _ := m.simScreen.GetContent(cursorX, cursorY)
		if mainc == ' ' {
			m.simScreen.SetContent(cursorX, cursorY, '\u2588', nil, tcell.StyleDefault.Foreground(tcell.ColorWhite).Background(tcell.ColorDefault))
		} else {
			m.simScreen.SetContent(cursorX, cursorY, mainc, nil, tcell.StyleDefault.Background(tcell.ColorWhite))
		}
		m.simScreen.Show()
	}

	setCursor()

	m.resultMu.Lock()

	cells, width, height := m.simScreen.GetContents()

	for h := 0; h < height; h++ {
		prevFg, prevBg := tcell.ColorDefault, tcell.ColorDefault

		for w := 0; w < width; w++ {
			cell := cells[h*width+w]
			fg, bg, attr := cell.Style.Decompose()
			if fg != prevFg || bg != prevBg {
				prevFg, prevBg = fg, bg

				s += "\x1b\x5b\x6d" // Reset previous color.
				v := parseAttrV2(fg, bg, attr)
				s += v
			}

			s += string(cell.Runes)
			rw := runewidth.RuneWidth(cell.Runes[0])
			if rw != 0 {
				w += rw - 1
			}
		}
		s += "\n"
	}
	s += "\x1b\x5b\x6d" // Reset previous color.

	m.resultMu.Unlock()

	return s
}

func (m *TerminalMock) init() error {
	return nil
}

func (m *TerminalMock) size() (width int, height int) {
	m.sizeMu.RLock()
	defer m.sizeMu.RUnlock()
	return m.width, m.height
}

func (m *TerminalMock) clear(fg termbox.Attribute, bg termbox.Attribute) error {
	// TODO
	return nil
}

func (m *TerminalMock) setCell(x int, y int, ch rune, fg termbox.Attribute, bg termbox.Attribute) {
	m.sizeMu.RLock()
	defer m.sizeMu.RUnlock()
	m.cellsMu.Lock()
	defer m.cellsMu.Unlock()

	if x < 0 || x >= m.width {
		return
	}
	if y < 0 || y >= m.height {
		return
	}
	m.cells[y*m.width+x] = &cell{ch: ch, fg: fg, bg: bg}
}

func (m *TerminalMock) setCursor(x int, y int) {
	m.sizeMu.RLock()
	defer m.sizeMu.RUnlock()
	m.cellsMu.Lock()
	defer m.cellsMu.Unlock()
	if x < 0 || x >= m.width {
		return
	}
	if y < 0 || y >= m.height {
		return
	}
	i := y*m.width + x
	if m.cells[i] == nil {
		m.cells[y*m.width+x] = &cell{ch: '\u2588', fg: termbox.ColorWhite, bg: termbox.ColorDefault}
	} else {
		// Cursor on a rune.
		m.cells[y*m.width+x].bg = termbox.ColorWhite
	}
}

func (m *TerminalMock) pollEvent() termbox.Event {
	m.eventsMu.Lock()
	defer m.eventsMu.Unlock()
	if len(m.events) == 0 {
		panic("pollEvent called with empty events. have you set expected events by SetEvents?")
	}
	e := m.events[0]
	m.events = m.events[1:]
	// Wait a moment for goroutine scheduling.
	time.Sleep(m.sleepDuration)
	return e
}

// flush displays all items with formatted layout.
func (m *TerminalMock) flush() {
	m.cellsMu.RLock()

	var s string
	for j := 0; j < m.height; j++ {
		prevFg, prevBg := termbox.ColorDefault, termbox.ColorDefault
		for i := 0; i < m.width; i++ {
			c := m.cells[j*m.width+i]
			if c == nil {
				s += " "
				prevFg, prevBg = termbox.ColorDefault, termbox.ColorDefault
				continue
			} else {
				var fgReset bool
				if c.fg != prevFg {
					s += "\x1b\x5b\x6d" // Reset previous color.
					s += parseAttr(c.fg, true)
					prevFg = c.fg
					prevBg = termbox.ColorDefault
					fgReset = true
				}
				if c.bg != prevBg {
					if !fgReset {
						s += "\x1b\x5b\x6d" // Reset previous color.
						prevFg = termbox.ColorDefault
					}
					s += parseAttr(c.bg, false)
					prevBg = c.bg
				}
				s += string(c.ch)
				rw := runewidth.RuneWidth(c.ch)
				if rw != 0 {
					i += rw - 1
				}
			}
		}
		s += "\n"
	}
	s += "\x1b\x5b\x6d" // Reset previous color.

	m.cellsMu.RUnlock()
	m.cellsMu.Lock()
	m.cells = make([]*cell, m.width*m.height)
	m.cellsMu.Unlock()

	m.resultMu.Lock()
	defer m.resultMu.Unlock()

	m.result = s
}

func (m *TerminalMock) close() {}

// UseMockedTerminal switches the terminal, which is used from
// this package to a mocked one.
func UseMockedTerminal() *TerminalMock {
	f := newFinder()
	return f.UseMockedTerminal()
}

// UseMockedTerminalV2 switches the terminal, which is used from
// this package to a mocked one.
func UseMockedTerminalV2() *TerminalMock {
	f := newFinder()
	return f.UseMockedTerminalV2()
}

func (f *finder) UseMockedTerminal() *TerminalMock {
	m := &TerminalMock{}
	f.term = m
	return m
}

func (f *finder) UseMockedTerminalV2() *TerminalMock {
	screen := tcell.NewSimulationScreen("UTF-8")
	if err := screen.Init(); err != nil {
		panic(err)
	}
	m := &TerminalMock{
		simScreen: screen,
		v2:        true,
	}
	f.term = m
	return m
}

// parseAttr parses an attribute of termbox
// as an escape sequence.
// parseAttr doesn't support output modes othar than color256 in termbox-go.
func parseAttr(attr termbox.Attribute, isFg bool) string {
	var buf bytes.Buffer
	buf.WriteString("\x1b[")
	if attr >= termbox.AttrReverse {
		buf.WriteString("7;")
		attr -= termbox.AttrReverse
	}
	if attr >= termbox.AttrUnderline {
		buf.WriteString("4;")
		attr -= termbox.AttrUnderline
	}
	if attr >= termbox.AttrBold {
		buf.WriteString("1;")
		attr -= termbox.AttrBold
	}

	if attr > termbox.ColorWhite {
		panic(fmt.Sprintf("invalid color code: %d", attr))
	}

	if attr == termbox.ColorDefault {
		if isFg {
			buf.WriteString("39")
		} else {
			buf.WriteString("49")
		}
	} else {
		color := int(attr) - 1
		if isFg {
			fmt.Fprintf(&buf, "38;5;%d", color)
		} else {
			fmt.Fprintf(&buf, "48;5;%d", color)
		}
	}
	buf.WriteString("m")

	return buf.String()
}

// parseAttrV2 parses color and attribute for testing.
func parseAttrV2(fg, bg tcell.Color, attr tcell.AttrMask) string {
	if attr == tcell.AttrInvalid {
		panic("invalid attribute")
	}

	var params []string
	if attr&tcell.AttrBold == tcell.AttrBold {
		params = append(params, "1")
		attr ^= tcell.AttrBold
	}
	if attr&tcell.AttrBlink == tcell.AttrBlink {
		params = append(params, "5")
		attr ^= tcell.AttrBlink
	}
	if attr&tcell.AttrReverse == tcell.AttrReverse {
		params = append(params, "7")
		attr ^= tcell.AttrReverse
	}
	if attr&tcell.AttrUnderline == tcell.AttrUnderline {
		params = append(params, "4")
		attr ^= tcell.AttrUnderline
	}
	if attr&tcell.AttrDim == tcell.AttrDim {
		params = append(params, "2")
		attr ^= tcell.AttrDim
	}
	if attr&tcell.AttrItalic == tcell.AttrItalic {
		params = append(params, "3")
		attr ^= tcell.AttrItalic
	}
	if attr&tcell.AttrStrikeThrough == tcell.AttrStrikeThrough {
		params = append(params, "9")
		attr ^= tcell.AttrStrikeThrough
	}

	switch {
	case fg == 0: // Ignore.
	case fg == tcell.ColorDefault:
		params = append(params, "39")
	case fg > tcell.Color255:
		r, g, b := fg.RGB()
		params = append(params, "38", "2", fmt.Sprint(r), fmt.Sprint(g), fmt.Sprint(b))
	default:
		params = append(params, "38", "5", fmt.Sprint(fg-tcell.ColorValid))
	}

	switch {
	case bg == 0: // Ignore.
	case bg == tcell.ColorDefault:
		params = append(params, "49")
	case bg > tcell.Color255:
		r, g, b := bg.RGB()
		params = append(params, "48", "2", fmt.Sprint(r), fmt.Sprint(g), fmt.Sprint(b))
	default:
		params = append(params, "48", "5", fmt.Sprint(bg-tcell.ColorValid))
	}

	return fmt.Sprintf("\x1b[%sm", strings.Join(params, ";"))
}

func toAnsi3bit(color tcell.Color) int {
	colors := []tcell.Color{
		tcell.ColorBlack, tcell.ColorRed, tcell.ColorGreen, tcell.ColorYellow, tcell.ColorBlue, tcell.ColorDarkMagenta, tcell.ColorDarkCyan, tcell.ColorWhite,
	}
	for i, c := range colors {
		if c == color {
			return i
		}
	}
	return 0
}
