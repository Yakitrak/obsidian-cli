package mocks

import (
	"github.com/stretchr/testify/mock"
)

type MockFuzzyFinder struct {
	mock.Mock
}

func (f *MockFuzzyFinder) Find(slice interface{}, itemFunc func(i int) string, opts ...interface{}) (int, error) {
	args := f.Called(slice, itemFunc, opts)
	return args.Int(0), args.Error(1)
}

func (f *MockFuzzyFinder) FindMulti(slice interface{}, itemFunc func(i int) string, opts ...interface{}) ([]int, error) {
	args := f.Called(slice, itemFunc, opts)
	return args.Get(0).([]int), args.Error(1)
}
