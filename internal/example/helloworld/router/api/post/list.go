package post

import "github.com/gin-gonic/gin"

type list struct {
	Name string
}

type _list struct {
}

type _List struct {
}

// pbr:name=list
type List struct {
}

func (l *list) GET(r string) {

}

func (*list) POST(r string) {

}

func (_list) GET(r string) {

}

func GET(g *gin.Context) {
	println("about", g)
}
