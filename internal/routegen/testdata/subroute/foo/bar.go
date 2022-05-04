package foo

import "github.com/gin-gonic/gin"

type Bar struct{}

func (*Bar) GET(g *gin.Context) {}

// expect not exported

type privateBar struct{}

func (*privateBar) GET(g *gin.Context) {}

var _ privateBar
