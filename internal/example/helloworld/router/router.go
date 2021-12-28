//go:build pbrinject

package router

import (
	xx "example.com/foo/router/api/post"

	"github.com/gin-gonic/gin"
	"github.com/serkodev/pbr"
)

type bty interface {
}

func Build(g *gin.Engine) { // gin.RouterGroup
	pbr.Build(g)
}

func Run() {
	rrrrrr := gin.Default()

	// r := &route.Route{}
	// about.GET(r)
	// xx.GET(r)

	Build(rrrrrr, nil, nil)

	rrrrrr.Run()
}
