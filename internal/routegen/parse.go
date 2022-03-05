package routegen

import (
	"errors"
	"go/ast"
	"go/types"
	"strings"
)

func isroutegenImport(path string) bool {
	// TODO(light): This is depending on details of the current loader.
	const vendorPart = "vendor/"
	if i := strings.LastIndex(path, vendorPart); i != -1 && (i == 0 || path[i-1] == '/') {
		path = path[i+len(vendorPart):]
	}
	return path == "github.com/serkodev/routegen"
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

// findInjectorBuild returns the routegen.Build call if fn is an injector template.
// It returns nil if the function is not an injector template.
func findInjectorBuild(info *types.Info, fn *ast.FuncDecl) (*ast.CallExpr, error) {
	if fn.Body == nil {
		return nil, nil
	}
	numStatements := 0
	invalid := false
	var routegenBuildCall *ast.CallExpr
	for _, stmt := range fn.Body.List {
		switch stmt := stmt.(type) {
		case *ast.ExprStmt:
			numStatements++
			if numStatements > 1 {
				invalid = true
			}
			call := getInjectorStmt(info, stmt)
			if call == nil {
				continue
			}
			routegenBuildCall = call
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
	if routegenBuildCall == nil {
		return nil, nil
	}
	if invalid {
		return nil, errors.New("a call to routegen.Build indicates that this function is an injector, but injectors must consist of only the routegen.Build call and an optional return")
	}
	return routegenBuildCall, nil
}

func getInjectorStmt(info *types.Info, stmt ast.Stmt) *ast.CallExpr {
	if es, ok := stmt.(*ast.ExprStmt); ok {
		call, ok := es.X.(*ast.CallExpr)
		if !ok {
			return nil
		}
		if qualifiedIdentObject(info, call.Fun) == types.Universe.Lookup("panic") {
			if len(call.Args) != 1 {
				return nil
			}
			call, ok = call.Args[0].(*ast.CallExpr)
			if !ok {
				return nil
			}
		}
		buildObj := qualifiedIdentObject(info, call.Fun)
		if buildObj == nil || buildObj.Pkg() == nil || !isroutegenImport(buildObj.Pkg().Path()) || buildObj.Name() != "Build" {
			return nil
		}
		return call
	}
	return nil
}
