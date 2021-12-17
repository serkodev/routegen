package router

import (
	"github.com/serkodev/pbr/route"
	xx "github.com/serkodev/pbr/router/_id"
	"github.com/serkodev/pbr/router/about"
)

//go:generate echo "route"
func Run() {
	r := &route.Route{}
	about.GET(r)
	xx.GET(r)
}
