package utils

import "testing"

func TestUrlConstructor(t *testing.T) {

	var tests = []struct {
		name string
		in   map[string]string
		want string
	}{
		{"Empty map", map[string]string{}, ""},
		{"One key", map[string]string{"key": "value"}, "?key=value"},
		{"Two keys", map[string]string{"key1": "value1", "key2": "value2"}, "?key1=value1&key2=value2"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := UriConstructor(tt.in)
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}

}
