package mocks

type MockNoteManager struct {
	ExecuteErr       error
	UpdateLinksError error
}

func (m *MockNoteManager) Delete(path string) error {
	return m.ExecuteErr
}

func (m *MockNoteManager) Move(_ string, _ string) error {
	return m.ExecuteErr
}

func (m *MockNoteManager) UpdateLinks(string, string, string) error {
	return m.UpdateLinksError
}
