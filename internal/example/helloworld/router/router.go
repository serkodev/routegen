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
	g := gin.Default()
	Build(g)
	g.Run()
}
