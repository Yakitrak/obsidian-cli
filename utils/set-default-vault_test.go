package utils

import "testing"

func TestSetDefaultVault(t *testing.T) {
	// TODO get full coverage

	// Set up default vault
	SetDefaultVault("new default value 2")

	var tests = []struct {
		testName string
		want     string
	}{
		{"Sets default vault", "new default value 2"},
	}
	for _, tt := range tests {
		t.Run(tt.testName, func(t *testing.T) {
			got := GetDefaultVault("")
			if got != tt.want {
				t.Errorf("got %s, want %s", got, tt.want)
			}
		})
	}
}
