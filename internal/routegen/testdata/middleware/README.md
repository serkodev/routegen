# Middleware Guide

After adding a `Middleware` selector in package, `routegen` will generate a `Group` including all sub-route with `middleware`.

Here we will use [Gin](https://github.com/gin-gonic/gin) as example.

```
📁
|-📁foo
| |-handle.go
| |-middleware.go ✨
| |-📁bar
| | |-handle.go
|-main.go
|-go.mod
```

## Defining Middleware

`./foo/middleware.go`

```go
package foo

import "github.com/gin-gonic/gin"

func Middleware(c *gin.Context) {
    // your code
}
```

## Output

`./main_gen.go`

```go
func Build(g *gin.Engine) {
	grp := g.Group("/")
	{
		grp.Use(Middleware)
		grp2 := grp.Group("/foo")
		{
			grp2.Use(routegen_r.Middleware) // foo/middleware.go
			grp2.GET("", routegen_r.GET)
			grp2.GET("/bar", routegen_r2.GET)
		}
	}
}
```
