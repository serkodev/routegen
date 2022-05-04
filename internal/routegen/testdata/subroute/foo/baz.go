package foo

import "github.com/gin-gonic/gin"

type Baz struct{}

func (*Baz) GET(g *gin.Context) {}
