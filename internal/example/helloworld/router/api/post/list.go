package post

import "github.com/gin-gonic/gin"

// pbr:name=list
type List struct {
}

func (*List) POST(g *gin.Context) {

}

func (*List) GET(g *gin.Context) {

}

type Action struct {
}

func (*Action) GET(g *gin.Context) {

}

func GET(g *gin.Context) {
	println("about", g)
}
