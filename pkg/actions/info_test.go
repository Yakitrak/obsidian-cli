package actions

import (
	"errors"
	"testing"

	"github.com/Yakitrak/obsidian-cli/mocks"
	"github.com/stretchr/testify/assert"
)

func TestGetFileInfo(t *testing.T) {
	tests := []struct {
		name          string
		filePath      string
		fileContent   string
		expectedInfo  *FileInfo
		expectedError error
		setupMocks    func(*mocks.VaultManager, *mocks.NoteManager)
	}{
		{
			name:     "file with frontmatter and hashtags",
			filePath: "test.md",
			fileContent: `---
title: Test Note
tags: [project, task]
---
This is a #todo note with #project-related content`,
			expectedInfo: &FileInfo{
				Frontmatter: map[string]interface{}{
					"title": "Test Note",
					"tags":  []string{"project", "task"},
				},
				Tags: []string{"project", "task", "todo", "project-related"},
			},
			setupMocks: func(v *mocks.VaultManager, n *mocks.NoteManager) {
				v.On("Path").Return("/test/vault", nil)
				n.On("GetContents", "/test/vault", "test.md").Return(`---
title: Test Note
tags: [project, task]
---
This is a #todo note with #project-related content`, nil)
			},
		},
		{
			name:        "file with only hashtags",
			filePath:    "test.md",
			fileContent: `This is a #todo note with #project tags`,
			expectedInfo: &FileInfo{
				Frontmatter: nil,
				Tags:        []string{"todo", "project"},
			},
			setupMocks: func(v *mocks.VaultManager, n *mocks.NoteManager) {
				v.On("Path").Return("/test/vault", nil)
				n.On("GetContents", "/test/vault", "test.md").Return(`This is a #todo note with #project tags`, nil)
			},
		},
		{
			name:     "file with only frontmatter tags",
			filePath: "test.md",
			fileContent: `---
tags: [todo, project]
---
Regular content without hashtags`,
			expectedInfo: &FileInfo{
				Frontmatter: map[string]interface{}{
					"tags": []string{"todo", "project"},
				},
				Tags: []string{"todo", "project"},
			},
			setupMocks: func(v *mocks.VaultManager, n *mocks.NoteManager) {
				v.On("Path").Return("/test/vault", nil)
				n.On("GetContents", "/test/vault", "test.md").Return(`---
tags: [todo, project]
---
Regular content without hashtags`, nil)
			},
		},
		{
			name:        "file with no tags",
			filePath:    "test.md",
			fileContent: `Just regular content`,
			expectedInfo: &FileInfo{
				Frontmatter: nil,
				Tags:        []string{},
			},
			setupMocks: func(v *mocks.VaultManager, n *mocks.NoteManager) {
				v.On("Path").Return("/test/vault", nil)
				n.On("GetContents", "/test/vault", "test.md").Return(`Just regular content`, nil)
			},
		},
		{
			name:          "vault path error",
			filePath:      "test.md",
			expectedError: errors.New("vault path error"),
			setupMocks: func(v *mocks.VaultManager, n *mocks.NoteManager) {
				v.On("Path").Return("", errors.New("vault path error"))
			},
		},
		{
			name:          "file not found",
			filePath:      "test.md",
			expectedError: errors.New("file not found"),
			setupMocks: func(v *mocks.VaultManager, n *mocks.NoteManager) {
				v.On("Path").Return("/test/vault", nil)
				n.On("GetContents", "/test/vault", "test.md").Return("", errors.New("file not found"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockVault := &mocks.VaultManager{}
			mockNote := &mocks.NoteManager{}
			tt.setupMocks(mockVault, mockNote)

			info, err := GetFileInfo(mockVault, mockNote, tt.filePath)

			if tt.expectedError != nil {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedError.Error(), err.Error())
				assert.Nil(t, info)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expectedInfo.Frontmatter, info.Frontmatter)
				assert.ElementsMatch(t, tt.expectedInfo.Tags, info.Tags)
			}

			mockVault.AssertExpectations(t)
			mockNote.AssertExpectations(t)
		})
	}
}
