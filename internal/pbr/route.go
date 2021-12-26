package pbr

import (
	"fmt"
	"go/ast"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

type PkgRoute struct {
	RelativePath string
	PkgPath      string
	Sels         []string
}

var targetSels = []string{"GET", "POST", "HANDLE"}

func parseRoute(root string) []*PkgRoute {
	println("root", root)
	var routes []*PkgRoute

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			// get relative path
			rel, _ := filepath.Rel(root, path)
			if rel == "." {
				return nil
			}

			cfg := &packages.Config{
				Mode: packages.NeedName | packages.NeedCompiledGoFiles | packages.NeedSyntax,
				Dir:  path,
			}
			pkgs, err := packages.Load(cfg)
			if err != nil {
				return err
			}

			for _, pkg := range pkgs {
				if sels := processPkgRouteSels(pkg); len(sels) > 0 {
					routes = append(routes, &PkgRoute{
						RelativePath: rel,
						PkgPath:      pkg.PkgPath,
						Sels:         sels,
					})
				}
			}
		}
		return nil
	})

	fmt.Printf("routes: %v", routes)
	return routes
}

func processPkgRouteSels(pkg *packages.Package) []string {
	var sels []string
	for _, f := range pkg.Syntax {
		for _, sel := range targetSels {
			if o := f.Scope.Lookup(sel); o != nil {
				if _, ok := o.Decl.(*ast.FuncDecl); ok {
					sels = append(sels, sel)
				}
			}
		}
	}
	return sels
}
