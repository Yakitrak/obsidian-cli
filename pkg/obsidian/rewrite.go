package obsidian

import (
	"regexp"
	"runtime"
	"strings"
)

// IsCaseInsensitiveFS returns true if the current OS typically uses a case-insensitive filesystem.
// macOS and Windows are case-insensitive by default; Linux is case-sensitive.
func IsCaseInsensitiveFS() bool {
	return runtime.GOOS == "darwin" || runtime.GOOS == "windows"
}

// pathsEqual compares two paths, using case-insensitive comparison on macOS/Windows.
func pathsEqual(a, b string) bool {
	if IsCaseInsensitiveFS() {
		return strings.EqualFold(a, b)
	}
	return a == b
}

// NormalizeForComparison returns a string suitable for use as a map key when
// detecting duplicates. On case-insensitive filesystems (macOS/Windows), this
// lowercases the string; on Linux it returns it unchanged.
func NormalizeForComparison(s string) string {
	if IsCaseInsensitiveFS() {
		return strings.ToLower(s)
	}
	return s
}

// RewriteLinksInContent updates internal wikilinks and markdown links targeting oldPath to point to newPath.
// It preserves alias text and fragments (headings/block refs) and returns the rewritten content plus count of replacements.
// If basenameUnique is true, links matching just the basename (without folder) will also be rewritten.
// Set basenameUnique to false if there are other notes with the same basename in different folders.
func RewriteLinksInContent(content, oldPath, newPath string) (string, int) {
	return RewriteLinksInContentWithOptions(content, oldPath, newPath, true)
}

// RewriteLinksInContentWithOptions is like RewriteLinksInContent but allows controlling basename matching.
func RewriteLinksInContentWithOptions(content, oldPath, newPath string, basenameUnique bool) (string, int) {
	oldNorm := NormalizePath(AddMdSuffix(oldPath))
	oldBase := strings.TrimSuffix(oldNorm, ".md")
	// Also match against just the filename (no folder), since Obsidian wikilinks often omit the path
	oldBasename := baseName(oldBase)
	newNorm := NormalizePath(AddMdSuffix(newPath))
	newBasename := baseName(strings.TrimSuffix(newNorm, ".md"))

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
		// Match against full path OR just basename (Obsidian links often omit folder)
		// Only match by basename if basenameUnique is true (no other notes share this basename)
		// Use OS-appropriate case sensitivity (case-insensitive on macOS/Windows)
		matchedByBasename := basenameUnique && pathsEqual(targetBase, oldBasename) && !pathsEqual(targetBase, oldBase)
		matchedByFullPath := pathsEqual(targetNorm, oldNorm) || pathsEqual(targetBase, oldBase)
		if !matchedByFullPath && !matchedByBasename {
			return match
		}

		// If matched by basename only, use the new basename (preserve link style without folder)
		var newBase string
		if matchedByBasename {
			newBase = newBasename
			if hadExt {
				newBase += ".md"
			}
		} else {
			newBase = newNorm
			if !hadExt {
				newBase = strings.TrimSuffix(newBase, ".md")
			}
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
		// Match against full path OR just basename (Obsidian links often omit folder)
		// Only match by basename if basenameUnique is true (no other notes share this basename)
		// Use OS-appropriate case sensitivity (case-insensitive on macOS/Windows)
		matchedByBasename := basenameUnique && pathsEqual(targetBase, oldBasename) && !pathsEqual(targetBase, oldBase)
		matchedByFullPath := pathsEqual(targetNorm, oldNorm) || pathsEqual(targetBase, oldBase)
		if !matchedByFullPath && !matchedByBasename {
			return match
		}

		// If matched by basename only, use the new basename (preserve link style without folder)
		var newBase string
		if matchedByBasename {
			newBase = newBasename
			if hadExt {
				newBase += ".md"
			}
		} else {
			newBase = newNorm
			if !hadExt {
				newBase = strings.TrimSuffix(newBase, ".md")
			}
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

// baseName returns the last element of a path (the filename without directory).
func baseName(path string) string {
	// Handle both forward and back slashes
	path = strings.ReplaceAll(path, "\\", "/")
	if idx := strings.LastIndex(path, "/"); idx >= 0 {
		return path[idx+1:]
	}
	return path
}
