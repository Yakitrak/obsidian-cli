package mocks

type MockNoteManager struct {
	DeleteErr        error
	MoveErr          error
	UpdateLinksError error
}

func (m *MockNoteManager) Delete(_ string) error {
	return m.DeleteErr
}

func (m *MockNoteManager) Move(_ string, _ string) error {
	return m.MoveErr
}

func (m *MockNoteManager) UpdateLinks(string, string, string) error {
	return m.UpdateLinksError
}
