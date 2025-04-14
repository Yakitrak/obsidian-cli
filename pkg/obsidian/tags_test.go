package obsidian

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractFrontmatter(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    map[string]interface{}
		wantErr bool
	}{
		{
			name: "valid frontmatter",
			content: `---
tags: [todo, project]
title: Test Note
---
Content here`,
			want: map[string]interface{}{
				"tags":  []string{"todo", "project"},
				"title": "Test Note",
			},
			wantErr: false,
		},
		{
			name:    "no frontmatter",
			content: "Just content here",
			want:    nil,
			wantErr: false,
		},
		{
			name: "invalid yaml",
			content: `---
tags: [todo, project
title: Test Note
---
Content here`,
			want:    nil,
			wantErr: true,
		},
		{
			name: "nested tags",
			content: `---
tags: [[todo, [project, subproject]]]
---
Content here`,
			want: map[string]interface{}{
				"tags": []string{"todo", "project", "subproject"},
			},
			wantErr: false,
		},
		{
			name: "comma-separated tags",
			content: `---
tags: todo, project, subproject
---
Content here`,
			want: map[string]interface{}{
				"tags": []string{"todo", "project", "subproject"},
			},
			wantErr: false,
		},
		{
			name: "list-style tags",
			content: `---
tags:
  - todo
  - project
  - subproject
---
Content here`,
			want: map[string]interface{}{
				"tags": []string{"todo", "project", "subproject"},
			},
			wantErr: false,
		},
		{
			name: "malformed frontmatter",
			content: `---
tags: [todo, project
---
Content here`,
			want:    nil,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ExtractFrontmatter(tt.content)
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestExtractHashtags(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    []string
	}{
		{
			name:    "multiple hashtags",
			content: "This is a #todo and a #project note",
			want:    []string{"#todo", "#project"},
		},
		{
			name:    "no hashtags",
			content: "Just regular text",
			want:    []string{},
		},
		{
			name:    "hashtag with hyphen",
			content: "A #todo-item and #project-1",
			want:    []string{"#todo-item", "#project-1"},
		},
		{
			name:    "hashtags with special characters",
			content: "This is a #todo! and a #project? note",
			want:    []string{"#todo", "#project"},
		},
		{
			name:    "hashtags with emojis",
			content: "This is a #todo🎯 and a #project🚀 note",
			want:    []string{"#todo", "#project"},
		},
		{
			name:    "hashtags in code blocks",
			content: "```\n#todo\n```\n#project",
			want:    []string{"#project"},
		},
		{
			name:    "hashtags in inline code",
			content: "`#todo` and #project",
			want:    []string{"#project"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ExtractHashtags(tt.content)
			assert.ElementsMatch(t, tt.want, got)
		})
	}
}

func TestHasAnyTags(t *testing.T) {
	tests := []struct {
		name    string
		content string
		tags    []string
		want    bool
	}{
		{
			name: "empty tags list",
			content: `---
tags: [project]
---
Content with #todo`,
			tags: []string{},
			want: false,
		},
		{
			name:    "match hashtag with # prefix",
			content: `Regular content with #todo and #project`,
			tags:    []string{"todo", "project"},
			want:    true,
		},
		{
			name:    "no match without # prefix",
			content: `Regular content with todo and project`,
			tags:    []string{"todo", "project"},
			want:    false,
		},
		{
			name: "match frontmatter array tags",
			content: `---
tags: [todo, project]
---
Regular content`,
			tags: []string{"todo", "project"},
			want: true,
		},
		{
			name: "match frontmatter comma-separated tags",
			content: `---
tags: todo, project
---
Regular content`,
			tags: []string{"todo", "project"},
			want: true,
		},
		{
			name: "match frontmatter single tag",
			content: `---
tags: todo
---
Regular content`,
			tags: []string{"todo", "project"},
			want: true,
		},
		{
			name:    "match with punctuation after hashtag",
			content: `This is a #todo. And a #project, next #task!`,
			tags:    []string{"todo", "project", "task"},
			want:    true,
		},
		{
			name:    "no match for partial tags",
			content: `This is a #todos and #projects`,
			tags:    []string{"todo", "project"},
			want:    false,
		},
		{
			name:    "match with mixed case",
			content: `This is a #TODO and #Project`,
			tags:    []string{"todo", "project"},
			want:    true,
		},
		{
			name:    "match tags with hyphens",
			content: `This is a #todo-item and #project-1`,
			tags:    []string{"todo-item", "project-1"},
			want:    true,
		},
		{
			name:    "no match for tags in code blocks",
			content: "```\n#todo\n```\nRegular content",
			tags:    []string{"todo"},
			want:    false,
		},
		{
			name:    "no match for tags in inline code",
			content: "This is `#todo` in code",
			tags:    []string{"todo"},
			want:    false,
		},
		{
			name:    "match tag with underscore",
			content: `This is a #todo_item`,
			tags:    []string{"todo_item"},
			want:    true,
		},
		{
			name:    "match tag with numbers",
			content: `This is a #todo123`,
			tags:    []string{"todo123"},
			want:    true,
		},
		{
			name:    "match tag at start of line",
			content: `#todo at start`,
			tags:    []string{"todo"},
			want:    true,
		},
		{
			name:    "match tag at end of line",
			content: `End with #todo`,
			tags:    []string{"todo"},
			want:    true,
		},
		{
			name:    "match tag with unicode characters",
			content: `This is a #todó and #プロジェクト`,
			tags:    []string{"todó", "プロジェクト"},
			want:    true,
		},
		{
			name:    "match any of multiple tags",
			content: `This is a #todo note`,
			tags:    []string{"project", "todo", "task"},
			want:    true,
		},
		{
			name: "match with mixed frontmatter and hashtags",
			content: `---
tags: [project]
---
This is a #todo note`,
			tags: []string{"todo", "project"},
			want: true,
		},
		{
			name:    "no match when similar but not exact",
			content: `This is a #todolist note`,
			tags:    []string{"todo"},
			want:    false,
		},
		// Hierarchical tag matching cases
		{
			name:    "search foo matches #foo, #foo/bar, #foo/bar/baz",
			content: `#foo some text #foo/bar more #foo/bar/baz` ,
			tags:    []string{"foo"},
			want:    true,
		},
		{
			name:    "search foo/bar matches #foo/bar and #foo/bar/baz, not #foo",
			content: `#foo some text #foo/bar more #foo/bar/baz` ,
			tags:    []string{"foo/bar"},
			want:    true,
		},
		{
			name:    "search foo/bar/baz matches only #foo/bar/baz",
			content: `#foo some text #foo/bar more #foo/bar/baz` ,
			tags:    []string{"foo/bar/baz"},
			want:    true,
		},
		{
			name:    "search foo/bar does not match #foo",
			content: `#foo only` ,
			tags:    []string{"foo/bar"},
			want:    false,
		},
		// Frontmatter hierarchical
		{
			name:    "search foo matches foo, foo/bar, foo/bar/baz in frontmatter",
			content: `---
tags: [foo, foo/bar, foo/bar/baz]
---
Body` ,
			tags:    []string{"foo"},
			want:    true,
		},
		{
			name:    "search foo/bar matches foo/bar and foo/bar/baz in frontmatter, not foo",
			content: `---
tags: [foo, foo/bar, foo/bar/baz]
---
Body` ,
			tags:    []string{"foo/bar"},
			want:    true,
		},
		{
			name:    "search foo/bar/baz matches only foo/bar/baz in frontmatter",
			content: `---
tags: [foo, foo/bar, foo/bar/baz]
---
Body` ,
			tags:    []string{"foo/bar/baz"},
			want:    true,
		},
		{
			name:    "search foo/bar does not match foo in frontmatter",
			content: `---
tags: [foo]
---
Body` ,
			tags:    []string{"foo/bar"},
			want:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := HasAnyTags(tt.content, tt.tags)
			assert.Equal(t, tt.want, got, "hasAnyTags(%v, %v)", tt.content, tt.tags)
		})
	}
}