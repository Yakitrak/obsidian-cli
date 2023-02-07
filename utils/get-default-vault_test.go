package utils

import "testing"

func TestGetDefaultVault(t *testing.T) {
	// Set up default vault
	SetDefaultVault("new default value")

	var tests = []struct {
		testName string
		want     string
	}{
		{"Sets default vault", "new default value"},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := GetDefaultVault()
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}
