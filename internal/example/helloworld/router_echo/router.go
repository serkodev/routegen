//go:build routegeninject

package router

import (
	"github.com/labstack/echo"
	"github.com/serkodev/routegen"
)

func Build(e *echo.Echo) { // gin.RouterGroup
	routegen.Build(e)
}
