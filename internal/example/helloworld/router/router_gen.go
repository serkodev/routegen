// Code generated by routegen. DO NOT EDIT.

//go:build !routegeninject
// +build !routegeninject

package router

import (
	"github.com/gin-gonic/gin"
	routegen_r "example.com/foo/router/_id"
	routegen_r2 "example.com/foo/router/about"
	routegen_r3 "example.com/foo/router/api"
	routegen_r4 "example.com/foo/router/api/post"
	routegen_r5 "example.com/foo/router/api/user/_"
	routegen_r6 "example.com/foo/router/api/user/_id"
	routegen_r7 "example.com/foo/router/api/user/list"
	routegen_r8 "example.com/foo/router/api/user/list/admin"
	routegen_r9 "example.com/foo/router/api/yser"
	routegen_r10 "example.com/foo/router/blog"
)

func Build(g *gin.Engine) {
	g.POST("/", POST)
	privateaction := &privateAction{}
	g.GET("/private-action", privateaction.GET)
	g.GET("/:id", routegen_r.GET)
	g.GET("/about", routegen_r2.GET)
	grp := g.Group("/api")
	{
		grp.Use(routegen_r3.Middleware)
		grp.GET("/post", routegen_r4.GET)
		action := &routegen_r4.Action{}
		grp.GET("/post/action", action.GET)
		blist := &routegen_r4.BList{}
		grp.POST("/post/b-list", blist.POST)
		list := &routegen_r4.List{}
		grp.GET("/post/hello", list.GET)
		grp.POST("/post/hello", list.POST)
		grp.GET("/user/*", routegen_r5.GET)
		grp.GET("/user/:id", routegen_r6.GET)
		grp2 := grp.Group("/user/list")
		{
			grp2.Use(routegen_r7.Middleware)
			grp2.GET("", routegen_r7.GET)
			grp2.POST("", routegen_r7.POST)
			grp3 := grp2.Group("/action")
			{
				action2 := &routegen_r7.Action{}
				grp3.Use(action2.Middleware)
				grp3.GET("", action2.GET)
				grp4 := grp2.Group("/admin")
				{
					grp4.Use(routegen_r8.Middleware)
					action3 := &routegen_r8.Action{}
					grp4.GET("/action", action3.GET)
				}
			}
		}
		grp.GET("/yser", routegen_r9.GET)
	}
	g.GET("/blog", routegen_r10.GET)
}

func Run() {
	g := gin.Default()
	Build(g)
	g.Run()
}
