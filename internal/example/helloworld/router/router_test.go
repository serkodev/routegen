//go:build pbrinject

package router_test

import (
	"github.com/gin-gonic/gin"
	"github.com/serkodev/pbr"
)

func TestBuild(g *gin.Engine) { // gin.RouterGroup
	pbr.Build(g)
}
