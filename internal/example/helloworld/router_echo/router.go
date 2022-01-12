//go:build pbrinject

package router

import (
	"github.com/labstack/echo"
	"github.com/serkodev/pbr"
)

func Build(e *echo.Echo) { // gin.RouterGroup
	pbr.Build(e)
}
