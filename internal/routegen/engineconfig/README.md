# Custom config for any web framework

You can use `routegen` to generate route without any additional config if you are using below web framework:

- [Gin](https://github.com/gin-gonic/gin)
- [echo](https://github.com/labstack/echo)

 Welcome to contribute any other engine config of famous web framework to let `routegen` support by default.

## Load custom config

You can either put `routegen.json` file on your folder where run `routegen` command or use `-e` flag to specific the custom engine config. For example:

```
routegen -e=./some-web-framework.json .
```

## Format

Here is engine config example for `gin`.

```json
{
  "types": [
    "*github.com/gin-gonic/gin.Engine",
    "*github.com/gin-gonic/gin.RouterGroup"
  ],
  "selectors": [
    "Request",
    "GET",
    "POST",
    "DELETE",
    "PATCH",
    "PUT",
    "OPTIONS",
    "HEAD"
  ],
  "expr": {
    "Middleware": "{{ .ident }}.Use({{ .handle }})",
    "Request": "{{ .ident }}.Any(\"{{ .route }}\", {{ .handle }})",
    "_default": "{{ .ident }}.{{ .sel }}(\"{{ .route }}\", {{ .handle }})"
  },
  "middleware": {
    "selector": "Middleware",
    "group_expr": "{{ .ident }}.Group(\"{{ .route }}\")"
  }
}
```

| Key          | Type                | Description                                                                                                                                                                       |
| ------------ | ------------------- | --------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `types`      | `string[]`          | Target package name with object type of your web frameworks                                                                                                                       |
| `selectors`  | `string[]`          | Target selector of your web frameworks HTTP methods or callback                                                                                                                   |
| `expr`       | `map[string]string` | `_default` is the format to generate route by `selectors`.<br />If you want to specific other formats from `selectors` you can add the `selector` as key and the format as value. |
| `middleware` | `map[string]string` | `selector` is the selector of middleware<br />`group_expr` is the format of the middleware replace group                                                                          |
