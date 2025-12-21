package mocks

import "github.com/Yakitrak/obsidian-cli/pkg/obsidian"

type MockNoteManager struct {
	DeleteErr        error
	MoveErr          error
	UpdateLinksError error
	GetContentsError error
	SetContentsError error
	NoMatches        bool
	Contents         string
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
	if m.Contents != "" {
		return m.Contents, m.GetContentsError
	}
	return "example contents", m.GetContentsError
}

func (m *MockNoteManager) SetContents(string, string, string) error {
	return m.SetContentsError
}

func (m *MockNoteManager) GetNotesList(string) ([]string, error) {
	return []string{"note1", "note2", "note3"}, m.GetContentsError
}

func (m *MockNoteManager) SearchNotesWithSnippets(string, string) ([]obsidian.NoteMatch, error) {
	if m.GetContentsError != nil {
		return nil, m.GetContentsError
	}
	if m.NoMatches {
		return []obsidian.NoteMatch{}, nil
	}
	return []obsidian.NoteMatch{
		{FilePath: "note1.md", LineNumber: 5, MatchLine: "example match line"},
		{FilePath: "note2.md", LineNumber: 10, MatchLine: "another match"},
	}, nil
}
