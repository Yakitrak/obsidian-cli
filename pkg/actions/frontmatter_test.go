package actions

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

// testVault is a local stub implementing obsidian.VaultManager behavior we need.
// It simply returns a fixed path for the vault and a fixed default name.
type testVault struct{ path string }

func (t testVault) DefaultName() (string, error)     { return "default", nil }
func (t testVault) SetDefaultName(name string) error { return nil }
func (t testVault) Path() (string, error)            { return t.path, nil }

func writeNote(t *testing.T, dir, name, content string) string {
	t.Helper()
	p := filepath.Join(dir, name)
	if err := os.MkdirAll(filepath.Dir(p), 0755); err != nil {
		t.Fatalf("failed to mkdir: %v", err)
	}
	if err := os.WriteFile(p, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write file: %v", err)
	}
	return p
}

func readFile(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("failed to read file: %v", err)
	}
	return string(b)
}

func TestGetFrontmatterValue(t *testing.T) {
	tmp := t.TempDir()
	v := testVault{path: tmp}

	content := "---\n" +
		"title: Test\n" +
		"tags:\n  - a\n  - b\n" +
		"count: 3\n" +
		"---\n" +
		"Body text here\n"
	writeNote(t, tmp, "note.md", content)

	// existing key (string)
	res, err := GetFrontmatterValue(v, "note", "title")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Found || res.Value != "Test" {
		t.Fatalf("expected title=Test found=true, got: %+v", res)
	}

	// existing key (sequence)
	res, err = GetFrontmatterValue(v, "note", "tags")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !res.Found {
		t.Fatalf("expected tags found=true")
	}
	// YAML unmarshals sequences as []interface{} with string elements
	arr, ok := res.Value.([]interface{})
	if !ok || len(arr) != 2 || arr[0] != "a" || arr[1] != "b" {
		t.Fatalf("expected tags [a b], got: %#v", res.Value)
	}

	// missing key
	res, err = GetFrontmatterValue(v, "note", "missing")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Found || res.Value != nil {
		t.Fatalf("expected missing key Found=false, Value=nil, got: %+v", res)
	}

	// note without frontmatter -> Found=false, Value=nil
	writeNote(t, tmp, "plain.md", "Just body\n")
	res, err = GetFrontmatterValue(v, "plain", "any")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if res.Found || res.Value != nil {
		t.Fatalf("expected no FM, got: %+v", res)
	}

	// validation errors
	if _, err := GetFrontmatterValue(v, "", "k"); err == nil {
		t.Fatalf("expected error for empty note name")
	}
	if _, err := GetFrontmatterValue(v, "n", ""); err == nil {
		t.Fatalf("expected error for empty key")
	}
	// not found
	if _, err := GetFrontmatterValue(v, "does-not-exist", "k"); err == nil {
		t.Fatalf("expected not found error")
	}
}

func TestEditFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	v := testVault{path: tmp}

	// 1) Add to file without frontmatter
	path := writeNote(t, tmp, "n1.md", "Body line\n")
	if err := EditFrontmatter(v, FrontmatterEditParams{NoteName: "n1", Key: "title", Value: "Hello"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	updated := readFile(t, path)
	fm, body, had, err := parseFrontmatter(updated)
	if err != nil || !had {
		t.Fatalf("expected frontmatter to be present, err=%v had=%v", err, had)
	}
	if val, ok := fm["title"]; !ok || val != "Hello" {
		t.Fatalf("expected title=Hello, got: %#v", fm["title"])
	}
	// body should be preserved (with a single blank line inserted by builder)
	if body != "\nBody line\n" {
		t.Fatalf("unexpected body: %q", body)
	}

	// 2) Update tags using comma-separated string -> array
	path = writeNote(t, tmp, "n2.md", "---\nother: x\n---\ncontent\n")
	if err := EditFrontmatter(v, FrontmatterEditParams{NoteName: "n2", Key: "tags", Value: "a, b, c"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	updated = readFile(t, path)
	fm, _, _, err = parseFrontmatter(updated)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	tags, ok := fm["tags"].([]interface{})
	if !ok || len(tags) != 3 || tags[0] != "a" || tags[1] != "b" || tags[2] != "c" {
		t.Fatalf("expected tags [a b c], got: %#v", fm["tags"])
	}
	if fm["other"] != "x" {
		t.Fatalf("expected other=x to be preserved")
	}

	// 3) Empty value behavior: tags -> [] ; other -> ""
	path = writeNote(t, tmp, "n3.md", "---\na: 1\n---\n")
	if err := EditFrontmatter(v, FrontmatterEditParams{NoteName: "n3", Key: "tags", Value: ""}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	updated = readFile(t, path)
	fm, _, _, _ = parseFrontmatter(updated)
	if arr, ok := fm["tags"].([]interface{}); !ok || len(arr) != 0 {
		t.Fatalf("expected empty tags list, got: %#v", fm["tags"])
	}
	if err := EditFrontmatter(v, FrontmatterEditParams{NoteName: "n3", Key: "summary", Value: ""}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	updated = readFile(t, path)
	fm, _, _, _ = parseFrontmatter(updated)
	if val, ok := fm["summary"]; !ok || val != "" {
		t.Fatalf("expected summary to be empty string, got: %#v", val)
	}
}

func TestClearFrontmatter(t *testing.T) {
	tmp := t.TempDir()
	v := testVault{path: tmp}

	// Prepare a note with various types
	note := "---\nlist:\n  - a\n  - b\nmap:\n  k: v\nstr: hello\n---\nB\n"
	path := writeNote(t, tmp, "c1.md", note)

	// Clear list -> []
	if err := ClearFrontmatter(v, "c1", "list"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fm, _, _, err := parseFrontmatter(readFile(t, path))
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if arr, ok := fm["list"].([]interface{}); !ok || len(arr) != 0 {
		t.Fatalf("expected cleared list [], got: %#v", fm["list"])
	}

	// Clear map -> {}
	if err := ClearFrontmatter(v, "c1", "map"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fm, _, _, _ = parseFrontmatter(readFile(t, path))
	if m, ok := fm["map"].(map[string]interface{}); !ok || len(m) != 0 {
		t.Fatalf("expected cleared map {}, got: %#v", fm["map"])
	}

	// Clear string -> ""
	if err := ClearFrontmatter(v, "c1", "str"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fm, _, _, _ = parseFrontmatter(readFile(t, path))
	if val, ok := fm["str"]; !ok || val != "" {
		t.Fatalf("expected cleared string \"\", got: %#v", fm["str"])
	}

	// Non-existent key -> no change
	before := readFile(t, path)
	if err := ClearFrontmatter(v, "c1", "missing"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	after := readFile(t, path)
	if before != after {
		t.Fatalf("expected no changes when clearing missing key")
	}

	// No frontmatter -> noop
	writeNote(t, tmp, "c2.md", "body only\n")
	if err := ClearFrontmatter(v, "c2", "any"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestRemoveFrontmatterKey(t *testing.T) {
	tmp := t.TempDir()
	v := testVault{path: tmp}

	// Remove one key, keep others
	path := writeNote(t, tmp, "r1.md", "---\na: 1\nb: 2\n---\nBODY\n")
	if err := RemoveFrontmatterKey(v, "r1", "a"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fm, body, had, err := parseFrontmatter(readFile(t, path))
	if err != nil || !had {
		t.Fatalf("parse error or no frontmatter: %v had=%v", err, had)
	}
	if _, exists := fm["a"]; exists {
		t.Fatalf("expected key 'a' removed")
	}
	if fm["b"] != 2 {
		t.Fatalf("expected key 'b' remains with value 2, got: %#v", fm["b"])
	}
	if body != "\nBODY\n" {
		t.Fatalf("unexpected body: %q", body)
	}

	// Remove last key -> entire FM removed and no leading blank line before body
	path = writeNote(t, tmp, "r2.md", "---\na: z\n---\nContent\n")
	if err := RemoveFrontmatterKey(v, "r2", "a"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	updated := readFile(t, path)
	// After removal, there should be no frontmatter; parse should report had=false and fm=nil
	fm2, body2, had2, err := parseFrontmatter(updated)
	if err != nil {
		t.Fatalf("parse error: %v", err)
	}
	if had2 || fm2 != nil {
		t.Fatalf("expected no frontmatter, got had=%v fm=%#v", had2, fm2)
	}
	// Body should not start with an extra blank line
	if body2 != "Content\n" {
		t.Fatalf("unexpected body after removing last key: %q", body2)
	}

	// Removing non-existent key -> no changes
	path = writeNote(t, tmp, "r3.md", "---\nx: 1\n---\nB\n")
	before := readFile(t, path)
	if err := RemoveFrontmatterKey(v, "r3", "nope"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	after := readFile(t, path)
	if before != after {
		t.Fatalf("expected file unchanged when removing missing key")
	}
}

// Sanity test for parseFrontmatter with CRLF input
func TestParseFrontmatter_CRLF(t *testing.T) {
	input := "---\r\nkey: val\r\n---\r\nBody\r\n"
	fm, body, had, err := parseFrontmatter(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !had || fm["key"] != "val" {
		t.Fatalf("expected fm with key=val, got had=%v fm=%#v", had, fm)
	}
	// Body normalization returns body using \n; we just ensure content sans frontmatter is preserved.
	if body != "Body\n" && body != "\nBody\n" {
		t.Fatalf("unexpected body: %q", body)
	}
}

// Additional safety: ensure YAML round-trip types are as expected for numbers
func TestYamlNumberTypePreservedOnRemove(t *testing.T) {
	tmp := t.TempDir()
	v := testVault{path: tmp}

	path := writeNote(t, tmp, "num.md", "---\nnum: 3\n---\nB\n")
	// Edit another key so FM marshals; then remove it and ensure num remains 3 (as int)
	if err := EditFrontmatter(v, FrontmatterEditParams{NoteName: "num", Key: "x", Value: "1"}); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := RemoveFrontmatterKey(v, "num", "x"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	fm, _, _, _ := parseFrontmatter(readFile(t, path))
	if v, ok := fm["num"]; !ok || reflect.TypeOf(v).Kind() != reflect.Int {
		t.Fatalf("expected num to remain integer, got: %#v (%T)", v, v)
	}
}
