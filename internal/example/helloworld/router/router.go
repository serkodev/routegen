//go:build pbrinject

package router

import (
	"github.com/gin-gonic/gin"
	"github.com/serkodev/pbr"
)

type bty interface {
}

func Build(r *gin.Engine) { // gin.RouterGroup
	pbr.Build(r)
	println("hihi")
}

func Run() {
	rrrrrr := gin.Default()

	// r := &route.Route{}
	// about.GET(r)
	// xx.GET(r)

	Build(rrrrrr, nil, nil)

	rrrrrr.Run()
}
