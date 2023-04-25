package mocks

type MockUriManager struct {
	ConstructedURI string
	ExecuteErr     error
}

func (m *MockUriManager) Construct(base string, params map[string]string) string {
	return m.ConstructedURI
}

func (m *MockUriManager) Execute(uri string) error {
	return m.ExecuteErr
}
