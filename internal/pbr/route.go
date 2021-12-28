package pbr

import (
	"fmt"
	"go/ast"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

type RouteSels = map[string][]string

type RoutePackage struct {
	RelativePath string
	PkgPath      string
	Handles      RouteSels
}

type routeGen struct {
	sels map[string]struct{}
}

func newRouteGen() *routeGen {
	r := &routeGen{}

	// target selectors
	var targetSels = []string{"GET", "POST", "HANDLE"}
	set := make(map[string]struct{}, len(targetSels))
	for _, s := range targetSels {
		set[s] = struct{}{}
	}
	r.sels = set

	return r
}

func (r *routeGen) parseRoute(root string) []*RoutePackage {
	println("root", root)
	var routes []*RoutePackage

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
				if sels := r.processPkgRouteSels(pkg); len(sels) > 0 {
					routes = append(routes, &RoutePackage{
						RelativePath: rel,
						PkgPath:      pkg.PkgPath,
						Handles:      sels,
					})
				}
			}
		}
		return nil
	})

	fmt.Printf("routes: %v", routes)
	return routes
}

func (r *routeGen) processPkgRouteSels(pkg *packages.Package) RouteSels {
	// if pkg.PkgPath != "example.com/foo/router/api/post" {
	// 	return []string{}
	// }

	var sels = make(RouteSels)
	for _, f := range pkg.Syntax {
		// fmt.Printf("routes: %v\n", f.Scope.String())
		// ast.Print(pkg.Fset, f)

		ast.Inspect(f, func(n ast.Node) bool {
			if fd, ok := n.(*ast.FuncDecl); ok {
				sel := fd.Name.Name

				rt := r.getFuncRecvType(fd)
				if rt != nil {
					// TODO: handle with recv
				} else {
					if r.isTargetSelector(sel) {
						sels[""] = append(sels[""], sel)
					}
					fmt.Println("func", fd.Name, pkg.PkgPath)
				}
			}
			return true
		})
	}
	return sels
}

func (r *routeGen) getFuncRecvType(fd *ast.FuncDecl) *ast.Ident {
	if fd.Recv == nil || len(fd.Recv.List) != 1 {
		return nil
	}

	switch recvType := fd.Recv.List[0].Type.(type) {
	case *ast.Ident:
		return recvType
	case *ast.StarExpr:
		if rt, ok := recvType.X.(*ast.Ident); ok {
			return rt
		}
	}

	return nil
}

func (r *routeGen) isTargetSelector(sel string) bool {
	_, ok := r.sels[sel]
	return ok
}
