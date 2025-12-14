package basic

import "testing"

func TestUserReset(t *testing.T) {
	u := &User{
		ID:    123,
		Name:  "alice",
		Email: "alice@example.com",
		Tags:  []string{"admin", "user"},
	}

	u.Reset()

	if u.ID != 0 {
		t.Errorf("ID: expected 0, got %d", u.ID)
	}
	if u.Name != "guest" {
		t.Errorf("Name: expected guest, got %s", u.Name)
	}
	if u.Email != "" {
		t.Errorf("Email: expected empty, got %s", u.Email)
	}
	if len(u.Tags) != 0 {
		t.Errorf("Tags: expected empty, got %v", u.Tags)
	}
	if cap(u.Tags) < 2 {
		t.Errorf("Tags: capacity should be preserved, got %d", cap(u.Tags))
	}
}

func TestConfigReset(t *testing.T) {
	c := &Config{
		Debug:    true,
		Timeout:  60,
		Settings: map[string]any{"key": "value"},
	}

	c.Reset()

	if c.Debug != false {
		t.Errorf("Debug: expected false, got %v", c.Debug)
	}
	if c.Timeout != 30 {
		t.Errorf("Timeout: expected 30, got %d", c.Timeout)
	}
	if len(c.Settings) != 0 {
		t.Errorf("Settings: expected empty, got %v", c.Settings)
	}
}

func BenchmarkUserReset(b *testing.B) {
	u := &User{
		ID:    123,
		Name:  "alice",
		Email: "alice@example.com",
		Tags:  make([]string, 0, 10),
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		u.Reset()
	}
}

func BenchmarkConfigReset(b *testing.B) {
	c := &Config{
		Debug:    true,
		Timeout:  60,
		Settings: make(map[string]any, 10),
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		c.Reset()
	}
}
