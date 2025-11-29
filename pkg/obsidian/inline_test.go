package obsidian

import "testing"

func TestExtractInlineProperties(t *testing.T) {
	content := `---
title: Test
tags: [a]
---

Office:: [[AORD]]
Notes:: Some text
Invalid: line

Archetype:: [[Explorer]]
Office:: [[AOGR]]
`
	props := ExtractInlineProperties(content)
	if len(props["Office"]) != 2 || props["Office"][0] != "[[AORD]]" || props["Office"][1] != "[[AOGR]]" {
		t.Fatalf("unexpected office props: %+v", props["Office"])
	}
	if len(props["Archetype"]) != 1 || props["Archetype"][0] != "[[Explorer]]" {
		t.Fatalf("unexpected archetype props: %+v", props["Archetype"])
	}
	if _, ok := props["Invalid"]; ok {
		t.Fatalf("invalid line should not be parsed")
	}
}
