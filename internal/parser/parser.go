package parser

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"reflect"
	"strings"

	"github.com/flaticols/resetgen/internal/types"
)

const (
	tagName       = "reset"
	toolDirective = "+resetgen"
)

// ParseFile parses a Go source file and extracts structs with reset tags.
// If structFilter is provided, only the listed struct names are processed.
func ParseFile(path string, structFilter map[string]bool) (*types.FileInfo, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse file %s: %w", path, err)
	}

	info := &types.FileInfo{
		Path:    path,
		PkgName: f.Name.Name,
	}

	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			structInfo := parseStruct(typeSpec.Name.Name, structType, genDecl, structFilter)
			if structInfo != nil {
				structInfo.PkgName = info.PkgName
				info.Structs = append(info.Structs, *structInfo)
			}
		}
	}

	return info, nil
}

// ParseSource parses Go source code from a string.
// Kept for backward compatibility with existing tests.
func ParseSource(src string) (*types.FileInfo, error) {
	return ParseSourceWithFilter(src, nil)
}

// ParseSourceWithFilter parses Go source code with an optional struct filter.
// If structFilter is provided, only the listed struct names are processed.
func ParseSourceWithFilter(src string, structFilter map[string]bool) (*types.FileInfo, error) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, "source.go", src, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("parse source: %w", err)
	}

	info := &types.FileInfo{
		Path:    "source.go",
		PkgName: f.Name.Name,
	}

	for _, decl := range f.Decls {
		genDecl, ok := decl.(*ast.GenDecl)
		if !ok || genDecl.Tok != token.TYPE {
			continue
		}

		for _, spec := range genDecl.Specs {
			typeSpec, ok := spec.(*ast.TypeSpec)
			if !ok {
				continue
			}

			structType, ok := typeSpec.Type.(*ast.StructType)
			if !ok {
				continue
			}

			structInfo := parseStruct(typeSpec.Name.Name, structType, genDecl, structFilter)
			if structInfo != nil {
				structInfo.PkgName = info.PkgName
				info.Structs = append(info.Structs, *structInfo)
			}
		}
	}

	return info, nil
}

// hasResetgenDirective reports whether genDecl has the +resetgen comment directive.
// Recognizes various formats: "//+resetgen", "// +resetgen", "/*+resetgen*/", etc.
func hasResetgenDirective(genDecl *ast.GenDecl) bool {
	if genDecl.Doc == nil {
		return false
	}

	for _, comment := range genDecl.Doc.List {
		text := strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
		text = strings.TrimSpace(strings.TrimPrefix(text, "/*"))
		text = strings.TrimSuffix(strings.TrimSpace(text), "*/")

		if strings.HasPrefix(text, toolDirective) {
			return true
		}
	}

	return false
}

// isExportedType reports whether expr refers to an exported type.
// Pointer-to-type and package-qualified types are considered exported.
func isExportedType(expr ast.Expr) bool {
	switch t := expr.(type) {
	case *ast.Ident:
		return ast.IsExported(t.Name)
	case *ast.StarExpr:
		return isExportedType(t.X)
	case *ast.SelectorExpr:
		return true
	default:
		return false
	}
}

// checkHasResetTag reports whether any field in the struct has a reset tag.
func checkHasResetTag(fields *ast.FieldList) bool {
	for _, field := range fields.List {
		if field.Tag != nil {
			if _, hasTag := parseTag(field.Tag.Value); hasTag {
				return true
			}
		}
	}
	return false
}

// parseStruct extracts struct field information based on reset tags and directives.
// When structFilter is provided, all exported fields are included; otherwise only
// fields with reset tags or structs with +resetgen directives are processed.
// Returns nil if the struct should not be processed or has no non-ignored fields.
func parseStruct(name string, st *ast.StructType, genDecl *ast.GenDecl, structFilter map[string]bool) *types.StructInfo {
	if st.Fields == nil {
		return nil
	}

	var shouldProcess bool
	var processAllExported bool

	if structFilter != nil {
		_, shouldProcess = structFilter[name]
		processAllExported = shouldProcess
	} else {
		hasResetTag := checkHasResetTag(st.Fields)
		hasDirective := hasResetgenDirective(genDecl)
		shouldProcess = hasResetTag || hasDirective
		processAllExported = hasDirective
	}

	if !shouldProcess {
		return nil
	}

	var fields []types.FieldInfo
	hasNonIgnoredFields := false

	for _, field := range st.Fields.List {
		var tag string
		var hasTag bool

		if field.Tag != nil {
			tag, hasTag = parseTag(field.Tag.Value)
		}

		if len(field.Names) == 0 {
			if !hasTag && !processAllExported {
				continue
			}

			if processAllExported && !isExportedType(field.Type) {
				continue
			}

			tagVal := ""
			if hasTag {
				tagVal = tag
			}
			fi := parseField("", field.Type, tagVal, true)
			fields = append(fields, fi)
			if fi.Action != types.ActionIgnore {
				hasNonIgnoredFields = true
			}
			continue
		}

		for _, ident := range field.Names {
			if !hasTag && !processAllExported {
				continue
			}

			if processAllExported && !ast.IsExported(ident.Name) {
				continue
			}

			tagVal := ""
			if hasTag {
				tagVal = tag
			}
			fi := parseField(ident.Name, field.Type, tagVal, false)
			fields = append(fields, fi)
			if fi.Action != types.ActionIgnore {
				hasNonIgnoredFields = true
			}
		}
	}

	if len(fields) == 0 || !hasNonIgnoredFields {
		return nil
	}

	return &types.StructInfo{
		Name:   name,
		Fields: fields,
	}
}

func parseTag(tagLit string) (string, bool) {
	if len(tagLit) < 2 {
		return "", false
	}
	tagStr := tagLit[1 : len(tagLit)-1]
	st := reflect.StructTag(tagStr)
	return st.Lookup(tagName)
}

// parseField creates a FieldInfo from an AST field expression and tag value.
// Determines the field's type kind, name, and reset action based on the tag.
func parseField(name string, typeExpr ast.Expr, tagVal string, embedded bool) types.FieldInfo {
	fi := types.FieldInfo{
		Name:       name,
		TypeExpr:   typeExpr,
		TypeStr:    exprToString(typeExpr),
		Kind:       getFieldKind(typeExpr),
		IsEmbedded: embedded,
		IsExported: name == "" || ast.IsExported(name),
	}

	if embedded {
		fi.Name = getEmbeddedName(typeExpr)
	}

	switch tagVal {
	case "-":
		fi.Action = types.ActionIgnore
	case "":
		fi.Action = types.ActionZero
	default:
		fi.Action = types.ActionDefault
		fi.Default = tagVal
	}

	switch t := typeExpr.(type) {
	case *ast.ArrayType:
		fi.ElemType = exprToString(t.Elt)
	case *ast.MapType:
		fi.KeyType = exprToString(t.Key)
		fi.ElemType = exprToString(t.Value)
	case *ast.StarExpr:
		fi.ElemType = exprToString(t.X)
	case *ast.ChanType:
		fi.ElemType = exprToString(t.Value)
	}

	return fi
}

func getFieldKind(expr ast.Expr) types.FieldKind {
	switch t := expr.(type) {
	case *ast.ArrayType:
		if t.Len == nil {
			return types.KindSlice
		}
		return types.KindArray
	case *ast.MapType:
		return types.KindMap
	case *ast.StarExpr:
		return types.KindPointer
	case *ast.StructType:
		return types.KindStruct
	case *ast.ChanType:
		return types.KindChan
	case *ast.InterfaceType:
		return types.KindInterface
	case *ast.Ident:
		return types.KindBasic
	case *ast.SelectorExpr:
		return types.KindStruct
	default:
		return types.KindBasic
	}
}

func getEmbeddedName(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return getEmbeddedName(t.X)
	case *ast.SelectorExpr:
		return t.Sel.Name
	default:
		return ""
	}
}

// exprToString converts an AST expression to its string representation.
func exprToString(expr ast.Expr) string {
	switch t := expr.(type) {
	case *ast.Ident:
		return t.Name
	case *ast.StarExpr:
		return "*" + exprToString(t.X)
	case *ast.ArrayType:
		if t.Len == nil {
			return "[]" + exprToString(t.Elt)
		}
		return "[" + exprToString(t.Len) + "]" + exprToString(t.Elt)
	case *ast.MapType:
		return "map[" + exprToString(t.Key) + "]" + exprToString(t.Value)
	case *ast.SelectorExpr:
		return exprToString(t.X) + "." + t.Sel.Name
	case *ast.ChanType:
		switch t.Dir {
		case ast.SEND:
			return "chan<- " + exprToString(t.Value)
		case ast.RECV:
			return "<-chan " + exprToString(t.Value)
		default:
			return "chan " + exprToString(t.Value)
		}
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.StructType:
		return "struct{}"
	case *ast.FuncType:
		return "func()"
	case *ast.BasicLit:
		return t.Value
	case *ast.Ellipsis:
		return "..." + exprToString(t.Elt)
	case *ast.IndexExpr:
		return exprToString(t.X) + "[" + exprToString(t.Index) + "]"
	case *ast.IndexListExpr:
		var indices []string
		for _, idx := range t.Indices {
			indices = append(indices, exprToString(idx))
		}
		return exprToString(t.X) + "[" + strings.Join(indices, ", ") + "]"
	default:
		return "unknown"
	}
}
