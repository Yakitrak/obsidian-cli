package obsidian

import (
	"regexp"
	"strings"
	"time"
)

// ExpandTemplateVariables expands common Obsidian template variables in content.
// Supported:
//   - {{date}} / {{date:FORMAT}}
//   - {{time}} / {{time:FORMAT}}
//   - {{title}}
//
// FORMAT uses the same curated Obsidian-style tokens as ExpandDatePattern (including [literal] blocks).
func ExpandTemplateVariables(content []byte, title string) []byte {
	return ExpandTemplateVariablesAt(content, title, time.Now())
}

// ExpandTemplateVariablesAt is like ExpandTemplateVariables but uses the provided time.
func ExpandTemplateVariablesAt(content []byte, title string, now time.Time) []byte {
	result := string(content)
	result = strings.ReplaceAll(result, "{{title}}", RemoveMdSuffix(title))

	result = expandTemplateDate(result, now)
	result = expandTemplateTime(result, now)

	return []byte(result)
}

func expandTemplateDate(content string, now time.Time) string {
	// {{date:FORMAT}}
	re := regexp.MustCompile(`\{\{date:([^}]+)\}\}`)
	content = re.ReplaceAllStringFunc(content, func(match string) string {
		format := match[7 : len(match)-2]
		s, err := FormatDatePattern(format, now)
		if err != nil {
			return now.Format("2006-01-02")
		}
		return s
	})

	// {{date}}
	return strings.ReplaceAll(content, "{{date}}", now.Format("2006-01-02"))
}

func expandTemplateTime(content string, now time.Time) string {
	// {{time:FORMAT}}
	re := regexp.MustCompile(`\{\{time:([^}]+)\}\}`)
	content = re.ReplaceAllStringFunc(content, func(match string) string {
		format := match[7 : len(match)-2]
		s, err := FormatDatePattern(format, now)
		if err != nil {
			return now.Format("15:04")
		}
		return s
	})

	// {{time}}
	return strings.ReplaceAll(content, "{{time}}", now.Format("15:04"))
}
