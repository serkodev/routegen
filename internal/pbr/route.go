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

type SubRoute struct {
	Sub  string
	Sels []string
	Path string // sub path
}

type RoutePackage struct {
	RelativePath      string
	RelativeGroupPath string
	PkgPath           string
	Sels              []string
	SubRoutes         []*SubRoute
	SubPackages       []*RoutePackage // for middleware

	importSpec *ast.ImportSpec
}

func (r *RoutePackage) hasMiddleware() bool {
	for _, sel := range r.Sels {
		if sel == "Middleware" {
			return true
		}
	}

	// for _, route := range r.SubRoutes {
	// 	for _, sel := range route.Sels {
	// 		if sel == "Middleware" {
	// 			return true
	// 		}
	// 	}
	// }

	return false
}

func (r *RoutePackage) routePath() string {
	routePath := r.RelativePath
	if r.RelativeGroupPath != "" {
		routePath = r.RelativeGroupPath
	}
	return getRoutePath(routePath)
}

type RouteTypeCustomOption struct {
	PathComponentAlias string
}

type routeGroup struct {
	path  string
	route *RoutePackage
}

type routeGen struct {
	targetSels []string // sorted target selectors
	sels       map[string]struct{}
}

var pbrRegex = regexp.MustCompile(`^//\s*pbr\s+(.*)$`)

func newRouteGen() *routeGen {
	r := &routeGen{
		targetSels: []string{"Middleware", "GET", "POST", "HANDLE"},
	}
	set := make(map[string]struct{}, len(r.targetSels))
	for _, s := range r.targetSels {
		set[s] = struct{}{}
	}
	r.sels = set
	return r
}

func (r *routeGen) parseRoute(root string) []*RoutePackage {
	var routes []*RoutePackage
	var groupStack []*routeGroup

	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if info.IsDir() {
			fmt.Println("path", path)

			var group *routeGroup
			for len(groupStack) > 0 {
				g := groupStack[len(groupStack)-1]
				if strings.HasPrefix(path, g.path+"/") || g.path == root {
					group = g
					fmt.Println("\tgroup ->", group.path)
					break
				} else {
					// pop stack
					groupStack = groupStack[:len(groupStack)-1]
				}
			}

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
				if route := r.processPkgRouteSels(pkg, rel); route != nil {

					if route.hasMiddleware() {
						fmt.Println("find middleware ->", path)

						// push stack
						groupStack = append(groupStack, &routeGroup{
							path:  path,
							route: route,
						})
					}

					// parent group
					if group != nil {
						grel, _ := filepath.Rel(group.path, path)
						route.RelativeGroupPath = grel
						// fmt.Println("group path", group.path)
						// fmt.Println("rel", path)
						// fmt.Println("grel", grel)

						group.route.SubPackages = append(group.route.SubPackages, route)
					} else {
						routes = append(routes, route)
					}
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

func (r *routeGen) processPkgRouteSels(pkg *packages.Package, relativePath string) *RoutePackage {
	route := &RoutePackage{
		RelativePath: relativePath,
		PkgPath:      pkg.PkgPath,
	}

	// var subRouteSet []RouteSel
	subRouteSet := make(map[string][]string)
	for _, f := range pkg.Syntax {
		ast.Inspect(f, func(n ast.Node) bool {
			if fd, ok := n.(*ast.FuncDecl); ok {
				sel := fd.Name.Name

				if r.isTargetSelector(sel) {
					if rt := r.getFuncRecvType(fd); rt != nil {
						if !isPublicVar(rt) && relativePath != "." {
							return true
						}
						subRouteSet[rt.Name] = append(subRouteSet[rt.Name], sel)
					} else {
						route.Sels = append(route.Sels, sel)
					}

					// fmt.Println("route", fd.Name, pkg.PkgPath)
				}
			}
			return true
		})
	}

	// sort sels
	route.Sels = r.sortSels(route.Sels)

	// generate sub routes
	if len(subRouteSet) > 0 {
		options := r.processTypeOption(pkg)
		rs := make([]*SubRoute, 0, len(subRouteSet))
		for sub, sels := range subRouteSet {
			opt := options[sub]
			rs = append(rs, &SubRoute{
				Sub:  sub,
				Sels: r.sortSels(sels),
				Path: getSubRoutePath(sub, opt),
			})
		}
		route.SubRoutes = rs
	}

	if len(route.Sels) == 0 && len(route.SubRoutes) == 0 {
		return nil
	}

	return route
}

func (r *routeGen) sortSels(sels []string) []string {
	if sels == nil {
		return nil
	}
	sorted := make([]string, 0, len(sels))
	for _, targetSel := range r.targetSels {
		for _, sel := range sels {
			if targetSel != sel {
				continue
			}
			sorted = append(sorted, sel)
		}
	}
	return sorted
}

func getRoutePath(relativePath string, subPath ...string) string {
	if relativePath == "." {
		relativePath = ""
	}
	subPath = append([]string{"/", relativePath}, subPath...)
	return buildParamPath(filepath.Join(subPath...))
}

func getSubRoutePath(sub string, opt *RouteTypeCustomOption) string {
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
