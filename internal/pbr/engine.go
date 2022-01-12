package pbr

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"go/ast"
	"go/types"
	"html/template"
	"io"
)

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

//go:embed engineconfig/gin.json
var ginJSON []byte

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

type engineManager struct {
	engines []*engine
}

func newEngineManager() *engineManager {
	e, _ := newEngine(ginJSON)
	return &engineManager{
		engines: []*engine{
			e,
		},
	}
}

func (m *engineManager) matchEngine(obj types.Object) *engine {
	for _, e := range m.engines {
		if e.ValidInjectType(obj.Type()) {
			return e
		}
	}
	return nil
}
