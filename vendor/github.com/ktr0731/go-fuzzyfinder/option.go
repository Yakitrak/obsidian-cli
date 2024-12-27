package fuzzyfinder

import (
	"context"
	"sync"
)

type opt struct {
	mode          mode
	previewFunc   func(i, width, height int) string
	multi         bool
	hotReload     bool
	hotReloadLock sync.Locker
	promptString  string
	header        string
	beginAtTop    bool
	context       context.Context
	query         string
	selectOne     bool
}

type mode int

const (
	// ModeSmart enables a smart matching. It is the default matching mode.
	// At the beginning, matching mode is ModeCaseInsensitive, but it switches
	// over to ModeCaseSensitive if an upper case character is inputted.
	ModeSmart mode = iota
	// ModeCaseSensitive enables a case-sensitive matching.
	ModeCaseSensitive
	// ModeCaseInsensitive enables a case-insensitive matching.
	ModeCaseInsensitive
)

var defaultOption = opt{
	promptString:  "> ",
	hotReloadLock: &sync.Mutex{}, // this won't resolve the race condition but avoid nil panic
}

// Option represents available fuzzy-finding options.
type Option func(*opt)

// WithMode specifies a matching mode. The default mode is ModeSmart.
func WithMode(m mode) Option {
	return func(o *opt) {
		o.mode = m
	}
}

// WithPreviewWindow enables to display a preview for the selected item.
// The argument f receives i, width and height. i is the same as Find's one.
// width and height are the size of the terminal so that you can use these to adjust
// a preview content. Note that width and height are calculated as a rune-based length.
//
// If there is no selected item, previewFunc passes -1 to previewFunc.
//
// If f is nil, the preview feature is disabled.
func WithPreviewWindow(f func(i, width, height int) string) Option {
	return func(o *opt) {
		o.previewFunc = f
	}
}

// WithHotReload reloads the passed slice automatically when some entries are appended.
// The caller must pass a pointer of the slice instead of the slice itself.
//
// Deprecated: use WithHotReloadLock instead.
func WithHotReload() Option {
	return func(o *opt) {
		o.hotReload = true
	}
}

// WithHotReloadLock reloads the passed slice automatically when some entries are appended.
// The caller must pass a pointer of the slice instead of the slice itself.
// The caller must pass a RLock which is used to synchronize access to the slice.
// The caller MUST NOT lock in the itemFunc passed to Find / FindMulti because it will be locked by the fuzzyfinder.
// If used together with WithPreviewWindow, the caller MUST use the RLock only in the previewFunc passed to WithPreviewWindow.
func WithHotReloadLock(lock sync.Locker) Option {
	return func(o *opt) {
		o.hotReload = true
		o.hotReloadLock = lock
	}
}

type cursorPosition int

const (
	CursorPositionBottom cursorPosition = iota
	CursorPositionTop
)

// WithCursorPosition sets the initial position of the cursor
func WithCursorPosition(position cursorPosition) Option {
	return func(o *opt) {
		switch position {
		case CursorPositionTop:
			o.beginAtTop = true
		case CursorPositionBottom:
			o.beginAtTop = false
		}
	}
}

// WithPromptString changes the prompt string. The default value is "> ".
func WithPromptString(s string) Option {
	return func(o *opt) {
		o.promptString = s
	}
}

// withMulti enables to select multiple items by tab key.
func withMulti() Option {
	return func(o *opt) {
		o.multi = true
	}
}

// WithHeader enables to set the header.
func WithHeader(s string) Option {
	return func(o *opt) {
		o.header = s
	}
}

// WithContext enables closing the fuzzy finder from parent.
func WithContext(ctx context.Context) Option {
	return func(o *opt) {
		o.context = ctx
	}
}

// WithQuery enables to set the initial query.
func WithQuery(s string) Option {
	return func(o *opt) {
		o.query = s
	}
}

// WithQuery enables to set the initial query.
func WithSelectOne() Option {
	return func(o *opt) {
		o.selectOne = true
	}
}
