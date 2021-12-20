package pbr

import (
	"go/ast"
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
			println("import", i.Path)
		}

		for _, f := range pkg.Syntax {
			for _, decl := range f.Decls {
				fn, ok := decl.(*ast.FuncDecl)
				if !ok {
					continue
				}

				println("found function", fn.Name.Name)
				buildCall, err := findInjectorBuild(pkg.TypesInfo, fn)
				if err != nil {
					println("findInjectorBuild error", err.Error())
					continue
				}
				if buildCall == nil {
					continue
				}

				println("checking args...")
				for _, arg := range buildCall.Args {
					t := pkg.TypesInfo.TypeOf(arg).Underlying().String()
					println(t)
				}
			}
		}
	}

	return nil
}
