package main

import (
	"log"
	"os"

	"github.com/serkodev/routegen/internal/routegen"
)

func main() {
	// work dir
	wd, err := os.Getwd()
	if err != nil {
		log.Println("failed to get working directory: ", err)
		return
	}

	routegen.Load(wd, os.Environ())
}
