package main

import (
	"flag"
	"log"
	"os"

	"github.com/serkodev/routegen/internal/routegen"
)

func main() {
	engineConfigPtr := flag.String("e", "", "engine config path for custom web framework")
	flag.Parse()

	// work dir
	wd, err := os.Getwd()
	if err != nil {
		log.Println("failed to get working directory: ", err)
		return
	}

	// get dir
	dir := packages(flag.Args())

	// generate routes
	results, err := routegen.Load(wd, os.Environ(), dir, *engineConfigPtr)
	if err != nil {
		log.Println("generate route error: ", err)
	}

	// save results to file
	for _, r := range results {
		if err := r.Save(); err != nil {
			log.Println("error output", r.OutPath(), err.Error())
		} else {
			log.Println("routegen saved", r.OutPath())
		}
	}
}

func packages(args []string) string {
	if len(args) == 0 {
		return "."
	}
	return args[0]
}
