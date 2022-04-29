package api

import "github.com/gin-gonic/gin"

func Middleware(c *gin.Context) {
	println("_id", c)
}
