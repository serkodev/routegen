package post

import "github.com/gin-gonic/gin"

// foo bar comment
// pbr alias=listavd
type List struct {
	// abc
}

// cannot export type define
type MyInt2 = List

func (*List) POST(g *gin.Context) {

}

func (*List) GET(g *gin.Context) {

}

type Action struct {
}

func (*Action) GET(g *gin.Context) {

}

type privateAction struct {
}

func (*privateAction) GET(g *gin.Context) {

}

func GET(g *gin.Context) {
	println("about", g)
}

type BList struct {
}

func (*BList) POST(g *gin.Context) {

}
