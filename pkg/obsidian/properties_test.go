package obsidian

import "testing"

func TestAnalyzePropertyValue(t *testing.T) {
	tests := []struct {
		name      string
		value     interface{}
		shape     string
		valueType string
		valuesLen int
	}{
		{name: "bool", value: true, shape: "scalar", valueType: "bool", valuesLen: 1},
		{name: "int", value: 42, shape: "scalar", valueType: "int", valuesLen: 1},
		{name: "float", value: 3.1400, shape: "scalar", valueType: "float", valuesLen: 1},
		{name: "url", value: "https://example.com", shape: "scalar", valueType: "url", valuesLen: 1},
		{name: "wikilink", value: "[[Note Name]]", shape: "scalar", valueType: "wikilink", valuesLen: 1},
		{name: "datetime", value: "2024-05-01T12:30:00", shape: "scalar", valueType: "datetime", valuesLen: 1},
		{name: "date", value: "2024-05-01", shape: "scalar", valueType: "date", valuesLen: 1},
		{name: "string default", value: "plain text", shape: "scalar", valueType: "string", valuesLen: 1},
		{name: "list homogeneous", value: []interface{}{"a", "b"}, shape: "list", valueType: "string", valuesLen: 2},
		{name: "list mixed", value: []interface{}{"a", 1}, shape: "list", valueType: "mixed", valuesLen: 2},
		{name: "object map", value: map[string]interface{}{"k": "v"}, shape: "object", valueType: "object", valuesLen: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info := AnalyzePropertyValue(tt.value)
			if info.Shape != tt.shape {
				t.Errorf("expected shape %s, got %s", tt.shape, info.Shape)
			}
			if info.ValueType != tt.valueType {
				t.Errorf("expected valueType %s, got %s", tt.valueType, info.ValueType)
			}
			if len(info.Values) != tt.valuesLen {
				t.Errorf("expected %d values, got %d", tt.valuesLen, len(info.Values))
			}
		})
	}
}
