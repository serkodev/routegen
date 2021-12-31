// Code generated by pbr. DO NOT EDIT.

//go:build !pbrinject
//+build !pbrinject

package router

import (
	"github.com/gin-gonic/gin"
	pbr_route "example.com/foo/router/_id"
	pbr_route2 "example.com/foo/router/about"
	pbr_route3 "example.com/foo/router/api/post"
	pbr_route4 "example.com/foo/router/api/user/_id"
	pbr_route5 "example.com/foo/router/api/user/list"
)

func Build(g *gin.Engine) {
	g.POST(".", POST)
	privateaction := &privateAction{}
	g.GET(".", privateaction.GET)
	g.GET("_id", pbr_route.GET)
	g.GET("about", pbr_route2.GET)
	list := &pbr_route3.List{}
	g.POST("api/post", list.POST)
	g.GET("api/post", list.GET)
	action := &pbr_route3.Action{}
	g.GET("api/post", action.GET)
	g.GET("api/post", pbr_route3.GET)
	g.GET("api/user/_id", pbr_route4.GET)
	g.GET("api/user/list", pbr_route5.GET)
	g.POST("api/user/list", pbr_route5.POST)
	action2 := &pbr_route5.Action{}
	g.GET("api/user/list", action2.GET)
}

func Run() {
	b := &gin.Engine{}
	println(b)
	rrrrrr := gin.Default()
	Build(rrrrrr)
	rrrrrr.Run()
}
