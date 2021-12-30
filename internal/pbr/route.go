package pbr

import (
	"fmt"
	"go/ast"
	"os"
	"path/filepath"

	"golang.org/x/tools/go/packages"
)

type RouteSel struct {
	Sub  string
	Sels []string
}

//= map[string][]string // key: sub route, if empty then means index page, value: handles (GET, POST, etc.)

type RoutePackage struct {
	RelativePath string
	PkgPath      string
	RouteSels    []*RouteSel
	importSpec   *ast.ImportSpec
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
	var routes []*RoutePackage
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			// get relative path
			rel, _ := filepath.Rel(root, path)

			cfg := &packages.Config{
				Mode: packages.NeedName | packages.NeedCompiledGoFiles | packages.NeedSyntax,
				Dir:  path,
			}
			pkgs, err := packages.Load(cfg)
			if err != nil {
				return err
			}

			for _, pkg := range pkgs {
				if rs := r.processPkgRouteSels(pkg); len(rs) > 0 {
					routes = append(routes, &RoutePackage{
						RelativePath: rel,
						PkgPath:      pkg.PkgPath,
						RouteSels:    rs,
					})
				}
			}
		}
		return nil
	})
	return routes
}

func (r *routeGen) processPkgRouteSels(pkg *packages.Package) []*RouteSel {
	// var selsSet []RouteSel
	selsSet := make(map[string][]string)
	for _, f := range pkg.Syntax {
		ast.Inspect(f, func(n ast.Node) bool {
			if fd, ok := n.(*ast.FuncDecl); ok {
				sel := fd.Name.Name

				rt := r.getFuncRecvType(fd)
				if rt != nil {
					// TODO: (sub route) handle with recv
				} else {
					if r.isTargetSelector(sel) {
						selsSet[""] = append(selsSet[""], sel)
					}
					fmt.Println("route", fd.Name, pkg.PkgPath)
				}
			}
			return true
		})
	}

	rs := make([]*RouteSel, 0, len(selsSet))
	for sub, sels := range selsSet {
		rs = append(rs, &RouteSel{
			Sub:  sub,
			Sels: sels,
		})
	}
	return rs
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
