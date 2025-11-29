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

func TestExtractInlinePropertiesSkipsBulletsAndCode(t *testing.T) {
	content := "Notes:: Good\n- Activity:: Should be ignored\n* Win:: Ignored\n+ Gratitude:: Ignored\n1. Numbered:: Ignored\n| Table:: Ignored\n![:clap:: should-be-ignored\n* ![:pray:][image184]4:+1:: also ignore\n> Quote:: Ignored\n```" + "\nCode:: Ignore\n```" + "\nValidKey_1:: Kept\n"
	props := ExtractInlineProperties(content)
	if _, ok := props["- Activity"]; ok {
		t.Fatalf("bullet line should be ignored")
	}
	if _, ok := props["![:clap"]; ok {
		t.Fatalf("emoji key should be ignored")
	}
	if val, ok := props["ValidKey_1"]; !ok || len(val) != 1 || val[0] != "Kept" {
		t.Fatalf("expected ValidKey_1 to be parsed, got %+v", props["ValidKey_1"])
	}
}

func TestExtractInlinePropertiesRequiresNoSpacesInKey(t *testing.T) {
	content := "White :: Neutral and objective\nRed :: Emotion\nBlue::Good\n"
	props := ExtractInlineProperties(content)
	if _, ok := props["White"]; ok {
		t.Fatalf("key with spaces before :: should be ignored")
	}
	if _, ok := props["Red"]; ok {
		t.Fatalf("key with spaces before :: should be ignored")
	}
	if val, ok := props["Blue"]; !ok || len(val) != 1 || val[0] != "Good" {
		t.Fatalf("expected Blue to be parsed, got %+v", props["Blue"])
	}
}
