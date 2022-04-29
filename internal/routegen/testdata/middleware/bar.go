package middleware

import "github.com/gin-gonic/gin"

type Bar struct{}

func (*Bar) Middleware(g *gin.Context) {}

func (*Bar) GET(g *gin.Context) {}
