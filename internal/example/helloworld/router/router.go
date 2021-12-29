//go:build pbrinject

package router

import (
	"github.com/gin-gonic/gin"
	"github.com/serkodev/pbr"
)

func Build(g *gin.Engine) { // gin.RouterGroup
	pbr.Build(g)
}

func Run() {
	rrrrrr := gin.Default()

	// r := &route.Route{}
	// about.GET(r)
	// xx.GET(r)

	Build(rrrrrr)

	rrrrrr.Run()
}
