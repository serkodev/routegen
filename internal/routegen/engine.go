package routegen

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"go/ast"
	"go/types"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
)

//go:embed engineconfig/gin.json
var ginJSON []byte

//go:embed engineconfig/echo.json
var echoJSON []byte

type engineManager struct {
	engines []*engine
}

func newEngineManager(customEngine string) (*engineManager, error) {
	engines := []*engine{}

	if customEngine == "" {
		if _, err := os.Stat("./routegen.json"); err == nil {
			customEngine = "./routegen.json"
		}
	}

	// load custom engine
	if customEngine != "" {
		f, err := os.Open(customEngine)
		if err != nil {
			return nil, err
		}
		defer f.Close()

		content, err := ioutil.ReadAll(f)
		if err != nil {
			return nil, err
		}

		engine, err := newEngine(content)
		if err != nil {
			return nil, err
		}
		engines = append(engines, engine)
		log.Println("custom engine loaded:", customEngine)
	}

	engines = append(engines,
		mustNewEngine(ginJSON),
		mustNewEngine(echoJSON),
	)

	return &engineManager{
		engines: engines,
	}, nil
}

func (m *engineManager) matchEngine(obj types.Object) *engine {
	for _, e := range m.engines {
		if e.ValidInjectType(obj.Type()) {
			return e
		}
	}
	return nil
}

type middleware struct {
	Selector  string `json:"selector"`
	GroupExpr string `json:"group_expr"`
	template  *template.Template
}

type engine struct {
	Types        []string          `json:"types"`
	Selectors    []string          `json:"selectors"`
	Expr         map[string]string `json:"expr"`
	Middleware   *middleware       `json:"middleware"`
	exprTemplate map[string]*template.Template
}

func mustNewEngine(data []byte) *engine {
	if e, err := newEngine(data); err != nil {
		panic("cannot parse engine")
	} else {
		return e
	}
}

func newEngine(data []byte) (*engine, error) {
	var e *engine
	if err := json.Unmarshal(data, &e); err != nil {
		return nil, err
	}

	if m := e.Middleware; m != nil {
		t, err := template.New("").Parse(m.GroupExpr)
		if err != nil {
			return nil, err
		}
		m.template = t
	}

	if len(e.Expr) > 0 {
		e.exprTemplate = make(map[string]*template.Template)
		for k, expr := range e.Expr {
			t, err := template.New("").Parse(expr)
			if err != nil {
				return nil, err
			}
			e.exprTemplate[k] = t
		}
	}

	return e, nil
}

func (e *engine) ValidInjectType(t types.Type) bool {
	if len(e.Types) == 0 {
		return true
	}
	inType := t.String()
	for _, t := range e.Types {
		if t == inType {
			return true
		}
	}
	return false
}

func (e *engine) TargetSels() []string {
	sels := e.Selectors
	if m := e.Middleware; m != nil {
		sels = append([]string{m.Selector}, sels...)
	}
	return sels
}

func (e *engine) MiddlewareSelector() string {
	if m := e.Middleware; m != nil {
		return m.Selector
	}
	return ""
}

func (e *engine) GenGroup(i *ast.Ident, route string) string {
	var expr bytes.Buffer
	if err := e.Middleware.template.Execute(io.Writer(&expr), map[string]string{
		"ident": i.Name,
		"route": route,
	}); err != nil {
		panic("generate expr error")
	}
	return expr.String()
}

func (e *engine) GenSel(i *ast.Ident, sel string, route string, handle string) string {
	t, ok := e.exprTemplate[sel]
	if !ok {
		t, ok = e.exprTemplate["_default"]
		if !ok {
			panic("not match any selector")
		}
	}

	var expr bytes.Buffer
	if err := t.Execute(io.Writer(&expr), map[string]string{
		"ident":  i.Name,
		"sel":    sel,
		"route":  route,
		"handle": handle,
	}); err != nil {
		panic("generate expr error")
	}
	return expr.String()
}
