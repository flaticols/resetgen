package main

import "testing"

func TestIsValidGoIdentifier(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		// Valid identifiers
		{"single uppercase letter", "U", true},
		{"simple struct name", "User", true},
		{"with underscore", "User_Data", true},
		{"with number", "User123", true},
		{"CamelCase", "UserConfig", true},
		{"with multiple underscores", "User_Config_Data", true},

		// Invalid identifiers
		{"empty string", "", false},
		{"lowercase start", "user", false},
		{"starts with number", "123User", false},
		{"starts with underscore", "_User", false},
		{"contains space", "User Type", false},
		{"contains hyphen", "User-Type", false},
		{"contains dot", "User.Type", false},
		{"lowercase only", "config", false},
		{"number only", "123", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidGoIdentifier(tt.input)
			if got != tt.valid {
				t.Errorf("isValidGoIdentifier(%q) = %v, want %v", tt.input, got, tt.valid)
			}
		})
	}
}
