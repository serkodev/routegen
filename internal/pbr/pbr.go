package pbr

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/printer"
	"go/token"
	"go/types"
	"io"
	"log"
	"os"

	"golang.org/x/tools/go/packages"
)

func Load() error {
	wd, err := os.Getwd()
	if err != nil {
		log.Println("failed to get working directory: ", err)
		return err
	}
	println(wd)

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

	pkgs, err := packages.Load(cfg, ".")
	if err != nil {
		return err
	}

	for _, pkg := range pkgs {
		println("path", pkg.PkgPath)

		// list import
		// for _, i := range pkg.Types.Imports() {
		// 	println("import", i.Path())
		// }

		// for _, f := range pkg.Syntax {
		// 	for _, sx := range f.Decls {
		// 		var buf bytes.Buffer
		// 		if err := printer.Fprint(io.Writer(&buf), pkg.Fset, sx); err != nil {
		// 			panic(err)
		// 		}
		// 		println("printer =====")
		// 		println(buf.String())
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
					println("findInjectorBuild error", err.Error())
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
						println("===== decl =====")
						findInject(pkg, decl.(*ast.FuncDecl))
					} else {
						println("===== copy =====")
						printAST(pkg.Fset, decl)
					}
				}
			}
		}
	}

	return nil
}

func findInject(pkg *packages.Package, fn *ast.FuncDecl) (string, error) {
	header, _ := getFuncHeader(pkg, fn)
	println(header, "{")

	for _, stmt := range fn.Body.List {
		if ij := getInjectorStmt(pkg.TypesInfo, stmt); ij != nil {
			println("=== !! replace !! ===")
		} else {
			printAST(pkg.Fset, stmt)
		}
	}
	println("}")

	return "", nil
}

func inject(pkg *packages.Package, fn *ast.FuncDecl, call *ast.CallExpr) {

	println("found injector @ func", fn.Name.Name)

	// println("comment", fn.Doc.Text())
	println("call pos", pkg.Fset.Position(call.Pos()).Offset, pkg.Fset.Position(call.End()).Offset)

	println("checking args...")
	for _, arg := range call.Args {
		println("call arg type", pkg.TypesInfo.TypeOf(arg).String())
		println("call arg underlying type", pkg.TypesInfo.TypeOf(arg).Underlying().String())

		fmt.Printf("arg: %T", arg)

		o := qualifiedIdentObject(pkg.TypesInfo, arg)
		fmt.Printf("o: %s", o.Type().String())
	}

	if fn.Doc != nil {
		println("fn pos (comment)", pkg.Fset.Position(fn.Doc.Pos()).Offset)
	} else {
		println("fn pos", pkg.Fset.Position(fn.Pos()).Offset)
	}

	header, _ := getFuncHeader(pkg, fn)
	println(header)

	printAST(pkg.Fset, fn)

	// fn.Type.Params
	// astutil.Apply()

	println("body.list")
	for _, stmt := range fn.Body.List {
		println("===")
		printAST(pkg.Fset, stmt)
	}

	// fn.Type.Params
}

func printAST(fset *token.FileSet, node interface{}) {
	// print function, ref: wire.go writeAST rewritePkgRefs
	var buf bytes.Buffer
	if err := printer.Fprint(io.Writer(&buf), fset, node); err != nil {
		panic(err)
	}
	println(buf.String())
}

func printType(typ types.Type, q types.Qualifier) {
	var buf bytes.Buffer
	types.WriteType(&buf, typ, q)
	println(buf.String())
}

func getFuncHeader(pkg *packages.Package, fn *ast.FuncDecl) (string, error) {
	// format: func Recv? > Name > Param > Results? { Body }

	// receiver (or nil)
	var recv string
	if fn.Recv != nil {
		if r, err := readTokenFromPkgFile(pkg, fn.Recv.Pos(), fn.Recv.End()); err == nil {
			recv = r + " "
		}
	}

	// params (non-nil)
	params, _ := readTokenFromPkgFile(pkg, fn.Type.Params.Pos(), fn.Type.Params.End())

	// results (or nil)
	var results string
	if fn.Type.Results != nil {
		if r, err := readTokenFromPkgFile(pkg, fn.Type.Results.Pos(), fn.Type.Results.End()); err == nil {
			results = " " + r
		}
	}

	return fmt.Sprintf("func %s%s%s%s", recv, fn.Name.Name, params, results), nil
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
	buf := make([]byte, slen)
	gf.Seek(int64(p.Offset), 0)
	if n, err := gf.Read(buf); err != nil {
		return "", err
	} else if n != slen {
		return "", fmt.Errorf("read token error")
	}

	return string(buf), nil
}
