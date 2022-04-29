package router

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func tss() {
	fmt.Println("hi")
}

func POST(c *gin.Context) {
}

type privateAction struct {
}

func (*privateAction) GET(g *gin.Context) {

}
