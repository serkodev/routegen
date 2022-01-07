// Code generated by pbr. DO NOT EDIT.

//go:build !pbrinject
// +build !pbrinject

package router

import (
	"github.com/gin-gonic/gin"
	pbr_route "example.com/foo/router/_id"
	pbr_route2 "example.com/foo/router/about"
	pbr_route3 "example.com/foo/router/api"
	pbr_route4 "example.com/foo/router/api/post"
	pbr_route5 "example.com/foo/router/api/user/_"
	pbr_route6 "example.com/foo/router/api/user/_id"
	pbr_route7 "example.com/foo/router/api/user/list"
	pbr_route8 "example.com/foo/router/api/user/list/admin"
	pbr_route9 "example.com/foo/router/api/yser"
	pbr_route10 "example.com/foo/router/blog"
)

func Build(grp *gin.Engine) {
	grp.POST("/", POST)
	privateaction := &privateAction{}
	grp.GET("/private-action", privateaction.GET)
	grp.GET("/:id", pbr_route.GET)
	grp.GET("/about", pbr_route2.GET)
	grp2 := grp.Group("/api")
	grp2.Use(pbr_route3.Middleware)
	grp2.GET("/post", pbr_route4.GET)
	action := &pbr_route4.Action{}
	grp2.GET("/post/action", action.GET)
	blist := &pbr_route4.BList{}
	grp2.POST("/post/b-list", blist.POST)
	list := &pbr_route4.List{}
	grp2.GET("/post/hello", list.GET)
	grp2.POST("/post/hello", list.POST)
	grp2.GET("/user/*", pbr_route5.GET)
	grp2.GET("/user/:id", pbr_route6.GET)
	grp3 := grp2.Group("/user/list")
	grp3.Use(pbr_route7.Middleware)
	grp3.GET("", pbr_route7.GET)
	grp3.POST("", pbr_route7.POST)
	grp4 := grp3.Group("/action")
	action2 := &pbr_route7.Action{}
	grp4.Use(action2.Middleware)
	grp4.GET("", action2.GET)
	grp5 := grp3.Group("/admin")
	grp5.Use(pbr_route8.Middleware)
	action3 := &pbr_route8.Action{}
	grp5.GET("/action", action3.GET)
	grp2.GET("/yser", pbr_route9.GET)
	grp.GET("/blog", pbr_route10.GET)
}

func Run() {
	g := gin.Default()
	Build(g)
	g.Run()
}
