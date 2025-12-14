package types

import "go/ast"

// FieldKind represents the kind of a struct field.
type FieldKind int

const (
	KindBasic FieldKind = iota
	KindSlice
	KindMap
	KindPointer
	KindStruct
	KindArray
	KindChan
	KindInterface
)

// TagAction represents what action to take for a reset tag.
type TagAction int

const (
	ActionZero TagAction = iota
	ActionDefault
	ActionIgnore
)

// FieldInfo holds information about a struct field with reset tag.
type FieldInfo struct {
	Name       string
	TypeExpr   ast.Expr
	TypeStr    string
	Kind       FieldKind
	Action     TagAction
	Default    string
	IsEmbedded bool
	IsExported bool
	ElemType   string
	KeyType    string
}

// StructInfo holds information about a struct with reset tags.
type StructInfo struct {
	Name    string
	Fields  []FieldInfo
	PkgName string
}

// FileInfo holds information about a parsed file.
type FileInfo struct {
	Path    string
	PkgName string
	Structs []StructInfo
	Imports []ImportInfo
}

// ImportInfo holds import information.
type ImportInfo struct {
	Path  string
	Alias string
}
