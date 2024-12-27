package mocks

type MockFuzzyFinder struct {
	SelectedIndex int
	FindErr       error
}

func (f *MockFuzzyFinder) Find(slice interface{}, itemFunc func(i int) string, opts ...interface{}) (int, error) {
	if f.FindErr != nil {
		return -1, f.FindErr
	}
	return f.SelectedIndex, nil
}
