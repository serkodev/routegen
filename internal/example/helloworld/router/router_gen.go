//go:build !pbrinject
// +build !pbrinject

package router

import (
	"github.com/gin-gonic/gin"
)

// TODO: route callback
func Build(r *gin.Engine) {
	r.GET("about/a")
	r.GET("about/b")
}

func Run() {
	router := gin.Default()
	Build(router)
	router.Run()
}
