package embeddings

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

const (
	minChunkChars    = 400
	targetChunkChars = 1200
	maxChunkChars    = 1800
	overlapChars     = 200
)

var (
	headingRE        = regexp.MustCompile(`^(#{1,6})\s+(.*)$`)
	frontmatterRegex = regexp.MustCompile(`(?s)^\s*---\r?\n(.*?)\r?\n---\s*\r?\n`)
)

type section struct {
	level   int
	heading string
	lines   []string
}

// ChunkNote produces heading-aware chunks with breadcrumbs and optional frontmatter context.
func ChunkNote(path, title, content string) ([]ChunkInput, error) {
	frontmatter := extractFrontmatter(content)
	body := stripFrontmatter(content)

	sections := parseSections(body)
	if len(sections) == 0 {
		sections = []section{{level: 1, heading: title, lines: strings.Split(body, "\n")}}
	}
	sections = coalesceShortSections(sections)

	var chunks []ChunkInput
	breadcrumbStack := []string{title}

	for idx, sec := range sections {
		// maintain breadcrumb stack
		for len(breadcrumbStack) > 1 && len(breadcrumbStack)-1 >= sec.level {
			breadcrumbStack = breadcrumbStack[:len(breadcrumbStack)-1]
		}
		breadcrumbStack = append(breadcrumbStack, strings.TrimSpace(sec.heading))
		breadcrumb := strings.Join(breadcrumbStack, " > ")

		bodyText := strings.TrimSpace(strings.Join(sec.lines, "\n"))
		if bodyText == "" {
			continue
		}

		parts := splitWithOverlap(bodyText, maxChunkChars, overlapChars)
		if len(parts) == 0 {
			continue
		}

		for partIdx, part := range parts {
			text := buildChunkText(path, title, breadcrumb, frontmatter, part, len(parts), partIdx+1)
			chunks = append(chunks, ChunkInput{
				Index:      len(chunks),
				Text:       text,
				Breadcrumb: breadcrumb,
				Heading:    sec.heading,
				Hash:       hashText(text),
			})
		}
		_ = idx // reserved for future numbering
	}

	return chunks, nil
}

func stripFrontmatter(content string) string {
	loc := frontmatterRegex.FindStringIndex(content)
	if len(loc) == 2 && loc[0] == 0 {
		return content[loc[1]:]
	}
	return content
}

func parseSections(body string) []section {
	lines := strings.Split(body, "\n")
	var sections []section
	current := section{level: 1, heading: ""}
	inCode := false

	flush := func() {
		if len(current.lines) == 0 && strings.TrimSpace(current.heading) == "" {
			return
		}
		sections = append(sections, current)
	}

	for _, raw := range lines {
		line := strings.TrimRight(raw, "\r")
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "```") || strings.HasPrefix(trim, "~~~") {
			inCode = !inCode
		}
		if inCode {
			current.lines = append(current.lines, line)
			continue
		}
		if m := headingRE.FindStringSubmatch(trim); m != nil {
			flush()
			level := len(m[1])
			current = section{level: level, heading: strings.TrimSpace(m[2])}
			continue
		}
		current.lines = append(current.lines, line)
	}
	flush()
	return sections
}

func coalesceShortSections(sections []section) []section {
	var out []section
	for i := 0; i < len(sections); i++ {
		sec := sections[i]
		text := strings.Join(sec.lines, "\n")
		for len(text) < minChunkChars && i+1 < len(sections) {
			next := sections[i+1]
			sec.lines = append(sec.lines, "")
			sec.lines = append(sec.lines, next.lines...)
			text = strings.Join(sec.lines, "\n")
			i++
		}
		out = append(out, sec)
	}
	return out
}

func splitWithOverlap(text string, maxChars, overlap int) []string {
	if len(text) <= maxChars && len(text) >= minChunkChars {
		return []string{text}
	}
	if len(text) <= maxChars {
		return []string{text}
	}
	var parts []string
	runes := []rune(text)
	start := 0
	for start < len(runes) {
		end := start + maxChars
		if end > len(runes) {
			end = len(runes)
		}
		parts = append(parts, string(runes[start:end]))
		if end == len(runes) {
			break
		}
		start = end - overlap
		if start < 0 {
			start = 0
		}
	}
	return parts
}

func buildChunkText(path, title, breadcrumb string, frontmatter map[string]interface{}, body string, totalParts, partNumber int) string {
	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("Title: %s\n", title))
	sb.WriteString(fmt.Sprintf("Path: %s\n", path))
	sb.WriteString(fmt.Sprintf("Headings: %s\n", breadcrumb))
	if fm := summarizeFrontmatter(frontmatter); fm != "" {
		sb.WriteString(fmt.Sprintf("Frontmatter: %s\n", fm))
	}
	if totalParts > 1 {
		sb.WriteString(fmt.Sprintf("Chunk: %d/%d\n", partNumber, totalParts))
	}
	sb.WriteString("\n")
	sb.WriteString(body)
	return sb.String()
}

func summarizeFrontmatter(fm map[string]interface{}) string {
	if len(fm) == 0 {
		return ""
	}
	keys := make([]string, 0, len(fm))
	for k := range fm {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	var parts []string
	for _, k := range keys {
		v := fm[k]
		switch t := v.(type) {
		case string:
			if strings.TrimSpace(t) != "" {
				parts = append(parts, fmt.Sprintf("%s=%s", k, t))
			}
		case []string:
			if len(t) > 0 {
				parts = append(parts, fmt.Sprintf("%s=%s", k, strings.Join(t, ",")))
			}
		case []interface{}:
			var vals []string
			for _, item := range t {
				if s, ok := item.(string); ok {
					vals = append(vals, s)
				}
			}
			if len(vals) > 0 {
				parts = append(parts, fmt.Sprintf("%s=%s", k, strings.Join(vals, ",")))
			}
		default:
			parts = append(parts, fmt.Sprintf("%s=%v", k, t))
		}
		if len(parts) >= 6 { // keep summary short
			break
		}
	}
	return strings.Join(parts, "; ")
}

func hashText(text string) string {
	sum := sha256.Sum256([]byte(text))
	return hex.EncodeToString(sum[:])
}

func extractFrontmatter(content string) map[string]interface{} {
	loc := frontmatterRegex.FindStringSubmatchIndex(content)
	if len(loc) < 4 || loc[0] != 0 {
		return nil
	}
	// capture group 1 contains YAML
	yamlStart, yamlEnd := loc[2], loc[3]
	if yamlStart < 0 || yamlEnd < 0 || yamlEnd <= yamlStart {
		return nil
	}
	yamlContent := content[yamlStart:yamlEnd]
	var fm map[string]interface{}
	if err := yaml.Unmarshal([]byte(yamlContent), &fm); err != nil || fm == nil {
		return nil
	}
	return fm
}
