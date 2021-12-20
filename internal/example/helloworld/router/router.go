//go:build pbrinject

//The build tag makes sure the stub is not built in the final build.

package router

import (
	"github.com/gin-gonic/gin"
	"github.com/serkodev/pbr"
)

type tts struct {
	a int
}

//go:generate echo "start gen pbr..."
////go:generate go run github.com/serkodev/pbr
//go:generate $HOME/go/bin/pbr
func Run() {
	rrrrrr := gin.Default()

	// r := &route.Route{}
	// about.GET(r)
	// xx.GET(r)

	// b := tts{a: 123}
	pbr.Build(rrrrrr)

	rrrrrr.Run()
}
