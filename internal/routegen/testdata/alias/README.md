# Sub-route Alias Guide

When using type-definded [sub-route](../subroute/README.md), you can add an alias to customize the route name.

Here we will use [Gin](https://github.com/gin-gonic/gin) as example.

```
📁
|-📁foo
| |-bar.go ✨
| |-baz.go ✨
|-main.go
|-go.mod
```

## Defining Sub-routes

`./foo/bar.go`

```go
package foo

import "github.com/gin-gonic/gin"

// routegen alias=bar-alias
type Bar struct{}

func (*Bar) GET(g *gin.Context) { /* your code */ }
```

`./foo/baz.go`

```go
package foo

import "github.com/gin-gonic/gin"

// routegen alias=baz-alias
type Baz struct{}

func (*Baz) GET(g *gin.Context) {}
```

## Output

`./main_gen.go`

```go
func Build(g *gin.Engine) {
	bar := &routegen_r.Bar{}
	g.GET("/foo/bar-alias", bar.GET)
	baz := &routegen_r.Baz{}
	g.GET("/foo/baz-alias", baz.GET)
}
```
