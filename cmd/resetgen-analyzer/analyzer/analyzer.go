// Package analyzer provides a static analyzer that detects sync.Pool usage
// without calling Reset() before Put().
package analyzer

import (
	"go/ast"
	"go/types"

	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/inspect"
	"golang.org/x/tools/go/ast/inspector"
)

var Analyzer = &analysis.Analyzer{
	Name:     "resetcheck",
	Doc:      "checks that Reset() is called before sync.Pool.Put()",
	Requires: []*analysis.Analyzer{inspect.Analyzer},
	Run:      run,
}

func run(pass *analysis.Pass) (any, error) {
	insp := pass.ResultOf[inspect.Analyzer].(*inspector.Inspector)

	// Analyze each function separately
	nodeFilter := []ast.Node{
		(*ast.FuncDecl)(nil),
		(*ast.FuncLit)(nil),
	}

	insp.Preorder(nodeFilter, func(n ast.Node) {
		var body *ast.BlockStmt
		switch fn := n.(type) {
		case *ast.FuncDecl:
			if fn.Body == nil {
				return
			}
			body = fn.Body
		case *ast.FuncLit:
			body = fn.Body
		}

		analyzeFunction(pass, body)
	})

	return nil, nil
}

func analyzeFunction(pass *analysis.Pass, body *ast.BlockStmt) {
	// Track variables that had Reset() called on them
	resetCalled := make(map[string]bool)

	// Walk statements in order
	ast.Inspect(body, func(n ast.Node) bool {
		stmt, ok := n.(*ast.ExprStmt)
		if !ok {
			return true
		}

		call, ok := stmt.X.(*ast.CallExpr)
		if !ok {
			return true
		}

		sel, ok := call.Fun.(*ast.SelectorExpr)
		if !ok {
			return true
		}

		// Check for x.Reset() calls - track any variable that had Reset called
		if sel.Sel.Name == "Reset" && len(call.Args) == 0 {
			varName := extractVarName(sel.X)
			if varName != "" {
				resetCalled[varName] = true
			}
		}

		// Check for sync.Pool.Put(x) calls
		if sel.Sel.Name == "Put" && isSyncPoolMethod(sel, pass.TypesInfo) {
			if len(call.Args) == 1 {
				varName := extractVarName(call.Args[0])
				if varName != "" && !resetCalled[varName] {
					pass.Reportf(call.Pos(), "sync.Pool.Put() called without Reset() on %s", varName)
				}
			}
		}

		return true
	})
}

// extractVarName gets the variable name from an expression
// Handles: x, s.x, s.field.x
func extractVarName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.SelectorExpr:
		// For s.field, we still track by the root identifier
		return extractVarName(e.X)
	case *ast.StarExpr:
		return extractVarName(e.X)
	}
	return ""
}

// isSyncPoolMethod checks if sel is a method on sync.Pool
func isSyncPoolMethod(sel *ast.SelectorExpr, info *types.Info) bool {
	tv, ok := info.Types[sel.X]
	if !ok {
		return false
	}

	t := tv.Type
	if ptr, isPtr := t.(*types.Pointer); isPtr {
		t = ptr.Elem()
	}

	named, isNamed := t.(*types.Named)
	if !isNamed {
		return false
	}

	obj := named.Obj()
	if obj == nil {
		return false
	}

	pkg := obj.Pkg()
	if pkg == nil {
		return false
	}

	return pkg.Path() == "sync" && obj.Name() == "Pool"
}
