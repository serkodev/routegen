# routegen

File-system based route generator for Go. Compatible with any web frameworks.

> ⚗ This project is in beta, it may contain bugs and have not being tested at all. Use under your own risk, but feel free to test, make pull request and improve this project.

## Features

- [x] Generate routes from file-system
- [x] All web frameworks compatible (customizable)
- [x] Web framework auto detection (currently supports [Gin](https://github.com/gin-gonic/gin) & [echo](https://github.com/labstack/echo))
- [x] Support middleware
- [x] Support route with wildcard `/foo/*`
- [x] Support route with named parameter `/foo/:id`
- [x] Support route with alias

## Install

```
go install github.com/serkodev/routegen@lastest
```

## How it works?

`routegen` will scan your go project folders and generate routes when detects special function name (`GET`, `POST`, etc) in your package. It will use the relative file path as the route path, you may also modify the route name by `alias` and use wildcard or named parameter. The method of code injection refers to [wire](https://github.com/google/wire).

Here we will use [Gin](https://github.com/gin-gonic/gin) as a test example. Create the folder strcuture as below

```
📁
|-📁foo
| |-handle.go
| |-📁bar
| | |-handle.go
|-main.go
|-go.mod
```

Create `./main.go`

```go
//go:build routegeninject

package main

import (
	"github.com/gin-gonic/gin"
	"github.com/serkodev/routegen"
)

func Build(g *gin.Engine) {
    // important! placeholder for routes output
	routegen.Build(g)
}

func main() {
	g := gin.Default()
	Build(g)
	g.Run()
}
```

Create `./foo/handle.go`

```go
package foo

import "github.com/gin-gonic/gin"

func GET(c *gin.Context) {}
```

Create `./foo/bar/handle.go`

```go
package bar

import "github.com/gin-gonic/gin"

func GET(c *gin.Context) {}
```

Run generate command at your project root
```
routegen .
```

`main_gen.go` will be generated. 🎉

```go
// Code generated by routegen. DO NOT EDIT.

//go:build !routegeninject
// +build !routegeninject

package main

import (
	"github.com/gin-gonic/gin"
	routegen_r "example.com/helloworld/foo"
	routegen_r2 "example.com/helloworld/foo/bar"
)

func Build(g *gin.Engine) {
	g.GET("/foo", routegen_r.GET)
	g.GET("/foo/bar", routegen_r2.GET)
}

func main() {
	g := gin.Default()
	Build(g)
	g.Run()
}
```

# Documentation

- [Wildcard & named parameter](./internal/routegen/testdata/wildcard/README.md)
- [Middleware](./internal/routegen/testdata/middleware/README.md)
- [Sub-route](./internal/routegen/testdata/subroute/README.md): Create routes with public type
- [Route alias](./internal/routegen/testdata/alias/README.md): Customize sub-route name
