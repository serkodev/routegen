package pbr

import (
	"bytes"
	"fmt"
	"go/printer"
	"go/token"
	"go/types"
	"io"
	"strconv"
)

func printAST(fset *token.FileSet, node interface{}) string {
	// print function, ref: wire.go writeAST rewritePkgRefs
	var buf bytes.Buffer
	if err := printer.Fprint(io.Writer(&buf), fset, node); err != nil {
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
