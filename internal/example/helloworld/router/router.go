//go:build pbrinject

package router

import (
	"github.com/gin-gonic/gin"
	"github.com/serkodev/pbr"
)

type bty interface {
}

// allow pbr.Build() call only
/* hi */
func (b *bty) Build(r *gin.Engine) /*hi*/ (bb /*on9*/ *gin.Engine, err error) { // gin.RouterGroup
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
