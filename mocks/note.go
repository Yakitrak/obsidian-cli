package mocks

type MockNoteManager struct {
	DeleteErr        error
	MoveErr          error
	UpdateLinksError error
	GetContentsError error
}

func (m *MockNoteManager) Delete(string) error {
	return m.DeleteErr
}

func (m *MockNoteManager) Move(string, string) error {
	return m.MoveErr
}

func (m *MockNoteManager) UpdateLinks(string, string, string) error {
	return m.UpdateLinksError
}

func (m *MockNoteManager) GetContents(string, string) (string, error) {
	return "example contents", m.GetContentsError
}

func (m *MockNoteManager) GetNotesList(string) ([]string, error) {
	return []string{"note1", "note2", "note3"}, m.GetContentsError
}
