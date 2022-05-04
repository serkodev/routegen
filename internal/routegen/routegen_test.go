package routegen

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"sync"
	"testing"
)

func TestRouteGen(t *testing.T) {
	wd := "./testdata"

	files, err := ioutil.ReadDir(wd)

	if err != nil {
		t.Error(err)
	}

	var wg sync.WaitGroup
	for _, f := range files {
		if !f.IsDir() {
			continue
		}
		wg.Add(1)
		go func(dir string) {
			defer wg.Done()
			if err := getTest(dir); err != nil {
				t.Error(dir, err.Error())
			}
		}(wd + "/" + f.Name())
	}
	wg.Wait()
}

func getTest(dir string) error {
	results, err := Load(dir, os.Environ(), ".")
	if err != nil {
		return err
	}
	for _, result := range results {
		o, err := os.ReadFile(result.outPath)
		if err != nil {
			return err
		}
		if c := bytes.Compare(result.content, o); c != 0 {
			return errors.New("result not match " + dir)
		}
	}
	return nil
}
