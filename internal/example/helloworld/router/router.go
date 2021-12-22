//go:build pbrinject

package router

import (
	"github.com/gin-gonic/gin"
	"github.com/serkodev/pbr"
)

// gin.RouterGroup
func Build(r *gin.Engine) {
	pbr.Build(r)
}

func Run() {
	rrrrrr := gin.Default()

	// r := &route.Route{}
	// about.GET(r)
	// xx.GET(r)

	Build(rrrrrr)

	rrrrrr.Run()
}
