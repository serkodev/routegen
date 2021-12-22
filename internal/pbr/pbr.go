package pbr

import (
	"bytes"
	"go/ast"
	"go/printer"
	"go/token"
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
		for _, i := range pkg.Types.Imports() {
			println("import", i.Path())
		}

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

				println("found injector @ func", fn.Name.Name)

				// func args
				for _, fnArg := range fn.Type.Params.List {
					fao := pkg.TypesInfo.ObjectOf(fnArg.Names[0]) // TODO: check when will Names length > 0

					// TODO: check parent type
					// pkg.TypesInfo.TypeOf(fnArg.Type)
					println("fnarg")
					println(fao.Pkg().Name())
					println(fao.Type().String())
					println(pkg.TypesInfo.TypeOf(fnArg.Type).String())
				}

				// println("comment", fn.Doc.Text())
				println("call pos", pkg.Fset.Position(buildCall.Pos()).Offset, pkg.Fset.Position(buildCall.End()).Offset)

				println("checking args...")
				for _, arg := range buildCall.Args {
					println("call arg type", pkg.TypesInfo.TypeOf(arg).String())
					println("call arg underlying type", pkg.TypesInfo.TypeOf(arg).Underlying().String())
				}

				if fn.Doc != nil {
					println("fn pos (comment)", pkg.Fset.Position(fn.Doc.Pos()).Offset)
				} else {
					println("fn pos", pkg.Fset.Position(fn.Pos()).Offset)
				}

				printAST(pkg.Fset, fn)

				println("body.list")
				for _, stmt := range fn.Body.List {
					println("===")
					printAST(pkg.Fset, stmt)
				}

				// fn.Type.Params
			}
		}
	}

	return nil
}

func printAST(fset *token.FileSet, node interface{}) {
	// print function, ref: wire.go writeAST rewritePkgRefs
	var buf bytes.Buffer
	if err := printer.Fprint(io.Writer(&buf), fset, node); err != nil {
		panic(err)
	}
	println(buf.String())
}
