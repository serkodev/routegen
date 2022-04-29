package router

import "github.com/gin-gonic/gin"

func Stretch() {
	g := gin.Default()
	g.GET("", func(c *gin.Context) {
		c.String(200, "hello")
	})

	gApiGet := g.Group("/api", func(c *gin.Context) {
		c.String(200, "gApiGet: "+c.FullPath())
		c.Next()
	})
	gApiGet.GET("/user", func(c *gin.Context) { c.String(200, c.FullPath()) })

	blog := gApiGet.Group("/:id")
	blog.GET("", func(c *gin.Context) {
		c.String(200, "/blog: "+c.FullPath())
	})

	gApiPost := g.Group("/api")
	gApiPost.Use(func(c *gin.Context) {
		c.String(200, "gApiPost: "+c.FullPath())
		c.Next()
	})
	gApiPost.POST("/user", func(c *gin.Context) { c.String(200, c.FullPath()) })

	// g.GET("/api/user", func(c *gin.Context) {
	// 	c.String(200, "sec: "+c.FullPath())
	// })

	g.Run()
}
