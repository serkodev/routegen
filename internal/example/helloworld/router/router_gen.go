//go:build !pbrinject
// +build !pbrinject

package router

import (
	pbr_route "example.com/foo/router/_id"
	pbr_route2 "example.com/foo/router/about"
	pbr_route3 "example.com/foo/router/api/post"
	pbr_route4 "example.com/foo/router/api/user/_id"
	pbr_route5 "example.com/foo/router/api/user/list"
	"github.com/gin-gonic/gin"
)

func Build(g *gin.Engine) {
	g.GET("_id", pbr_route.GET)
	g.GET("about", pbr_route2.GET)
	g.GET("api/post", pbr_route3.GET)
	g.GET("api/user/_id", pbr_route4.GET)
	g.GET("api/user/list", pbr_route5.GET)
	g.POST("api/user/list", pbr_route5.POST)
}

func Run() {
	rrrrrr := gin.Default()
	Build(rrrrrr)
	rrrrrr.Run()
}
