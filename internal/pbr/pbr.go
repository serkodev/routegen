package pbr

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"

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

				fmt.Println("found pbr.Build", pkg.Fset.File(f.Pos()).Name())
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

func (g *gen) generate(f *ast.File) {
	fmt.Println("generate", g.pkg.Fset.File(f.Pos()).Name())

	// printAST(g.pkg.Fset, f)
	for _, decl := range f.Decls {
		printAST(token.NewFileSet(), decl)
		fmt.Print("\n")
	}
}

func (g *gen) importsFromRoutes(routes []*RoutePackage) []*ast.ImportSpec {
	imports := make([]*ast.ImportSpec, 0, len(routes))

	newNames := make(map[string]bool)
	inNewNames := func(n string) bool {
		_, ok := newNames[n]
		return ok
	}

	for _, r := range routes {
		pkgPath := r.PkgPath
		fmt.Println(pkgPath, r.RelativePath)
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
				Value: `"` + pkgPath + `"`,
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

func (g *gen) injectFunction(fn *ast.FuncDecl, imps []*ast.ImportSpec, routes []*RoutePackage) (string, error) {
	header, _ := getFuncHeader(g.pkg, fn)
	fmt.Println(header)

	astutil.Apply(fn.Body, func(c *astutil.Cursor) bool {
		if stmt, ok := c.Node().(ast.Stmt); ok {
			if s := getInjectorStmt(g.pkg.TypesInfo, stmt); s != nil {

				// // ident *ast.Ident
				// st, err := getExpr(s.Args[0].(*ast.Ident).Name + `.bar("baz", foo(bar))`)
				// if err != nil {
				// 	panic("cannot gen expr")
				// }
				// c.InsertBefore(st)

				ident := s.Args[0].(*ast.Ident)
				for i, route := range routes {
					for _, stmt := range injectPkgRoute(ident, imps[i].Name.Name, route) {
						c.InsertBefore(stmt)
					}
				}

				// _ = st
				c.Delete()
			}

		}
		return true
	}, nil)

	// fmt.Println("===== decl =====")
	// printAST(token.NewFileSet(), fn.Body)

	return "", nil
}

// TODO: sub route
func injectPkgRoute(ident *ast.Ident, pkgImportName string, route *RoutePackage) []*ast.ExprStmt {
	var stmts []*ast.ExprStmt
	for k, sels := range route.Handles {
		if k == "" {
			for _, sel := range sels {
				expr, _ := getExpr(fmt.Sprintf(`%s.%s("%s", %s.%s)`, ident.Name, sel, route.RelativePath, pkgImportName, sel))
				stmts = append(stmts, expr)
			}
		}
	}
	return stmts
}

// TODO: input format
func getExpr(expr string) (*ast.ExprStmt, error) {
	c, err := parser.ParseExpr(expr) // `r.bar("baz",foo(bar),struct{}{abc: 123})`
	// c, err := parser.ParseExprFrom(pkg.Fset, "", []byte(ident.Name+`.bar("baz", foo(bar))`), 0)
	if err != nil {
		return nil, err
	}
	return &ast.ExprStmt{X: c}, nil
}

func getFuncHeader(pkg *packages.Package, fn *ast.FuncDecl) (string, error) {
	// format: func Recv? > Name > Param > Results? { Body }
	if fn.Type.Func == token.NoPos {
		return "", fmt.Errorf("cannot generate func header, invalid Type.Func")
	}
	return readTokenFromPkgFile(pkg, fn.Type.Func, fn.Type.End())
}

// TODO: need improve performance, reading from file maybe is not a good idea
func readTokenFromPkgFile(pkg *packages.Package, pos token.Pos, end token.Pos) (string, error) {
	f := pkg.Fset.File(pos)
	p := f.Position(pos)
	e := f.Position(end)

	gf, err := os.Open(f.Name())
	if err != nil {
		return "", err
	}
	defer gf.Close()

	// seek pos and read go file
	slen := e.Offset - p.Offset
	if slen < 0 {
		return "", fmt.Errorf("invalid token pos")
	}
	buf := make([]byte, slen)
	if _, err = gf.Seek(int64(p.Offset), 0); err != nil {
		return "", err
	}
	if n, err := gf.Read(buf); err != nil {
		return "", err
	} else if n != slen {
		return "", fmt.Errorf("read token error")
	}
	return string(buf), nil
}
