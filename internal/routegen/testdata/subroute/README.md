# Sub-route Guide

In addition to adding a folder to generate route, you can also define request selectors of `public struct{} type` in package, `routegen` will generate sub-routes from them.

Here we will use [Gin](https://github.com/gin-gonic/gin) as example.

```
📁
|-📁foo
| |-bar.go ✨
| |-baz.go ✨
| |-handle.go
|-main.go
|-go.mod
```

## Defining Sub-routes

`./foo/bar.go`

```go
package foo

import "github.com/gin-gonic/gin"

type Bar struct{}

func (*Bar) GET(g *gin.Context) { /* your code */ }
```

`./foo/baz.go`

```go
package foo

import "github.com/gin-gonic/gin"

type Baz struct{}

func (*Baz) GET(g *gin.Context) {}
```

## Output

`./main_gen.go`

```go
func Build(g *gin.Engine) {
	g.GET("/foo", routegen_r.GET)
	bar := &routegen_r.Bar{}
	g.GET("/foo/bar", bar.GET)
	baz := &routegen_r.Baz{}
	g.GET("/foo/baz", baz.GET)
}
```
