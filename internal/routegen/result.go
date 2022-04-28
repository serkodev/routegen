package routegen

import "os"

type result struct {
	outPath string
	content []byte
}

func (r result) Save() error {
	out, err := os.Create(r.outPath)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = out.Write(r.content)
	return err
}
