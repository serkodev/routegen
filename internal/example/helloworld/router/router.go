//go:build pbrinject

package router

import (
	"github.com/gin-gonic/gin"
	"github.com/serkodev/pbr"
)

func Build(grp *gin.Engine) { // gin.RouterGroup
	pbr.Build(grp)
}

func Run() {
	g := gin.Default()
	Build(g)
	g.Run()
}
