// Code generated by routegen. DO NOT EDIT.

//go:build !routegeninject
// +build !routegeninject

package router

import (
	"github.com/labstack/echo"
	routegen_r "example.com/foo/router_echo/blog"
)

func Build(e *echo.Echo) {
	e.GET("/blog", routegen_r.GET)
}
