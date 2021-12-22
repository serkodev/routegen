package pbr

import (
	"errors"
	"go/ast"
	"go/types"
	"strings"
)

func isWireImport(path string) bool {
	// TODO(light): This is depending on details of the current loader.
	const vendorPart = "vendor/"
	if i := strings.LastIndex(path, vendorPart); i != -1 && (i == 0 || path[i-1] == '/') {
		path = path[i+len(vendorPart):]
	}
	return path == "github.com/serkodev/pbr"
}

// qualifiedIdentObject finds the object for an identifier or a
// qualified identifier, or nil if the object could not be found.
func qualifiedIdentObject(info *types.Info, expr ast.Expr) types.Object {
	switch expr := expr.(type) {
	case *ast.Ident:
		return info.ObjectOf(expr)
	case *ast.SelectorExpr:
		pkgName, ok := expr.X.(*ast.Ident)
		if !ok {
			return nil
		}
		if _, ok := info.ObjectOf(pkgName).(*types.PkgName); !ok {
			return nil
		}
		return info.ObjectOf(expr.Sel)
	default:
		return nil
	}
}

// findInjectorBuild returns the wire.Build call if fn is an injector template.
// It returns nil if the function is not an injector template.
func findInjectorBuild(info *types.Info, fn *ast.FuncDecl) (*ast.CallExpr, error) {
	if fn.Body == nil {
		return nil, nil
	}
	numStatements := 0
	invalid := false
	var wireBuildCall *ast.CallExpr
	for _, stmt := range fn.Body.List {
		switch stmt := stmt.(type) {
		case *ast.ExprStmt:
			numStatements++
			if numStatements > 1 {
				invalid = true
			}
			call, ok := stmt.X.(*ast.CallExpr)
			if !ok {
				continue
			}
			if qualifiedIdentObject(info, call.Fun) == types.Universe.Lookup("panic") {
				if len(call.Args) != 1 {
					continue
				}
				call, ok = call.Args[0].(*ast.CallExpr)
				if !ok {
					continue
				}
			}
			buildObj := qualifiedIdentObject(info, call.Fun)
			if buildObj == nil || buildObj.Pkg() == nil || !isWireImport(buildObj.Pkg().Path()) || buildObj.Name() != "Build" {
				continue
			}
			wireBuildCall = call
		case *ast.EmptyStmt:
			// Do nothing.
		case *ast.ReturnStmt:
			// Allow the function to end in a return.
			if numStatements == 0 {
				return nil, nil
			}
		default:
			invalid = true
		}

	}
	if wireBuildCall == nil {
		return nil, nil
	}
	if invalid && wireBuildCall == nil /* TODO: FIX */ {
		return nil, errors.New("a call to wire.Build indicates that this function is an injector, but injectors must consist of only the wire.Build call and an optional return")
	}
	return wireBuildCall, nil
}

func getInjectorStmt(info *types.Info, stmt ast.Stmt) *ast.CallExpr {
	if es, ok := stmt.(*ast.ExprStmt); ok {
		if call, ok := es.X.(*ast.CallExpr); ok {
			buildObj := qualifiedIdentObject(info, call.Fun)
			if buildObj == nil || buildObj.Pkg() == nil || !isWireImport(buildObj.Pkg().Path()) || buildObj.Name() != "Build" {
				return nil
			}
			return call
		}
	}
	return nil
}
