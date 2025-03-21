package obsidian

import (
	"errors"

	"github.com/ktr0731/go-fuzzyfinder"
)

// FuzzyFinder implements the FuzzyFinderManager interface
type FuzzyFinder struct{}

// FuzzyFinderManager defines the interface for fuzzy finding functionality
type FuzzyFinderManager interface {
	Find(slice interface{}, itemFunc func(i int) string, opts ...interface{}) (int, error)
	FindMulti(slice interface{}, itemFunc func(i int) string, opts ...interface{}) ([]int, error)
}

// Find performs a fuzzy search on a slice and returns a single selected index
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

// FindMulti performs a fuzzy search on a slice and returns multiple selected indices
func (f *FuzzyFinder) FindMulti(slice interface{}, itemFunc func(i int) string, opts ...interface{}) ([]int, error) {
	items, ok := slice.([]string)
	if !ok {
		return nil, errors.New("invalid slice type, expected []string")
	}

	indices, err := fuzzyfinder.FindMulti(items, func(i int) string {
		return itemFunc(i)
	})
	if err != nil {
		return nil, errors.New(NoteDoesNotExistError)
	}
	return indices, nil
}
