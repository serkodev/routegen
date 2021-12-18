//go:build pbrinject

//The build tag makes sure the stub is not built in the final build.

package router

import (
	"github.com/gin-gonic/gin"
	"github.com/serkodev/pbr/pbr"
)

type PlaceOrderRequest struct {
	// A combination of case-sensitive alphanumerics, all numbers, or all letters of up to 32 characters.
	clientOrderID *string `param:"clientOid,required" defaultValuer:"uuid()"`
}

//go:generate echo "start gen pbr..."
////go:generate go run github.com/serkodev/pbr
//go:generate $HOME/go/bin/pbr
func Run() {
	router := gin.Default()

	// r := &route.Route{}
	// about.GET(r)
	// xx.GET(r)

	pbr.Build(router)

	router.Run()
}
