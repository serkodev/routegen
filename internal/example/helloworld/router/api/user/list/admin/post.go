package admin

import "github.com/gin-gonic/gin"

func POST(c *gin.Context) {
	println("_id", c)
}

type Action struct {
}

func (*Action) GET(g *gin.Context) {

}