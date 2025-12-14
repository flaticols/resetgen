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

const tagName = "reset"

// ParseFile parses a Go source file and extracts structs with reset tags.
func ParseFile(path string) (*types.FileInfo, error) {
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

			structInfo := parseStruct(typeSpec.Name.Name, structType)
			if structInfo != nil {
				structInfo.PkgName = info.PkgName
				info.Structs = append(info.Structs, *structInfo)
			}
		}
	}

	return info, nil
}

// ParseSource parses Go source code from a string.
func ParseSource(src string) (*types.FileInfo, error) {
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

			structInfo := parseStruct(typeSpec.Name.Name, structType)
			if structInfo != nil {
				structInfo.PkgName = info.PkgName
				info.Structs = append(info.Structs, *structInfo)
			}
		}
	}

	return info, nil
}

func parseStruct(name string, st *ast.StructType) *types.StructInfo {
	if st.Fields == nil {
		return nil
	}

	var fields []types.FieldInfo
	hasResetTag := false

	for _, field := range st.Fields.List {
		if field.Tag == nil {
			continue
		}

		tag, ok := parseTag(field.Tag.Value)
		if !ok {
			continue
		}

		hasResetTag = true

		if len(field.Names) == 0 {
			fi := parseField("", field.Type, tag, true)
			fields = append(fields, fi)
			continue
		}

		for _, ident := range field.Names {
			fi := parseField(ident.Name, field.Type, tag, false)
			fields = append(fields, fi)
		}
	}

	if !hasResetTag {
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
