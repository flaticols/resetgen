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

// Directive tests

func TestParseSource_DirectiveOnly(t *testing.T) {
	src := `package test

// +resetgen
type User struct {
	ID    int64
	Name  string
	Email string
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
	if len(s.Fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(s.Fields))
	}

	// All fields should have ActionZero
	for i, f := range s.Fields {
		if f.Action != types.ActionZero {
			t.Errorf("field %d: expected ActionZero, got %d", i, f.Action)
		}
	}
}

func TestParseSource_DirectiveWithTags(t *testing.T) {
	src := `package test

//+resetgen
type User struct {
	ID    int64
	Name  string  ` + "`reset:\"guest\"`" + `
	Email string
	Age   int     ` + "`reset:\"-\"`" + `
}
`
	info, err := ParseSource(src)
	if err != nil {
		t.Fatalf("ParseSource failed: %v", err)
	}

	s := info.Structs[0]
	if len(s.Fields) != 4 {
		t.Fatalf("expected 4 fields, got %d", len(s.Fields))
	}

	tests := []struct {
		name   string
		action types.TagAction
		def    string
	}{
		{"ID", types.ActionZero, ""},
		{"Name", types.ActionDefault, "guest"},
		{"Email", types.ActionZero, ""},
		{"Age", types.ActionIgnore, ""},
	}

	for i, tt := range tests {
		f := s.Fields[i]
		if f.Name != tt.name {
			t.Errorf("field %d: expected name %s, got %s", i, tt.name, f.Name)
		}
		if f.Action != tt.action {
			t.Errorf("field %d: expected action %d, got %d", i, tt.action, f.Action)
		}
		if f.Default != tt.def {
			t.Errorf("field %d: expected default %q, got %q", i, tt.def, f.Default)
		}
	}
}

func TestParseSource_DirectiveRespectsIgnore(t *testing.T) {
	src := `package test

// +resetgen
type Config struct {
	Host   string
	Port   int
	Secret string ` + "`reset:\"-\"`" + `
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

	// Host and Port should be ActionZero
	if s.Fields[0].Action != types.ActionZero {
		t.Errorf("Host: expected ActionZero, got %d", s.Fields[0].Action)
	}
	if s.Fields[1].Action != types.ActionZero {
		t.Errorf("Port: expected ActionZero, got %d", s.Fields[1].Action)
	}
	// Secret should be ActionIgnore
	if s.Fields[2].Action != types.ActionIgnore {
		t.Errorf("Secret: expected ActionIgnore, got %d", s.Fields[2].Action)
	}
}

func TestParseSource_DirectiveSkipsUnexported(t *testing.T) {
	src := `package test

// +resetgen
type Request struct {
	ID    string
	name  string
	Token string
}
`
	info, err := ParseSource(src)
	if err != nil {
		t.Fatalf("ParseSource failed: %v", err)
	}

	s := info.Structs[0]
	if len(s.Fields) != 2 {
		t.Fatalf("expected 2 fields (unexported 'name' skipped), got %d", len(s.Fields))
	}

	// Should have ID and Token, but not name
	if s.Fields[0].Name != "ID" {
		t.Errorf("expected first field ID, got %s", s.Fields[0].Name)
	}
	if s.Fields[1].Name != "Token" {
		t.Errorf("expected second field Token, got %s", s.Fields[1].Name)
	}
}

func TestParseSource_DirectiveAllIgnored(t *testing.T) {
	src := `package test

// +resetgen
type Config struct {
	Field1 int ` + "`reset:\"-\"`" + `
	Field2 int ` + "`reset:\"-\"`" + `
}
`
	info, err := ParseSource(src)
	if err != nil {
		t.Fatalf("ParseSource failed: %v", err)
	}

	// Struct should be skipped entirely (no fields to reset)
	if len(info.Structs) != 0 {
		t.Fatalf("expected 0 structs (all fields ignored), got %d", len(info.Structs))
	}
}

func TestParseSource_DirectiveFormats(t *testing.T) {
	tests := []struct {
		name     string
		src      string
		expected int
	}{
		{
			"no space", `package test
// +resetgen
type User struct {
	ID int
}`,
			1,
		},
		{
			"single space", `package test
// +resetgen
type User struct {
	ID int
}`,
			1,
		},
		{
			"multiple spaces", `package test
//  +resetgen
type User struct {
	ID int
}`,
			1,
		},
		{
			"no space after slash", `package test
//+resetgen
type User struct {
	ID int
}`,
			1,
		},
		{
			"wrong prefix no plus", `package test
// resetgen
type User struct {
	ID int
}`,
			0,
		},
		{
			"case sensitive", `package test
// +ResetGen
type User struct {
	ID int
}`,
			0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := ParseSource(tt.src)
			if err != nil {
				t.Fatalf("ParseSource failed: %v", err)
			}

			if len(info.Structs) != tt.expected {
				t.Errorf("expected %d structs, got %d", tt.expected, len(info.Structs))
			}
		})
	}
}

func TestParseSource_DirectiveEmbedded(t *testing.T) {
	src := `package test

// +resetgen
type Request struct {
	Body io.Reader
	Name string
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

	// Both fields should have ActionZero
	if s.Fields[0].Action != types.ActionZero {
		t.Errorf("Body: expected ActionZero, got %d", s.Fields[0].Action)
	}
	if s.Fields[1].Action != types.ActionZero {
		t.Errorf("Name: expected ActionZero, got %d", s.Fields[1].Action)
	}
}

func TestParseSource_BackwardCompatibility(t *testing.T) {
	// Verify that tag-based detection still works without directive
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

	// Should only have SomeTags (backward compatibility)
	if len(info.Structs) != 1 {
		t.Fatalf("expected 1 struct, got %d", len(info.Structs))
	}

	if info.Structs[0].Name != "SomeTags" {
		t.Errorf("expected SomeTags, got %s", info.Structs[0].Name)
	}
}

// Tests for -structs filter functionality

func TestParseSourceWithFilter_SpecificStructs(t *testing.T) {
	src := `package test

type User struct {
	ID    int64
	Name  string
}

type Config struct {
	Host string
	Port int
}

type Logger struct {
	Level string
}
`
	// Only process User and Config
	filter := map[string]bool{
		"User":   true,
		"Config": true,
	}

	info, err := ParseSourceWithFilter(src, filter)
	if err != nil {
		t.Fatalf("ParseSourceWithFilter failed: %v", err)
	}

	if len(info.Structs) != 2 {
		t.Fatalf("expected 2 structs, got %d", len(info.Structs))
	}

	// Should have User and Config, not Logger
	names := make(map[string]bool)
	for _, s := range info.Structs {
		names[s.Name] = true
	}

	if !names["User"] {
		t.Error("expected User struct")
	}
	if !names["Config"] {
		t.Error("expected Config struct")
	}
	if names["Logger"] {
		t.Error("Logger should not be included")
	}
}

func TestParseSourceWithFilter_AllExportedFields(t *testing.T) {
	src := `package test

type User struct {
	ID      int64
	Name    string
	email   string
	Age     int ` + "`reset:\"-\"`" + `
}
`
	filter := map[string]bool{"User": true}

	info, err := ParseSourceWithFilter(src, filter)
	if err != nil {
		t.Fatalf("ParseSourceWithFilter failed: %v", err)
	}

	s := info.Structs[0]

	// Should have ID, Name, and Age (but not email)
	if len(s.Fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(s.Fields))
	}

	// Check field names
	hasID := false
	hasName := false
	hasAge := false
	for _, f := range s.Fields {
		if f.Name == "ID" {
			hasID = true
		}
		if f.Name == "Name" {
			hasName = true
		}
		if f.Name == "Age" {
			hasAge = true
		}
	}

	if !hasID || !hasName || !hasAge {
		t.Errorf("missing expected fields: ID=%v, Name=%v, Age=%v", hasID, hasName, hasAge)
	}
}

func TestParseSourceWithFilter_RespectsTagsInFilteredStructs(t *testing.T) {
	src := `package test

type User struct {
	ID      int64           ` + "`reset:\"\"`" + `
	Name    string          ` + "`reset:\"guest\"`" + `
	Email   string
	Secret  string          ` + "`reset:\"-\"`" + `
}
`
	filter := map[string]bool{"User": true}

	info, err := ParseSourceWithFilter(src, filter)
	if err != nil {
		t.Fatalf("ParseSourceWithFilter failed: %v", err)
	}

	s := info.Structs[0]

	tests := []struct {
		name   string
		action types.TagAction
		def    string
	}{
		{"ID", types.ActionZero, ""},
		{"Name", types.ActionDefault, "guest"},
		{"Email", types.ActionZero, ""},
		{"Secret", types.ActionIgnore, ""},
	}

	for i, tt := range tests {
		f := s.Fields[i]
		if f.Name != tt.name {
			t.Errorf("field %d: expected name %s, got %s", i, tt.name, f.Name)
		}
		if f.Action != tt.action {
			t.Errorf("field %d (%s): expected action %d, got %d", i, tt.name, tt.action, f.Action)
		}
		if f.Default != tt.def {
			t.Errorf("field %d (%s): expected default %q, got %q", i, tt.name, tt.def, f.Default)
		}
	}
}

func TestParseSourceWithFilter_EmptyFilter(t *testing.T) {
	src := `package test

type User struct {
	ID   int64
	Name string
}
`
	filter := map[string]bool{}

	info, err := ParseSourceWithFilter(src, filter)
	if err != nil {
		t.Fatalf("ParseSourceWithFilter failed: %v", err)
	}

	// Empty filter means process nothing
	if len(info.Structs) != 0 {
		t.Fatalf("expected 0 structs with empty filter, got %d", len(info.Structs))
	}
}

func TestParseSourceWithFilter_NilFilterUsesDefaultBehavior(t *testing.T) {
	src := `package test

type Tagged struct {
	ID int64 ` + "`reset:\"\"`" + `
}

type NotTagged struct {
	ID int64
}
`
	info, err := ParseSourceWithFilter(src, nil)
	if err != nil {
		t.Fatalf("ParseSourceWithFilter failed: %v", err)
	}

	// Nil filter should use default behavior (only Tagged)
	if len(info.Structs) != 1 {
		t.Fatalf("expected 1 struct with nil filter, got %d", len(info.Structs))
	}

	if info.Structs[0].Name != "Tagged" {
		t.Errorf("expected Tagged struct, got %s", info.Structs[0].Name)
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
