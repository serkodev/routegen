package main

import (
	"log"
	"os"

	"github.com/serkodev/pbr/internal/pbr"
)

func main() {
	// work dir
	wd, err := os.Getwd()
	if err != nil {
		log.Println("failed to get working directory: ", err)
		return
	}

	pbr.Load(wd, os.Environ())
}
