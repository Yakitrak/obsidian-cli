package obsidian

import (
	"errors"
	"github.com/ktr0731/go-fuzzyfinder"
)

type FuzzyFinder struct{}

type FuzzyFinderManager interface {
	Find(slice interface{}, itemFunc func(i int) string, opts ...interface{}) (int, error)
}

func (f *FuzzyFinder) Find(slice interface{}, itemFunc func(i int) string, opts ...interface{}) (int, error) {
	items, ok := slice.([]string)
	if !ok {
		return -1, errors.New("invalid slice type, expected []string")
	}

	index, err := fuzzyfinder.Find(items, func(i int) string {
		return itemFunc(i)
	})
	if err != nil {
		return -1, errors.New(NoteDoesNotExistError)
	}
	return index, nil
}
