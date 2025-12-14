package parser

import (
	"testing"

	"github.com/flaticols/resetgen/internal/types"
)

func TestParseSource_BasicTypes(t *testing.T) {
	src := `package test

type User struct {
	ID    int64   ` + "`reset:\"\"`" + `
	Name  string  ` + "`reset:\"guest\"`" + `
	Email string  ` + "`reset:\"\"`" + `
	Age   int     ` + "`reset:\"-\"`" + `
}
`
	info, err := ParseSource(src)
	if err != nil {
		t.Fatalf("ParseSource failed: %v", err)
	}

	if len(info.Structs) != 1 {
		t.Fatalf("expected 1 struct, got %d", len(info.Structs))
	}

	s := info.Structs[0]
	if s.Name != "User" {
		t.Errorf("expected struct name User, got %s", s.Name)
	}

	if len(s.Fields) != 4 {
		t.Fatalf("expected 4 fields, got %d", len(s.Fields))
	}

	tests := []struct {
		name   string
		typ    string
		action types.TagAction
		def    string
	}{
		{"ID", "int64", types.ActionZero, ""},
		{"Name", "string", types.ActionDefault, "guest"},
		{"Email", "string", types.ActionZero, ""},
		{"Age", "int", types.ActionIgnore, ""},
	}

	for i, tt := range tests {
		f := s.Fields[i]
		if f.Name != tt.name {
			t.Errorf("field %d: expected name %s, got %s", i, tt.name, f.Name)
		}
		if f.TypeStr != tt.typ {
			t.Errorf("field %d: expected type %s, got %s", i, tt.typ, f.TypeStr)
		}
		if f.Action != tt.action {
			t.Errorf("field %d: expected action %d, got %d", i, tt.action, f.Action)
		}
		if f.Default != tt.def {
			t.Errorf("field %d: expected default %q, got %q", i, tt.def, f.Default)
		}
	}
}

func TestParseSource_SliceAndMap(t *testing.T) {
	src := `package test

type Container struct {
	Items   []string          ` + "`reset:\"\"`" + `
	Mapping map[string]int    ` + "`reset:\"\"`" + `
	Nested  [][]byte          ` + "`reset:\"\"`" + `
}
`
	info, err := ParseSource(src)
	if err != nil {
		t.Fatalf("ParseSource failed: %v", err)
	}

	s := info.Structs[0]
	if len(s.Fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(s.Fields))
	}

	if s.Fields[0].Kind != types.KindSlice {
		t.Errorf("Items: expected KindSlice, got %d", s.Fields[0].Kind)
	}
	if s.Fields[1].Kind != types.KindMap {
		t.Errorf("Mapping: expected KindMap, got %d", s.Fields[1].Kind)
	}
	if s.Fields[2].Kind != types.KindSlice {
		t.Errorf("Nested: expected KindSlice, got %d", s.Fields[2].Kind)
	}
}

func TestParseSource_Embedded(t *testing.T) {
	src := `package test

type Inner struct {
	Value int ` + "`reset:\"\"`" + `
}

type Outer struct {
	Inner   ` + "`reset:\"\"`" + `
	Name string ` + "`reset:\"\"`" + `
}
`
	info, err := ParseSource(src)
	if err != nil {
		t.Fatalf("ParseSource failed: %v", err)
	}

	if len(info.Structs) != 2 {
		t.Fatalf("expected 2 structs, got %d", len(info.Structs))
	}

	outer := info.Structs[1]
	if outer.Name != "Outer" {
		t.Errorf("expected struct name Outer, got %s", outer.Name)
	}

	if len(outer.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(outer.Fields))
	}

	// Check embedded field
	embedded := outer.Fields[0]
	if !embedded.IsEmbedded {
		t.Error("expected Inner to be embedded")
	}
	if embedded.Name != "Inner" {
		t.Errorf("expected embedded name Inner, got %s", embedded.Name)
	}
}

func TestParseSource_Pointer(t *testing.T) {
	src := `package test

type Config struct {
	Value   *int    ` + "`reset:\"\"`" + `
	Name    *string ` + "`reset:\"\"`" + `
}
`
	info, err := ParseSource(src)
	if err != nil {
		t.Fatalf("ParseSource failed: %v", err)
	}

	s := info.Structs[0]
	for _, f := range s.Fields {
		if f.Kind != types.KindPointer {
			t.Errorf("%s: expected KindPointer, got %d", f.Name, f.Kind)
		}
	}
}

func TestParseSource_NoResetTag(t *testing.T) {
	src := `package test

type NoTags struct {
	ID   int
	Name string
}

type SomeTags struct {
	ID   int
	Name string ` + "`reset:\"\"`" + `
}
`
	info, err := ParseSource(src)
	if err != nil {
		t.Fatalf("ParseSource failed: %v", err)
	}

	// Should only have SomeTags
	if len(info.Structs) != 1 {
		t.Fatalf("expected 1 struct, got %d", len(info.Structs))
	}

	if info.Structs[0].Name != "SomeTags" {
		t.Errorf("expected SomeTags, got %s", info.Structs[0].Name)
	}
}

func TestParseSource_GenericTypes(t *testing.T) {
	src := `package test

type Container[T any] struct {
	Items []T ` + "`reset:\"\"`" + `
}
`
	info, err := ParseSource(src)
	if err != nil {
		t.Fatalf("ParseSource failed: %v", err)
	}

	if len(info.Structs) != 1 {
		t.Fatalf("expected 1 struct, got %d", len(info.Structs))
	}

	s := info.Structs[0]
	if s.Name != "Container" {
		t.Errorf("expected Container, got %s", s.Name)
	}
}

func TestParseSource_ExternalTypes(t *testing.T) {
	src := `package test

import "time"

type Event struct {
	Timestamp time.Time ` + "`reset:\"\"`" + `
	Duration  time.Duration ` + "`reset:\"\"`" + `
}
`
	info, err := ParseSource(src)
	if err != nil {
		t.Fatalf("ParseSource failed: %v", err)
	}

	s := info.Structs[0]
	if len(s.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(s.Fields))
	}

	if s.Fields[0].TypeStr != "time.Time" {
		t.Errorf("expected time.Time, got %s", s.Fields[0].TypeStr)
	}
}

func BenchmarkParseSource(b *testing.B) {
	src := `package test

type User struct {
	ID       int64    ` + "`reset:\"\"`" + `
	Name     string   ` + "`reset:\"guest\"`" + `
	Email    string   ` + "`reset:\"\"`" + `
	Tags     []string ` + "`reset:\"\"`" + `
	Settings map[string]any ` + "`reset:\"\"`" + `
	Active   bool     ` + "`reset:\"-\"`" + `
}

type Config struct {
	Debug    bool   ` + "`reset:\"\"`" + `
	Timeout  int    ` + "`reset:\"30\"`" + `
	MaxRetry int    ` + "`reset:\"3\"`" + `
}
`
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := ParseSource(src)
		if err != nil {
			b.Fatal(err)
		}
	}
}
