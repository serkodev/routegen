package blog

import "github.com/gin-gonic/gin"

func GET(c *gin.Context) {
	println("_id", c)
}
