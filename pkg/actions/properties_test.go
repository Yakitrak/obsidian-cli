package actions

import (
	"reflect"
	"testing"

	"github.com/atomicobject/obsidian-cli/mocks"
)

func TestProperties(t *testing.T) {
	notes := map[string]string{
		"alpha.md": `---
office: AOGR
reviewed: true
links: ["[[Alpha]]", "[[Beta|B]]"]
url: https://example.com
count: 5
---`,
		"beta.md": `---
office: AORD
reviewed: false
links:
  - "[[Alpha]]"
  - "[[Gamma]]"
count: 7
date: 2024-05-02
---`,
		"gamma.md": `---
office: AOGR
note: some text
---`,
	}

	vaultManager := &mocks.VaultManager{}
	noteManager := &mocks.NoteManager{}

	vaultManager.On("Path").Return("/mock/vault", nil)

	var notesList []string
	for notePath := range notes {
		notesList = append(notesList, notePath)
	}
	noteManager.On("GetNotesList", "/mock/vault").Return(notesList, nil)
	for path, content := range notes {
		noteManager.On("GetContents", "/mock/vault", path).Return(content, nil)
	}

	result, err := Properties(vaultManager, noteManager, PropertiesOptions{
		ExcludeTags: false,
		ValueLimit:  5,
		MaxValues:   50,
		Notes:       []string{"alpha.md", "beta.md", "gamma.md"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	expected := []PropertySummary{
		{
			Name:               "office",
			NoteCount:          3,
			Shape:              "scalar",
			ValueType:          "string",
			EnumValues:         []string{"AOGR", "AORD"},
			DistinctValueCount: 2,
		},
		{
			Name:               "count",
			NoteCount:          2,
			Shape:              "scalar",
			ValueType:          "int",
			EnumValues:         []string{"5", "7"},
			DistinctValueCount: 2,
		},
		{
			Name:               "links",
			NoteCount:          2,
			Shape:              "list",
			ValueType:          "wikilink",
			EnumValues:         []string{"[[Alpha]]", "[[Beta|B]]", "[[Gamma]]"},
			DistinctValueCount: 3,
		},
		{
			Name:               "reviewed",
			NoteCount:          2,
			Shape:              "scalar",
			ValueType:          "bool",
			EnumValues:         []string{"false", "true"},
			DistinctValueCount: 2,
		},
		{
			Name:               "date",
			NoteCount:          1,
			Shape:              "scalar",
			ValueType:          "date",
			EnumValues:         nil,
			DistinctValueCount: 1,
		},
		{
			Name:               "note",
			NoteCount:          1,
			Shape:              "scalar",
			ValueType:          "string",
			EnumValues:         []string{"some text"},
			DistinctValueCount: 1,
		},
		{
			Name:               "url",
			NoteCount:          1,
			Shape:              "scalar",
			ValueType:          "url",
			EnumValues:         nil,
			DistinctValueCount: 1,
		},
	}

	if len(result) != len(expected) {
		t.Fatalf("expected %d summaries, got %d", len(expected), len(result))
	}

	for i, got := range result {
		want := expected[i]
		if got.Name != want.Name || got.NoteCount != want.NoteCount || got.Shape != want.Shape || got.ValueType != want.ValueType {
			t.Fatalf("summary %d mismatch: got %+v want %+v", i, got, want)
		}
		if !reflect.DeepEqual(got.EnumValues, want.EnumValues) {
			t.Fatalf("enum values mismatch for %s: got %v want %v", got.Name, got.EnumValues, want.EnumValues)
		}
		if got.DistinctValueCount != want.DistinctValueCount {
			t.Fatalf("distinct count mismatch for %s: got %d want %d", got.Name, got.DistinctValueCount, want.DistinctValueCount)
		}
	}
}

func TestPropertiesExcludeTags(t *testing.T) {
	notes := map[string]string{
		"note.md": `---
tags: [project, work]
custom: yes
---`,
	}

	vaultManager := &mocks.VaultManager{}
	noteManager := &mocks.NoteManager{}

	vaultManager.On("Path").Return("/mock/vault", nil)
	noteManager.On("GetNotesList", "/mock/vault").Return([]string{"note.md"}, nil)
	noteManager.On("GetContents", "/mock/vault", "note.md").Return(notes["note.md"], nil)

	result, err := Properties(vaultManager, noteManager, PropertiesOptions{
		ExcludeTags: true,
		ValueLimit:  10,
		MaxValues:   10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, s := range result {
		if s.Name == "tags" {
			t.Fatalf("expected tags property to be excluded")
		}
	}
}

func TestPropertiesIncludesTagsByDefault(t *testing.T) {
	notes := map[string]string{
		"note.md": `---
tags: [project, work]
---`,
	}

	vaultManager := &mocks.VaultManager{}
	noteManager := &mocks.NoteManager{}

	vaultManager.On("Path").Return("/mock/vault", nil)
	noteManager.On("GetNotesList", "/mock/vault").Return([]string{"note.md"}, nil)
	noteManager.On("GetContents", "/mock/vault", "note.md").Return(notes["note.md"], nil)

	result, err := Properties(vaultManager, noteManager, PropertiesOptions{
		ValueLimit: 10,
		MaxValues:  10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	foundTags := false
	for _, s := range result {
		if s.Name == "tags" && len(s.EnumValues) == 2 {
			foundTags = true
			break
		}
	}
	if !foundTags {
		t.Fatalf("expected tags property to be included by default")
	}
}

func TestPropertiesWithSubset(t *testing.T) {
	notes := map[string]string{
		"one.md": `---
prop: a
---`,
		"two.md": `---
prop: b
---`,
	}

	vaultManager := &mocks.VaultManager{}
	noteManager := &mocks.NoteManager{}

	vaultManager.On("Path").Return("/mock/vault", nil)
	noteManager.On("GetNotesList", "/mock/vault").Return([]string{"one.md", "two.md"}, nil)
	noteManager.On("GetContents", "/mock/vault", "one.md").Return(notes["one.md"], nil)
	noteManager.On("GetContents", "/mock/vault", "two.md").Return(notes["two.md"], nil)

	result, err := Properties(vaultManager, noteManager, PropertiesOptions{
		Notes:      []string{"one.md"},
		ValueLimit: 10,
		MaxValues:  10,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 || result[0].Name != "prop" || result[0].EnumValues[0] != "a" {
		t.Fatalf("expected only subset property, got %+v", result)
	}
}

func TestPropertiesOnlyFiltersProperties(t *testing.T) {
	notes := map[string]string{
		"note.md": `---
keep: front
drop: skip
---
keep:: inline
drop:: nope
`,
	}

	vaultManager := &mocks.VaultManager{}
	noteManager := &mocks.NoteManager{}

	vaultManager.On("Path").Return("/mock/vault", nil)
	noteManager.On("GetNotesList", "/mock/vault").Return([]string{"note.md"}, nil)
	noteManager.On("GetContents", "/mock/vault", "note.md").Return(notes["note.md"], nil)

	result, err := Properties(vaultManager, noteManager, PropertiesOptions{
		ValueLimit: 10,
		MaxValues:  10,
		Only:       []string{"keep"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(result) != 1 {
		t.Fatalf("expected 1 property, got %d", len(result))
	}
	if result[0].Name != "keep" {
		t.Fatalf("expected only 'keep' property, got %s", result[0].Name)
	}
	if result[0].DistinctValueCount != 2 {
		t.Fatalf("expected two values for keep, got %d", result[0].DistinctValueCount)
	}
}

func TestPropertiesMaxValuesFollowsValueLimit(t *testing.T) {
	notes := map[string]string{
		"one.md": `---
prop: a
---`,
		"two.md": `---
prop: b
---`,
		"three.md": `---
prop: c
---`,
		"four.md": `---
prop: d
---`,
	}

	vaultManager := &mocks.VaultManager{}
	noteManager := &mocks.NoteManager{}

	vaultManager.On("Path").Return("/mock/vault", nil)
	noteManager.On("GetNotesList", "/mock/vault").Return([]string{"one.md", "two.md", "three.md", "four.md"}, nil)
	for path, content := range notes {
		noteManager.On("GetContents", "/mock/vault", path).Return(content, nil)
	}

	result, err := Properties(vaultManager, noteManager, PropertiesOptions{
		ValueLimit: 4,
		MaxValues:  3, // deliberately low; should be raised to valueLimit+1 internally
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	for _, summary := range result {
		if summary.Name == "prop" {
			if summary.TruncatedValueSet {
				t.Fatalf("expected prop values not to be truncated")
			}
			if len(summary.EnumValues) != 4 {
				t.Fatalf("expected 4 enum values, got %d", len(summary.EnumValues))
			}
			return
		}
	}
	t.Fatalf("expected prop summary to be present")
}

func TestPropertiesForceEnumMixed(t *testing.T) {
	notes := map[string]string{
		"one.md": `---
prop: a
---`,
		"two.md": `---
prop: 1
---`,
	}

	vaultManager := &mocks.VaultManager{}
	noteManager := &mocks.NoteManager{}

	vaultManager.On("Path").Return("/mock/vault", nil)
	noteManager.On("GetNotesList", "/mock/vault").Return([]string{"one.md", "two.md"}, nil)
	noteManager.On("GetContents", "/mock/vault", "one.md").Return(notes["one.md"], nil)
	noteManager.On("GetContents", "/mock/vault", "two.md").Return(notes["two.md"], nil)

	result, err := Properties(vaultManager, noteManager, PropertiesOptions{
		ValueLimit:     10,
		MaxValues:      10,
		ForceEnumMixed: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var found bool
	for _, s := range result {
		if s.Name == "prop" {
			found = true
			if len(s.EnumValues) != 2 {
				t.Fatalf("expected enum values for mixed property, got %+v", s.EnumValues)
			}
		}
	}
	if !found {
		t.Fatalf("expected prop property to be present")
	}
}

func TestPropertiesEnumValueCounts(t *testing.T) {
	notes := map[string]string{
		"one.md": `---
office: AOG
---`,
		"two.md": `---
office: AOG
---`,
	}

	vaultManager := &mocks.VaultManager{}
	noteManager := &mocks.NoteManager{}

	vaultManager.On("Path").Return("/mock/vault", nil)
	noteManager.On("GetNotesList", "/mock/vault").Return([]string{"one.md", "two.md"}, nil)
	noteManager.On("GetContents", "/mock/vault", "one.md").Return(notes["one.md"], nil)
	noteManager.On("GetContents", "/mock/vault", "two.md").Return(notes["two.md"], nil)

	result, err := Properties(vaultManager, noteManager, PropertiesOptions{
		ValueLimit:         5,
		MaxValues:          10,
		IncludeValueCounts: true,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	var found bool
	for _, s := range result {
		if s.Name == "office" {
			found = true
			if s.EnumValueCounts["AOG"] != 2 {
				t.Fatalf("expected count 2 for AOG, got %d", s.EnumValueCounts["AOG"])
			}
		}
	}
	if !found {
		t.Fatalf("expected office property")
	}
}
