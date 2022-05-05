package routegen

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

type gen struct {
	pkg    *packages.Package
	routes []*RoutePackage
}

func newGen(pkg *packages.Package, routes []*RoutePackage) *gen {
	return &gen{
		pkg:    pkg,
		routes: routes,
	}
}

func Load(wd string, env []string, path string, customEngine string) ([]result, error) {
	cfg := &packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | // LoadFiles
			packages.NeedImports | // LoadImports
			packages.NeedTypes | packages.NeedTypesSizes | // LoadTypes
			packages.NeedSyntax | packages.NeedTypesInfo | // LoadSyntax
			packages.NeedDeps, // LoadTypes
		Dir:        wd,
		Env:        env,
		BuildFlags: []string{"-tags=routegeninject"},
	}

	pkgs, err := packages.Load(cfg, "pattern="+path)
	if err != nil {
		return nil, err
	}

	var engine *engine
	em, err := newEngineManager(customEngine)
	if err != nil {
		return nil, err
	}

	var results []result

	for _, pkg := range pkgs {
		for _, f := range pkg.Syntax {
			injectFuncsIdentSet := make(map[*ast.FuncDecl]*ast.Ident)

			ast.Inspect(f, func(n ast.Node) bool {
				fn, ok := n.(*ast.FuncDecl)
				if !ok {
					return true
				}

				buildCall, err := findInjectorBuild(pkg.TypesInfo, fn)
				if err != nil {
					fmt.Println("findInjectorBuild error", err.Error())
					return true
				}

				if buildCall != nil {
					// inject build found, assign engine
					for _, arg := range buildCall.Args {
						if obj := qualifiedIdentObject(pkg.TypesInfo, arg); obj != nil {
							if e := em.matchEngine(obj); e != nil {
								if engine != nil && engine != e {
									panic("not support multi type builder.")
								}
								injectFuncsIdentSet[fn] = arg.(*ast.Ident)
								engine = e
								return true
							}
						}
					}
				}
				return true
			})

			if len(injectFuncsIdentSet) > 0 {
				r := newRouteGen(engine.TargetSels(), engine.MiddlewareSelector())
				routes := r.parseRoute(filepath.Join(wd, path))

				g := newGen(pkg, routes)
				result, err := g.inject(f, engine, injectFuncsIdentSet)
				if err != nil {
					return nil, err
				}
				results = append(results, result)
			}
		}
	}

	return results, nil
}

func (g *gen) generate(f *ast.File) ([]byte, error) {
	var out bytes.Buffer

	out.WriteString("// Code generated by routegen. DO NOT EDIT.\n")
	out.WriteString("\n")
	out.WriteString("//go:build !routegeninject\n")
	out.WriteString("// +build !routegeninject\n")
	out.WriteString("\n")
	out.WriteString("package " + g.pkg.Name + "\n")
	out.WriteString("\n")

	for i, decl := range f.Decls {
		// printAST(token.NewFileSet(), decl)
		// fmt.Print("\n")

		if err := printer.Fprint(&out, token.NewFileSet(), decl); err != nil {
			return nil, err
		}

		if i < len(f.Decls)-1 {
			out.WriteString("\n\n")
		} else {
			out.WriteString("\n") // EOF
		}
	}

	return out.Bytes(), nil
}

func (g *gen) getOuputPath(f *ast.File) string {
	oriFilePath := g.pkg.Fset.File(f.Pos()).Name()
	filename := filepath.Base(oriFilePath)
	ext := strings.ToLower(filepath.Ext(filename))
	nonExt := filename[:len(filename)-len(ext)]
	dir := filepath.Dir(oriFilePath)
	return dir + "/" + nonExt + "_gen" + ext
}

func (g *gen) inject(f *ast.File, engine *engine, injectFuncsIdentSet map[*ast.FuncDecl]*ast.Ident) (result, error) {
	// build imports
	scopes := make([]*types.Scope, len(injectFuncsIdentSet))
	for fn := range injectFuncsIdentSet {
		scopes = append(scopes, g.pkg.TypesInfo.Scopes[fn.Type])
	}
	n := newNamerWithScopes(scopes)
	specs, routePackagesImports := g.importsFromRoutes(g.routes, n)

	astutil.Apply(f, func(c *astutil.Cursor) bool {
		if imp, ok := c.Node().(*ast.ImportSpec); ok {
			// inject imports
			if imp.Path.Value == `"github.com/serkodev/routegen"` {
				for _, spec := range specs {
					c.InsertBefore(spec)
				}
				c.Delete()
			}
		} else if fn, ok := c.Node().(*ast.FuncDecl); ok {
			// inject body
			if ident, ok := injectFuncsIdentSet[fn]; ok {
				scope := g.pkg.TypesInfo.Scopes[fn.Type]
				n := newNamer(scope)

				sbuf := new(strings.Builder)
				g.buildInjectStmts(sbuf, engine, ident, g.routes, routePackagesImports, n)

				stmts, err := parseExprs(sbuf.String())
				if err != nil {
					panic("parse exprs error.")
				}

				block := &ast.BlockStmt{List: stmts}
				fn.Body = block
			}
		}
		return true
	}, nil)

	// generate result
	content, err := g.generate(f)
	if err != nil {
		return result{}, err
	}

	return result{
		outPath: g.getOuputPath(f),
		content: content,
	}, nil
}

func (g *gen) importsFromRoutes(routes []*RoutePackage, n *namer) ([]*ast.ImportSpec, map[*RoutePackage]*ast.ImportSpec) {
	var specs []*ast.ImportSpec
	routePackagesImports := make(map[*RoutePackage]*ast.ImportSpec)
	for _, r := range routes {
		if r.RelativePath != "." {
			newName := n.gen("routegen_r")
			spec := &ast.ImportSpec{
				Path: &ast.BasicLit{
					Kind:  token.STRING,
					Value: `"` + r.PkgPath + `"`,
				},
				Name: ast.NewIdent(newName),
			}
			routePackagesImports[r] = spec
			specs = append(specs, spec)
		}
		// sub packages
		if len(r.SubPackages) > 0 {
			subSpecs, subImports := g.importsFromRoutes(r.SubPackages, n)
			specs = append(specs, subSpecs...)
			for k, v := range subImports {
				routePackagesImports[k] = v
			}
		}
	}
	return specs, routePackagesImports
}

func (g *gen) buildInjectStmts(buf *strings.Builder, e *engine, ident *ast.Ident, routePkgs []*RoutePackage, routePackagesImports map[*RoutePackage]*ast.ImportSpec, n *namer) {
	// var stmts []ast.Stmt

	for _, routePkg := range routePkgs {
		groupCount := 0

		routePkgIdent := ident
		routePkgPath := routePkg.routePath()

		imp := ""
		if routeImport, ok := routePackagesImports[routePkg]; ok {
			imp = routeImport.Name.Name
		}

		for _, r := range routePkg.Routes {

			rIdent := routePkgIdent
			rPath := buildRoutePath(routePkgPath, r.Path)

			if r.hasMiddleware() && e.Middleware != nil {
				var groupIdent = ast.NewIdent(n.gen("grp"))
				var groupRoutePath = rPath

				if r.isRootRoute() {
					groupRoutePath = routePkgPath

					// update routePkg
					routePkgIdent = groupIdent
					routePkgPath = ""
				}

				// generate expr with template
				expr := e.GenGroup(rIdent, groupRoutePath)
				buf.WriteString(fmt.Sprintf("%s := %s", groupIdent.Name, expr) + "\n")
				rIdent = groupIdent
				rPath = ""

				groupCount++
				buf.WriteString("{\n")
			}

			callObj := imp

			// sub
			if !r.isRootRoute() {
				typeVar := n.gen(strings.ToLower(r.Name))
				sub := r.Name
				if imp != "" {
					sub = imp + "." + sub
				}
				buf.WriteString(fmt.Sprintf("%s := &%s{}", typeVar, sub) + "\n")
				callObj = typeVar
			}

			// build sels
			for _, sel := range r.Sels {
				var selector = sel
				if callObj != "" {
					selector = callObj + "." + selector
				}
				if expr := e.GenSel(rIdent, sel, rPath, selector); expr != "" {
					buf.WriteString(expr + "\n")
				}
			}
		}

		// sub packages
		if len(routePkg.SubPackages) > 0 {
			g.buildInjectStmts(buf, e, routePkgIdent, routePkg.SubPackages, routePackagesImports, n)
		}

		for i := 0; i < groupCount; i++ {
			buf.WriteString("}\n")
		}
	}
}
