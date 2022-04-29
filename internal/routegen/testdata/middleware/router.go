//go:build routegeninject

package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/serkodev/routegen"
)

func Build(g *gin.Engine) {
	routegen.Build(g)
}
