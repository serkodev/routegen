//go:build routegeninject

package router

import (
	"github.com/gin-gonic/gin"
	"github.com/serkodev/routegen"
)

func Build(g *gin.Engine) { // gin.RouterGroup
	routegen.Build(g)
}

func Run() {
	g := gin.Default()
	Build(g)
	g.Run()
}
