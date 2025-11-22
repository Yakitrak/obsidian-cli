package obsidian

import (
	"regexp"
	"strings"
)

// RewriteLinksInContent updates internal wikilinks and markdown links targeting oldPath to point to newPath.
// It preserves alias text and fragments (headings/block refs) and returns the rewritten content plus count of replacements.
func RewriteLinksInContent(content, oldPath, newPath string) (string, int) {
	oldNorm := NormalizePath(AddMdSuffix(oldPath))
	oldBase := strings.TrimSuffix(oldNorm, ".md")
	newNorm := NormalizePath(AddMdSuffix(newPath))

	// Wikilinks and embeds: ![[...]] or [[...]]
	wikiPattern := regexp.MustCompile(`(!)?\[\[(.+?)\]\]`)
	rewriteCount := 0

	content = wikiPattern.ReplaceAllStringFunc(content, func(match string) string {
		m := wikiPattern.FindStringSubmatch(match)
		if len(m) < 3 {
			return match
		}
		isEmbed := m[1] == "!"
		inner := m[2]

		targetPart := inner
		aliasPart := ""
		if pipeIdx := strings.Index(inner, "|"); pipeIdx != -1 {
			targetPart = inner[:pipeIdx]
			aliasPart = inner[pipeIdx+1:]
		}

		fragment := ""
		base := targetPart
		if hashIdx := strings.Index(base, "#"); hashIdx != -1 {
			fragment = base[hashIdx:]
			base = base[:hashIdx]
		}

		hadExt := strings.HasSuffix(base, ".md")
		targetNorm := NormalizePath(AddMdSuffix(base))
		targetBase := strings.TrimSuffix(targetNorm, ".md")
		if targetNorm != oldNorm && targetBase != oldBase {
			return match
		}

		newBase := newNorm
		if !hadExt {
			newBase = strings.TrimSuffix(newBase, ".md")
		}
		newTarget := newBase + fragment
		if aliasPart != "" {
			newTarget += "|" + aliasPart
		}
		rewriteCount++
		prefix := ""
		if isEmbed {
			prefix = "!"
		}
		return prefix + "[[" + newTarget + "]]"
	})

	// Markdown links and embeds: [text](...) or ![text](...)
	mdPattern := regexp.MustCompile(`(!)?\[([^\]]*)\]\(([^)]+)\)`)
	content = mdPattern.ReplaceAllStringFunc(content, func(match string) string {
		m := mdPattern.FindStringSubmatch(match)
		if len(m) < 4 {
			return match
		}
		isEmbed := m[1] == "!"
		text := m[2]
		href := strings.TrimSpace(m[3])

		// Ignore external links
		if strings.HasPrefix(href, "http://") || strings.HasPrefix(href, "https://") {
			return match
		}

		fragment := ""
		base := href
		if hashIdx := strings.Index(base, "#"); hashIdx != -1 {
			fragment = base[hashIdx:]
			base = base[:hashIdx]
		}

		hadExt := strings.HasSuffix(base, ".md")
		targetNorm := NormalizePath(AddMdSuffix(base))
		targetBase := strings.TrimSuffix(targetNorm, ".md")
		if targetNorm != oldNorm && targetBase != oldBase {
			return match
		}

		newBase := newNorm
		if !hadExt {
			newBase = strings.TrimSuffix(newBase, ".md")
		}
		newHref := newBase + fragment

		rewriteCount++
		prefix := ""
		if isEmbed {
			prefix = "!"
		}
		return prefix + "[" + text + "](" + newHref + ")"
	})

	return content, rewriteCount
}
