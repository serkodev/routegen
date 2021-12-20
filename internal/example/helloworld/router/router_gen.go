//go:build !pbrinject
// +build !pbrinject

package router

import (
	xx "example.com/foo/router/_id"
	"example.com/foo/router/about"
	"github.com/gin-gonic/gin"
)

func Run() {
	router := gin.Default()

	about.GET("123")
	xx.GET("123")

	router.Run()
}
