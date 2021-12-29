package pbr

import (
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

type gen struct {
	pkg       *packages.Package
	buildFunc map[*ast.FuncDecl]bool
}

func newGen(pkg *packages.Package, buildFunc map[*ast.FuncDecl]bool) *gen {
	return &gen{
		pkg:       pkg,
		buildFunc: buildFunc,
	}
}

func Load() error {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("failed to get working directory: ", err)
		return err
	}
	fmt.Println(wd)

	cfg := &packages.Config{
		// Context:    ctx,
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedCompiledGoFiles | // LoadFiles
			packages.NeedImports | // LoadImports
			packages.NeedTypes | packages.NeedTypesSizes | // LoadTypes
			packages.NeedSyntax | packages.NeedTypesInfo | // LoadSyntax
			packages.NeedDeps, // LoadTypes
		Dir: wd,
		// Env:        env,
		BuildFlags: []string{"-tags=pbrinject"},
		// TODO(light): Use ParseFile to skip function bodies and comments in indirect packages.
	}

	pkgs, err := packages.Load(cfg)
	if err != nil {
		return err
	}

	for _, pkg := range pkgs {
		fmt.Println("path", pkg.PkgPath)

		for _, f := range pkg.Syntax {
			buildFuncs := make(map[*ast.FuncDecl]bool)

			ast.Inspect(f, func(n ast.Node) bool {
				fn, ok := n.(*ast.FuncDecl)
				if !ok {
					return true
				}
				if buildCall, err := findInjectorBuild(pkg.TypesInfo, fn); err != nil {
					fmt.Println("findInjectorBuild error", err.Error())
					return true
				} else if buildCall != nil {
					// inject build found
					buildFuncs[fn] = true
				}
				return true
			})

			if len(buildFuncs) > 0 {
				r := newRouteGen()
				routes := r.parseRoute(wd)

				g := newGen(pkg, buildFuncs)
				g.inject(f, routes)
			}
		}
	}

	return nil
}

func (g *gen) canUseImportName(name string) bool {
	for fn := range g.buildFunc {
		_, o := g.pkg.TypesInfo.Scopes[fn.Type].LookupParent(name, token.NoPos)
		if o != nil {
			return false
		}
	}
	return true
}

func (g *gen) generate(f *ast.File) error {
	oriFilePath := g.pkg.Fset.File(f.Pos()).Name()

	filename := filepath.Base(oriFilePath)
	ext := strings.ToLower(filepath.Ext(filename))
	nonExt := filename[:len(filename)-len(ext)]
	dir := filepath.Dir(oriFilePath)
	outputPath := dir + "/" + nonExt + "_gen" + ext

	out, err := os.Create(outputPath)
	if err != nil {
		return err
	}
	defer out.Close()

	out.WriteString("// Code generated by pbr. DO NOT EDIT.\n")
	out.WriteString("\n")
	out.WriteString("//go:build !pbrinject\n")
	out.WriteString("//+build !pbrinject\n")
	out.WriteString("\n")
	out.WriteString("package " + g.pkg.Name + "\n")
	out.WriteString("\n")

	for i, decl := range f.Decls {
		printAST(token.NewFileSet(), decl)
		fmt.Print("\n")

		if err := printer.Fprint(out, token.NewFileSet(), decl); err != nil {
			return err
		}

		if i < len(f.Decls)-1 {
			out.WriteString("\n\n")
		} else {
			out.WriteString("\n") // EOF
		}
	}

	return nil
}

func (g *gen) importsFromRoutes(routes []*RoutePackage) []*ast.ImportSpec {
	imports := make([]*ast.ImportSpec, 0, len(routes))

	newNames := make(map[string]bool)
	inNewNames := func(n string) bool {
		_, ok := newNames[n]
		return ok
	}

	for _, r := range routes {
		newName := disambiguate("pbr_route", func(s string) bool {
			if !g.canUseImportName(s) || inNewNames(s) {
				return true
			}
			return false
		})
		newNames[newName] = true

		imports = append(imports, &ast.ImportSpec{
			Path: &ast.BasicLit{
				Kind:  token.STRING,
				Value: `"` + r.PkgPath + `"`,
			},
			Name: ast.NewIdent(newName),
		})

	}
	return imports
}

func (g *gen) inject(f *ast.File, routes []*RoutePackage) {
	// TODO: check original imported pkg
	imps := g.importsFromRoutes(routes)
	astutil.Apply(f, func(c *astutil.Cursor) bool {
		// inject imports
		if imp, ok := c.Node().(*ast.ImportSpec); ok {
			if imp.Path.Value == `"github.com/serkodev/pbr"` {
				for i := len(imps) - 1; i >= 0; i-- {
					c.InsertAfter(imps[i])
				}
				c.Delete()
			}
		} else if fn, ok := c.Node().(*ast.FuncDecl); ok && g.buildFunc[fn] {
			g.injectFunction(fn, imps, routes)
		}
		return true
	}, nil)

	g.generate(f)
}

func (g *gen) injectFunction(fn *ast.FuncDecl, imps []*ast.ImportSpec, routes []*RoutePackage) error {
	astutil.Apply(fn.Body, func(c *astutil.Cursor) bool {
		if stmt, ok := c.Node().(ast.Stmt); ok {
			if s := getInjectorStmt(g.pkg.TypesInfo, stmt); s != nil {
				ident := s.Args[0].(*ast.Ident)
				for i, route := range routes {
					for _, stmt := range injectPkgRoute(ident, imps[i].Name.Name, route) {
						c.InsertBefore(stmt)
					}
				}
				c.Delete()
			}
		}
		return true
	}, nil)
	// printAST(token.NewFileSet(), fn.Body)
	return nil
}

func injectPkgRoute(ident *ast.Ident, pkgImportName string, route *RoutePackage) []*ast.ExprStmt {
	var stmts []*ast.ExprStmt
	for k, sels := range route.Handles {
		if k == "" {
			for _, sel := range sels {
				expr, _ := parseExpr(fmt.Sprintf(`%s.%s("%s", %s.%s)`, ident.Name, sel, route.RelativePath, pkgImportName, sel))
				stmts = append(stmts, expr)
			}
		}
		// TODO: (sub route) else inject sel with func recv
	}
	return stmts
}
