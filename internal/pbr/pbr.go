package pbr

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"go/types"
	"io"
	"os"

	"golang.org/x/tools/go/ast/astutil"
	"golang.org/x/tools/go/packages"
)

func Load() error {
	wd, err := os.Getwd()
	if err != nil {
		fmt.Println("failed to get working directory: ", err)
		return err
	}
	fmt.Println(wd)

	parseRoute(wd)

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

		// list import
		// for _, i := range pkg.Types.Imports() {
		// 	fmt.Println("import", i.Path())
		// }

		// for _, f := range pkg.Syntax {
		// 	for _, sx := range f.Decls {
		// 		var buf bytes.Buffer
		// 		if err := printer.Fprint(io.Writer(&buf), pkg.Fset, sx); err != nil {
		// 			panic(err)
		// 		}
		// 		fmt.Println("printer =====")
		// 		fmt.Println(buf.String())
		// 	}
		// }

		for _, f := range pkg.Syntax {
			buildFuncs := make(map[ast.Decl]bool)

			for _, decl := range f.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok {
					continue
				}

				buildCall, err := findInjectorBuild(pkg.TypesInfo, fn)
				if err != nil {
					fmt.Println("findInjectorBuild error", err.Error())
					continue
				}
				if buildCall == nil {
					continue
				}

				// buildFuncs = append(buildFuncs, fn)
				buildFuncs[fn] = true
			}

			if len(buildFuncs) > 0 {
				for _, decl := range f.Decls {
					if buildFuncs[decl] {
						fmt.Println("===== decl =====")
						findInject(pkg, decl.(*ast.FuncDecl), f)
					} else {
						fmt.Println("===== copy =====")
						printAST(pkg.Fset, decl)
					}
				}
			}
		}
	}

	return nil
}

func findInject(pkg *packages.Package, fn *ast.FuncDecl, ff *ast.File) (string, error) {
	header, _ := getFuncHeader(pkg, fn)
	fmt.Println(header)

	// ast.Inspect(fn.Body, func(node ast.Node) bool {
	// 	if node != nil {
	// 		printAST(pkg.Fset, node)
	// 	}
	// 	return true
	// })

	astutil.Apply(fn.Body, func(c *astutil.Cursor) bool {
		if stmt, ok := c.Node().(ast.Stmt); ok {
			if s := getInjectorStmt(pkg.TypesInfo, stmt); s != nil {
				st, err := getExpr(pkg, s.Args[0].(*ast.Ident))
				if err != nil {
					panic("cannot gen expr")
				}
				c.InsertBefore(st)
				c.Delete()
			}

		}
		return true
	}, nil)

	printAST(pkg.Fset, fn.Body)

	return "", nil
}

func inject(pkg *packages.Package, fn *ast.FuncDecl, call *ast.CallExpr) {
	fmt.Println("found injector @ func", fn.Name.Name)

	// fmt.Println("comment", fn.Doc.Text())
	fmt.Println("call pos", pkg.Fset.Position(call.Pos()).Offset, pkg.Fset.Position(call.End()).Offset)

	fmt.Println("checking args...")
	for _, arg := range call.Args {
		fmt.Println("call arg type", pkg.TypesInfo.TypeOf(arg).String())
		fmt.Println("call arg underlying type", pkg.TypesInfo.TypeOf(arg).Underlying().String())

		fmt.Printf("arg: %T", arg)

		o := qualifiedIdentObject(pkg.TypesInfo, arg)
		fmt.Printf("o: %s", o.Type().String())
	}

	if fn.Doc != nil {
		fmt.Println("fn pos (comment)", pkg.Fset.Position(fn.Doc.Pos()).Offset)
	} else {
		fmt.Println("fn pos", pkg.Fset.Position(fn.Pos()).Offset)
	}

	header, _ := getFuncHeader(pkg, fn)
	fmt.Println(header)

	printAST(pkg.Fset, fn)

	// fn.Type.Params
	// astutil.Apply()

	fmt.Println("body.list")
	for _, stmt := range fn.Body.List {
		fmt.Println("===")
		printAST(pkg.Fset, stmt)
	}

	// fn.Type.Params
}

func printAST(fset *token.FileSet, node interface{}) string {
	// print function, ref: wire.go writeAST rewritePkgRefs
	var buf bytes.Buffer
	if err := format.Node(io.Writer(&buf), fset, node); err != nil {
		panic(err)
	}
	s := buf.String()
	fmt.Println(s)
	return s
}

func printType(typ types.Type, q types.Qualifier) {
	var buf bytes.Buffer
	types.WriteType(&buf, typ, q)
	fmt.Println(buf.String())
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
