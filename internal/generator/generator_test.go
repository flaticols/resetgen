package generator

import (
	"strings"
	"testing"

	"github.com/flaticols/resetgen/internal/parser"
	"github.com/flaticols/resetgen/internal/types"
)

func TestGenerate_BasicTypes(t *testing.T) {
	info := &types.FileInfo{
		PkgName: "test",
		Structs: []types.StructInfo{
			{
				Name: "User",
				Fields: []types.FieldInfo{
					{Name: "ID", TypeStr: "int64", Kind: types.KindBasic, Action: types.ActionZero},
					{Name: "Name", TypeStr: "string", Kind: types.KindBasic, Action: types.ActionDefault, Default: "guest"},
					{Name: "Active", TypeStr: "bool", Kind: types.KindBasic, Action: types.ActionIgnore},
				},
			},
		},
	}

	code := Generate(info)

	if !strings.Contains(code, "func (s *User) Reset()") {
		t.Error("missing Reset() method signature")
	}
	if !strings.Contains(code, "s.ID = 0") {
		t.Error("missing ID zero value")
	}
	if !strings.Contains(code, `s.Name = "guest"`) {
		t.Error("missing Name default value")
	}
	if strings.Contains(code, "s.Active") {
		t.Error("Active should be ignored")
	}
}

func TestGenerate_SliceAndMap(t *testing.T) {
	info := &types.FileInfo{
		PkgName: "test",
		Structs: []types.StructInfo{
			{
				Name: "Container",
				Fields: []types.FieldInfo{
					{Name: "Items", TypeStr: "[]string", Kind: types.KindSlice, Action: types.ActionZero},
					{Name: "Data", TypeStr: "map[string]int", Kind: types.KindMap, Action: types.ActionZero},
				},
			},
		},
	}

	code := Generate(info)

	if !strings.Contains(code, "s.Items = s.Items[:0]") {
		t.Error("missing slice truncation")
	}
	if !strings.Contains(code, "clear(s.Data)") {
		t.Error("missing map clear")
	}
}

func TestGenerate_Embedded(t *testing.T) {
	info := &types.FileInfo{
		PkgName: "test",
		Structs: []types.StructInfo{
			{
				Name: "Outer",
				Fields: []types.FieldInfo{
					{Name: "Inner", TypeStr: "Inner", Kind: types.KindBasic, Action: types.ActionZero, IsEmbedded: true},
					{Name: "Value", TypeStr: "int", Kind: types.KindBasic, Action: types.ActionZero},
				},
			},
		},
	}

	code := Generate(info)

	if !strings.Contains(code, "s.Inner.Reset()") {
		t.Error("missing embedded Reset() call")
	}
}

func TestGenerate_EmbeddedPointer(t *testing.T) {
	info := &types.FileInfo{
		PkgName: "test",
		Structs: []types.StructInfo{
			{
				Name: "Doc",
				Fields: []types.FieldInfo{
					{Name: "Meta", TypeStr: "*Meta", Kind: types.KindPointer, Action: types.ActionZero, IsEmbedded: true},
				},
			},
		},
	}

	code := Generate(info)

	if !strings.Contains(code, "if s.Meta != nil") {
		t.Error("missing nil check for embedded pointer")
	}
	if !strings.Contains(code, "s.Meta.Reset()") {
		t.Error("missing Reset() call for embedded pointer")
	}
}

func TestGenerate_Pointer(t *testing.T) {
	info := &types.FileInfo{
		PkgName: "test",
		Structs: []types.StructInfo{
			{
				Name: "Config",
				Fields: []types.FieldInfo{
					{Name: "Value", TypeStr: "*int", Kind: types.KindPointer, Action: types.ActionZero},
				},
			},
		},
	}

	code := Generate(info)

	if !strings.Contains(code, "s.Value = nil") {
		t.Error("missing pointer nil assignment")
	}
}

func TestGenerate_Integration(t *testing.T) {
	src := `package test

type User struct {
	ID       int64    ` + "`reset:\"\"`" + `
	Name     string   ` + "`reset:\"guest\"`" + `
	Tags     []string ` + "`reset:\"\"`" + `
	Settings map[string]any ` + "`reset:\"\"`" + `
	Active   bool     ` + "`reset:\"-\"`" + `
}
`
	info, err := parser.ParseSource(src)
	if err != nil {
		t.Fatalf("ParseSource failed: %v", err)
	}

	code := Generate(info)

	// Verify generated code
	expected := []string{
		"func (s *User) Reset()",
		"s.ID = 0",
		`s.Name = "guest"`,
		"s.Tags = s.Tags[:0]",
		"clear(s.Settings)",
	}

	for _, exp := range expected {
		if !strings.Contains(code, exp) {
			t.Errorf("missing expected code: %s", exp)
		}
	}

	// Active should be ignored
	if strings.Contains(code, "s.Active") {
		t.Error("Active should be ignored")
	}
}

func TestGenerate_MultipleStructs(t *testing.T) {
	src := `package test

type User struct {
	ID   int64  ` + "`reset:\"\"`" + `
	Name string ` + "`reset:\"\"`" + `
}

type Config struct {
	Debug   bool ` + "`reset:\"\"`" + `
	Timeout int  ` + "`reset:\"30\"`" + `
}
`
	info, err := parser.ParseSource(src)
	if err != nil {
		t.Fatalf("ParseSource failed: %v", err)
	}

	code := Generate(info)

	if !strings.Contains(code, "func (s *User) Reset()") {
		t.Error("missing User Reset() method")
	}
	if !strings.Contains(code, "func (s *Config) Reset()") {
		t.Error("missing Config Reset() method")
	}
}

func TestGenerate_Array(t *testing.T) {
	info := &types.FileInfo{
		PkgName: "test",
		Structs: []types.StructInfo{
			{
				Name: "Buffer",
				Fields: []types.FieldInfo{
					{Name: "Data", TypeStr: "[32]byte", Kind: types.KindArray, Action: types.ActionZero},
				},
			},
		},
	}

	code := Generate(info)

	if !strings.Contains(code, "s.Data = [32]byte{}") {
		t.Error("missing array zero")
	}
}

func TestGenerate_Interface(t *testing.T) {
	info := &types.FileInfo{
		PkgName: "test",
		Structs: []types.StructInfo{
			{
				Name: "Handler",
				Fields: []types.FieldInfo{
					{Name: "Callback", TypeStr: "interface{}", Kind: types.KindInterface, Action: types.ActionZero},
				},
			},
		},
	}

	code := Generate(info)

	if !strings.Contains(code, "s.Callback = nil") {
		t.Error("missing interface nil")
	}
}

func TestGenerate_Empty(t *testing.T) {
	info := &types.FileInfo{
		PkgName: "test",
		Structs: []types.StructInfo{},
	}

	code := Generate(info)
	if code != "" {
		t.Errorf("expected empty code, got %q", code)
	}
}

func BenchmarkGenerate(b *testing.B) {
	info := &types.FileInfo{
		PkgName: "test",
		Structs: []types.StructInfo{
			{
				Name: "User",
				Fields: []types.FieldInfo{
					{Name: "ID", TypeStr: "int64", Kind: types.KindBasic, Action: types.ActionZero},
					{Name: "Name", TypeStr: "string", Kind: types.KindBasic, Action: types.ActionDefault, Default: "guest"},
					{Name: "Email", TypeStr: "string", Kind: types.KindBasic, Action: types.ActionZero},
					{Name: "Tags", TypeStr: "[]string", Kind: types.KindSlice, Action: types.ActionZero},
					{Name: "Settings", TypeStr: "map[string]any", Kind: types.KindMap, Action: types.ActionZero},
				},
			},
			{
				Name: "Config",
				Fields: []types.FieldInfo{
					{Name: "Debug", TypeStr: "bool", Kind: types.KindBasic, Action: types.ActionZero},
					{Name: "Timeout", TypeStr: "int", Kind: types.KindBasic, Action: types.ActionDefault, Default: "30"},
					{Name: "MaxRetry", TypeStr: "int", Kind: types.KindBasic, Action: types.ActionDefault, Default: "3"},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = Generate(info)
	}
}

func BenchmarkGenerateStruct(b *testing.B) {
	s := &types.StructInfo{
		Name: "User",
		Fields: []types.FieldInfo{
			{Name: "ID", TypeStr: "int64", Kind: types.KindBasic, Action: types.ActionZero},
			{Name: "Name", TypeStr: "string", Kind: types.KindBasic, Action: types.ActionDefault, Default: "guest"},
			{Name: "Email", TypeStr: "string", Kind: types.KindBasic, Action: types.ActionZero},
			{Name: "Tags", TypeStr: "[]string", Kind: types.KindSlice, Action: types.ActionZero},
			{Name: "Settings", TypeStr: "map[string]any", Kind: types.KindMap, Action: types.ActionZero},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GenerateStruct(s)
	}
}
