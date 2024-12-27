package fuzzyfinder

import (
	"github.com/gdamore/tcell/v2"
)

type screen tcell.Screen

type terminal interface {
	screen
}

type termImpl struct {
	screen
}
