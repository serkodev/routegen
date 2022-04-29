package list

import "github.com/gin-gonic/gin"

func POST(c *gin.Context) {
	println("_id", c)
}

type Action struct {
}

func (*Action) Middleware(g *gin.Context) {

}

func (*Action) GET(g *gin.Context) {

}
