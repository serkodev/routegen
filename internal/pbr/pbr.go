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

func newGen(pkg *packages.Package) *gen {
	return &gen{
		pkg: pkg,
	}
}

func Load() error {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("failed to get working directory: ", err)
		return err
	}
	fmt.Println(wd)

	r := newRouteGen()
	_ = r.parseRoute(wd)

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

			astutil.Apply(f, func(c *astutil.Cursor) bool {
				fn, ok := c.Node().(*ast.FuncDecl)
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
			}, nil)

			if len(buildFuncs) > 0 {
				fmt.Println("found pbr.Build", pkg.Fset.File(f.Pos()).Name())
				g := newGen(pkg)
				g.buildFunc = buildFuncs
				g.inject(f)
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
		printAST(g.pkg.Fset, decl)
		fmt.Print("\n")
	}
}

func (g *gen) inject(f *ast.File) {
	// TODO: import fmt.Println("r", g.canUseImportName("r"))

	astutil.Apply(f, func(c *astutil.Cursor) bool {
		if _, ok := c.Node().(*ast.ImportSpec); ok {

			//if gen.Tok == token.IMPORT {
			//	fmt.Println("import token")
			//	}
			c.InsertAfter(&ast.ImportSpec{
				Path: &ast.BasicLit{
					Kind:  token.STRING,
					Value: "\"abc\"",
				},
				Name: ast.NewIdent("wtf"),
			})
		} else if fn, ok := c.Node().(*ast.FuncDecl); ok && g.buildFunc[fn] {
			fmt.Println("===== decl =====")
			g.injectFunction(fn)
		}
		return true
	}, nil)

	g.generate(f)
}

func (g *gen) injectFunction(fn *ast.FuncDecl) (string, error) {
	header, _ := getFuncHeader(g.pkg, fn)
	fmt.Println(header)

	astutil.Apply(fn.Body, func(c *astutil.Cursor) bool {
		if stmt, ok := c.Node().(ast.Stmt); ok {
			if s := getInjectorStmt(g.pkg.TypesInfo, stmt); s != nil {
				st, err := getExpr(g.pkg, s.Args[0].(*ast.Ident))
				if err != nil {
					panic("cannot gen expr")
				}
				c.InsertBefore(st)
				c.Delete()
			}

		}
		return true
	}, nil)

	printAST(g.pkg.Fset, fn.Body)

	return "", nil
}

// TODO: input format
func getExpr(pkg *packages.Package, ident *ast.Ident) (*ast.ExprStmt, error) {
	// c, err := parser.ParseExpr(`r.bar("baz",foo(bar),struct{}{abc: 123})`)
	c, err := parser.ParseExprFrom(pkg.Fset, "", []byte(ident.Name+`.bar("baz", foo(bar))`), 0)
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
