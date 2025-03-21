package mocks

import (
	"errors"

	"github.com/stretchr/testify/mock"
)

type NoteManager struct {
	mock.Mock
}

func (m *NoteManager) Move(originalPath string, newPath string) error {
	args := m.Called(originalPath, newPath)
	return args.Error(0)
}

func (m *NoteManager) Delete(path string) error {
	args := m.Called(path)
	return args.Error(0)
}

func (m *NoteManager) UpdateLinks(vaultPath string, oldNoteName string, newNoteName string) error {
	args := m.Called(vaultPath, oldNoteName, newNoteName)
	return args.Error(0)
}

func (m *NoteManager) GetContents(vaultPath string, noteName string) (string, error) {
	args := m.Called(vaultPath, noteName)
	return args.String(0), args.Error(1)
}

func (m *NoteManager) GetNotesList(vaultPath string) ([]string, error) {
	args := m.Called(vaultPath)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	if notes, ok := args.Get(0).([]string); ok {
		return notes, args.Error(1)
	}
	return nil, errors.New("invalid type returned for notes list")
}
