package foo

import "github.com/gin-gonic/gin"

// make /foo/bar to /foo/bar-alias
// routegen alias=bar-alias
type Bar struct{}

func (*Bar) GET(g *gin.Context) {}
