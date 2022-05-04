package foo

import "github.com/gin-gonic/gin"

// make /foo/baz to /foo/baz-alias
// routegen alias=baz-alias
type Baz struct{}

func (*Baz) GET(g *gin.Context) {}
