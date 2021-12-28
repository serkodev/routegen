//go:build !pbrinject
// +build !pbrinject

package router

import (
	xx1 "example.com/foo/router/api/post"
	"github.com/gin-gonic/gin"
)

// var B = 2

// TODO: route callback
func Build(xx *gin.Engine) {
	xx.GET("about/a")
	xx.GET("about/b")
	xx.GET("abc", xx1.GET)
}

func Run() {
	router := gin.Default()
	Build(router)
	router.Run()
	_ = xx1.GET
}
