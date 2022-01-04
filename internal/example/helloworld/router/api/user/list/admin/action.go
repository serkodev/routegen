package admin

import "github.com/gin-gonic/gin"

type Action struct {
}

func Middleware(c *gin.Context) {
	println("_id", c)
}

func (*Action) GET(g *gin.Context) {

}
