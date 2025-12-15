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

func TestIsValidPackagePath(t *testing.T) {
	tests := []struct {
		name  string
		input string
		valid bool
	}{
		// Valid package paths
		{"simple lowercase", "models", true},
		{"with underscore", "api_v2", true},
		{"with dot", "github.com/user/pkg", true},
		{"nested path", "internal/api", true},
		{"complex path", "github.com/flaticols/resetgen", true},
		{"with numbers", "v2", true},

		// Invalid package paths
		{"empty string", "", false},
		{"uppercase only", "MODELS", false},
		{"starts with uppercase", "Models", false},
		{"mixed case not allowed", "myPackage", false},
		{"contains space", "api models", false},
		{"contains hyphen", "api-v2", false},
		{"starts with dot", ".models", false},
		{"contains special chars", "api@models", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isValidPackagePath(tt.input)
			if got != tt.valid {
				t.Errorf("isValidPackagePath(%q) = %v, want %v", tt.input, got, tt.valid)
			}
		})
	}
}

func TestShouldProcessStruct(t *testing.T) {
	tests := []struct {
		name       string
		structName string
		pkgName    string
		filter     map[string]bool
		want       bool
	}{
		// No filter - process all
		{"no filter", "User", "models", nil, true},

		// Simple name matches
		{"simple name match", "User", "models", map[string]bool{"User": true}, true},
		{"simple name no match", "User", "models", map[string]bool{"Order": true}, false},

		// Package-qualified matches
		{"qualified match exact", "User", "models", map[string]bool{"models.User": true}, true},
		{"qualified no match different pkg", "User", "api", map[string]bool{"models.User": true}, false},

		// Mixed filters
		{"simple name takes precedence", "User", "api", map[string]bool{"User": true, "models.User": true}, true},
		{"qualified matches in correct pkg", "User", "api", map[string]bool{"api.User": true}, true},

		// Empty filter
		{"empty filter", "User", "models", map[string]bool{}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldProcessStruct(tt.structName, tt.pkgName, tt.filter)
			if got != tt.want {
				t.Errorf("shouldProcessStruct(%q, %q, %v) = %v, want %v",
					tt.structName, tt.pkgName, tt.filter, got, tt.want)
			}
		})
	}
}
