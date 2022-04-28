package routegen

import (
	"bytes"
	"errors"
	"os"
	"testing"
)

func TestRouteGen(t *testing.T) {
	wd := "./testdata/echo"
	if err := getTest(wd); err != nil {
		t.Error(err.Error())
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
