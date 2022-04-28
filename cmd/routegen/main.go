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

	results, err := routegen.Load(wd, os.Environ())
	if err != nil {
		log.Println("generate route error: ", err)
	}
	for _, r := range results {
		r.Save()
	}
}
