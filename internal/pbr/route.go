package pbr

import (
	"fmt"
	"go/ast"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"golang.org/x/tools/go/packages"
)

type RouteSel struct {
	Sub  string
	Sels []string
	Path string // sub path
}

type RoutePackage struct {
	RelativePath      string
	PkgPath           string
	RouteSels         []*RouteSel
	SubPackages       []*RoutePackage // for middleware

	importSpec *ast.ImportSpec
}

type RouteTypeCustomOption struct {
	PathComponentAlias string
}

type routeGen struct {
	sels map[string]struct{}
}

var pbrRegex = regexp.MustCompile(`^//\s*pbr\s+(.*)$`)

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
				if rs := r.processPkgRouteSels(pkg, rel); len(rs) > 0 {
					routes = append(routes, &RoutePackage{
						RelativePath: relatvePath(rel),
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

func (r *routeGen) processTypeOption(pkg *packages.Package) map[string]*RouteTypeCustomOption {
	options := make(map[string]*RouteTypeCustomOption)

	for _, f := range pkg.Syntax {
		ast.Inspect(f, func(n ast.Node) bool {
			if decl, ok := n.(*ast.GenDecl); ok {
				if decl.Tok == token.TYPE && decl.Doc != nil {
					var option RouteTypeCustomOption
					setted := false

					for _, comment := range decl.Doc.List {
						match := pbrRegex.FindStringSubmatch(comment.Text)
						if match != nil {
							query := match[1]
							if s := strings.SplitN(query, "=", 2); len(s) == 2 {
								key, val := s[0], s[1]
								if key == "alias" {
									option.PathComponentAlias = val
									setted = true
								}
							}
						}
					}

					if setted {
						if typeIdent := getTypeIdentFromGenDecl(decl); typeIdent != nil {
							options[typeIdent.Name] = &option
						}
					}
				}
			}
			return true
		})
	}

	return options
}

func getTypeIdentFromGenDecl(decl *ast.GenDecl) *ast.Ident {
	if decl.Tok == token.TYPE {
		if len(decl.Specs) == 1 {
			if ts, ok := decl.Specs[0].(*ast.TypeSpec); ok {
				return ts.Name
			}
		}
	}
	return nil
}

func (r *routeGen) processPkgRouteSels(pkg *packages.Package, relativePath string) []*RouteSel {
	// var selsSet []RouteSel
	selsSet := make(map[string][]string)
	for _, f := range pkg.Syntax {
		ast.Inspect(f, func(n ast.Node) bool {
			if fd, ok := n.(*ast.FuncDecl); ok {
				sel := fd.Name.Name

				if r.isTargetSelector(sel) {
					sub := ""
					if rt := r.getFuncRecvType(fd); rt != nil {
						if !isPublicVar(rt) && relativePath != "." {
							return true
						}
						sub = rt.Name
					}
					selsSet[sub] = append(selsSet[sub], sel)

					fmt.Println("route", fd.Name, pkg.PkgPath)
				}
			}
			return true
		})
	}

	options := r.processTypeOption(pkg)
	rs := make([]*RouteSel, 0, len(selsSet))
	for sub, sels := range selsSet {
		opt := options[sub]
		rs = append(rs, &RouteSel{
			Sub:  sub,
			Sels: sels,
			Path: r.getRoutePath(sub, opt),
		})
	}
	return rs
}

func relatvePath(path string) string {
	if path == "." {
		path = ""
	}
	return buildParamPath(filepath.Join("/", path))
}

func buildParamPath(path string) string {
	pathComponents := strings.Split(path, "/")
	for i, pathComponent := range pathComponents {
		if len(pathComponent) >= 2 {
			if pathComponent[0:2] == "__" {
				pathComponents[i] = pathComponent[1:]
			} else if pathComponent[0:1] == "_" {
				pathComponents[i] = ":" + pathComponent[1:]
			}
		}
	}
	return strings.Join(pathComponents, "/")
}

func (r *routeGen) getRoutePath(sub string, opt *RouteTypeCustomOption) string {
	path := ""

	// sub route
	if sub != "" {
		// apply alias
		if opt != nil && opt.PathComponentAlias != "" {
			path = filepath.Join(path, opt.PathComponentAlias)
		} else {
			path = filepath.Join(path, kebabCaseString(sub))
		}
	}

	return buildParamPath(path)
}

func isPublicVar(ident *ast.Ident) bool {
	if len(ident.Name) == 0 {
		return false
	}
	firstChar := string(ident.Name[0:1])
	return firstChar == strings.ToUpper(firstChar)
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
