//go:build routegeninject

package router_test

import (
	"github.com/gin-gonic/gin"
	"github.com/serkodev/routegen"
)

func TestBuild(g *gin.Engine) { // gin.RouterGroup
	routegen.Build(g)
}
