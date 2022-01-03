package pbr

import (
	"go/token"
	"go/types"
)

type namer struct {
	names  map[string]bool
	scopes []*types.Scope
}

func newNamer(scope *types.Scope) *namer {
	return newNamerWithScopes([]*types.Scope{scope})
}

func newNamerWithScopes(scopes []*types.Scope) *namer {
	return &namer{
		names:  make(map[string]bool),
		scopes: scopes,
	}
}

func (n *namer) add(name string) {
	n.names[name] = true
}

func (n *namer) has(name string) bool {
	_, ok := n.names[name]
	return ok
}

func (n *namer) gen(name string) string {
	newName := disambiguate(name, func(s string) bool {
		if n.has(s) {
			return true
		}
		for _, scope := range n.scopes {
			if _, o := scope.LookupParent(s, token.NoPos); o != nil {
				return true
			}
		}
		return false
	})
	n.add(newName)
	return newName
}
