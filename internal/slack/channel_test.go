package slack

import "testing"

func TestIsChannelID(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"public channel ID", "C01234ABCDE", true},
		{"private channel ID", "G01234ABCDE", true},
		{"DM channel ID", "D01234ABCDE", true},
		{"short ID", "C01234AB", false},
		{"lowercase", "c01234abcde", false},
		{"channel name starting with C", "ci-notifications", false},
		{"channel name", "general", false},
		{"with hash", "#general", false},
		{"empty", "", false},
		{"long valid ID", "C0123456789ABCDEF", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsChannelID(tt.input)
			if got != tt.want {
				t.Errorf("IsChannelID(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}
