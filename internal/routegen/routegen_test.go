package routegen

import (
	"bytes"
	"errors"
	"io/ioutil"
	"os"
	"testing"
)

func TestRouteGen(t *testing.T) {
	wd := "./testdata"

	files, err := ioutil.ReadDir(wd)

	if err != nil {
		t.Error(err)
	}

	for _, f := range files {
		if err := getTest(wd + "/" + f.Name()); err != nil {
			t.Error(err.Error())
		}
	}
}

func getTest(dir string) error {
	results, err := Load(dir, os.Environ())
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
