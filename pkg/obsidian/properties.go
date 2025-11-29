package obsidian

import (
	"regexp"
	"strconv"
	"strings"
	"time"
)

// PropertyValueInfo describes the shape and detected type of a frontmatter property value.
// Values contains flattened string representations for enum/sample analysis (scalars and list items only).
type PropertyValueInfo struct {
	Shape     string   // scalar, list, object, mixed, unknown
	ValueType string   // bool, int, float, date, datetime, url, wikilink, string, object, mixed, unknown
	Values    []string // flattened scalar/list values; maps are omitted
}

var (
	urlPattern      = regexp.MustCompile(`(?i)^https?://\S+$`)
	wikilinkPattern = regexp.MustCompile(`^!?\[\[[^\]]+\]\]$`)
	dateLayouts     = []string{
		"2006-01-02",
		"2006/01/02",
	}
	dateTimeLayouts = []string{
		time.RFC3339,
		"2006-01-02 15:04",
		"2006-01-02 15:04:05",
		"2006/01/02 15:04",
		"2006/01/02 15:04:05",
		"2006-01-02T15:04",
		"2006-01-02T15:04:05",
	}
)

// AnalyzePropertyValue inspects a frontmatter value and returns shape/type details plus flattened values.
// Intended for vault-wide property analysis (e.g., enum detection and pattern inference).
func AnalyzePropertyValue(v interface{}) PropertyValueInfo {
	switch val := v.(type) {
	case nil:
		return PropertyValueInfo{Shape: "unknown", ValueType: "unknown"}
	case map[string]interface{}:
		return PropertyValueInfo{Shape: "object", ValueType: "object"}
	case map[interface{}]interface{}:
		return PropertyValueInfo{Shape: "object", ValueType: "object"}
	case []interface{}:
		return analyzeList(val)
	case []string:
		items := make([]interface{}, len(val))
		for i, s := range val {
			items[i] = s
		}
		return analyzeList(items)
	case bool:
		return PropertyValueInfo{Shape: "scalar", ValueType: "bool", Values: []string{strconv.FormatBool(val)}}
	case int:
		return PropertyValueInfo{Shape: "scalar", ValueType: "int", Values: []string{strconv.Itoa(val)}}
	case int64:
		return PropertyValueInfo{Shape: "scalar", ValueType: "int", Values: []string{strconv.FormatInt(val, 10)}}
	case float64:
		return PropertyValueInfo{Shape: "scalar", ValueType: "float", Values: []string{trimTrailingZeros(val)}}
	case float32:
		return PropertyValueInfo{Shape: "scalar", ValueType: "float", Values: []string{trimTrailingZeros(float64(val))}}
	case string:
		return analyzeString(val)
	case time.Time:
		if hasTimeComponent(val) {
			return PropertyValueInfo{Shape: "scalar", ValueType: "datetime", Values: []string{val.Format(time.RFC3339)}}
		}
		return PropertyValueInfo{Shape: "scalar", ValueType: "date", Values: []string{val.Format("2006-01-02")}}
	default:
		return PropertyValueInfo{Shape: "unknown", ValueType: "unknown"}
	}
}

func analyzeList(list []interface{}) PropertyValueInfo {
	typeCounts := make(map[string]int)
	var values []string

	for _, item := range list {
		info := AnalyzePropertyValue(item)
		if info.Shape == "list" {
			// Nested lists add ambiguity
			typeCounts["mixed"]++
		} else {
			typeCounts[info.ValueType]++
		}
		values = append(values, info.Values...)
	}

	valueType := "mixed"
	if len(typeCounts) == 1 {
		for t := range typeCounts {
			valueType = t
		}
	}

	return PropertyValueInfo{
		Shape:     "list",
		ValueType: valueType,
		Values:    values,
	}
}

func analyzeString(raw string) PropertyValueInfo {
	value := strings.TrimSpace(raw)
	switch {
	case value == "":
		return PropertyValueInfo{Shape: "scalar", ValueType: "string"}
	case urlPattern.MatchString(value):
		return PropertyValueInfo{Shape: "scalar", ValueType: "url", Values: []string{value}}
	case wikilinkPattern.MatchString(value):
		return PropertyValueInfo{Shape: "scalar", ValueType: "wikilink", Values: []string{value}}
	case isDateTime(value):
		return PropertyValueInfo{Shape: "scalar", ValueType: "datetime", Values: []string{value}}
	case isDate(value):
		return PropertyValueInfo{Shape: "scalar", ValueType: "date", Values: []string{value}}
	default:
		return PropertyValueInfo{Shape: "scalar", ValueType: "string", Values: []string{value}}
	}
}

func isDateTime(s string) bool {
	for _, layout := range dateTimeLayouts {
		if _, err := time.Parse(layout, s); err == nil {
			return true
		}
	}
	return false
}

func isDate(s string) bool {
	for _, layout := range dateLayouts {
		if t, err := time.Parse(layout, s); err == nil && t.Hour() == 0 && t.Minute() == 0 && t.Second() == 0 {
			return true
		}
	}
	return false
}

func hasTimeComponent(t time.Time) bool {
	return t.Hour() != 0 || t.Minute() != 0 || t.Second() != 0 || t.Nanosecond() != 0
}

func trimTrailingZeros(f float64) string {
	s := strconv.FormatFloat(f, 'f', -1, 64)
	if strings.Contains(s, ".") {
		s = strings.TrimRight(strings.TrimRight(s, "0"), ".")
	}
	if s == "" {
		return "0"
	}
	return s
}
