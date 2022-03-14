package routegen

import (
	"os"
	"testing"
)

func TestRouteGen(t *testing.T) {
	wd := "../example/helloworld/router_echo"
	Load(wd, os.Environ())
}
