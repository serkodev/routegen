package pbr

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"io"
	"strconv"
)

func parseExpr(expr string) (ast.Stmt, error) {
	c, err := parser.ParseExpr(expr)
	if err != nil {
		if e, err := parser.ParseExpr("func(){\n\t" + expr + "\n}"); err == nil {
			node := e.(*ast.FuncLit).Body.List[0]
			if node == nil {
				return nil, fmt.Errorf("%T not supported", node)
			} else {
				return node, nil
			}
		}
		return nil, err
	}
	return &ast.ExprStmt{X: c}, nil
}

func printAST(fset *token.FileSet, node interface{}) string {
	var buf bytes.Buffer
	if err := printer.Fprint(io.Writer(&buf), fset, node); err != nil {
		panic(err)
	}
	s := buf.String()
	fmt.Println(s)
	return s
}

// disambiguate picks a unique name, preferring name if it is already unique.
// It also disambiguates against Go's reserved keywords.
func disambiguate(name string, collides func(string) bool) string {
	if !token.Lookup(name).IsKeyword() && !collides(name) {
		return name
	}
	buf := []byte(name)
	if len(buf) > 0 && buf[len(buf)-1] >= '0' && buf[len(buf)-1] <= '9' {
		buf = append(buf, '_')
	}
	base := len(buf)
	for n := 2; ; n++ {
		buf = strconv.AppendInt(buf[:base], int64(n), 10)
		sbuf := string(buf)
		if !token.Lookup(sbuf).IsKeyword() && !collides(sbuf) {
			return sbuf
		}
	}
}
