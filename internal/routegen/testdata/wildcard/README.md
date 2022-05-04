# Wildcard & named-parameter Guide

Some of the web framework supports using wildcard & named-parameter in routes. You can define these routes using folder name too.

| Folder Name | route  | Type                              |
| ----------- | ------ | --------------------------------- |
| `_`         | `*`    | Wildcard ✨                       |
| `_id`       | `:id`  | Named-parameter ✨                |
| `__id`      | `_id`  | **Normal route** with `_` symbol  |
| `___id`     | `__id` | **Normal route** with `__` symbol |

Here we will use [Gin](https://github.com/gin-gonic/gin) as example.

```
📁
|-📁any
| |-📁_
|   |-handle.go
|-named
| |-📁_id
|   |-handle.go
|-main.go
|-go.mod
```

## Output

`./main_gen.go`

```go
func Build(g *gin.Engine) {
	g.GET("/any/*", routegen_r.GET)
	g.GET("/named/:id", routegen_r2.GET)
}
```
