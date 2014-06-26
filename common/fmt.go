package common

import (
	"bytes"
	"errors"
	"io"
	"os/exec"
)

func GoFmt(dst io.Writer, src io.Reader) error {
	cmd := exec.Command("gofmt")
	cmd.Stdin = src
	cmd.Stdout = dst

	errOut := new(bytes.Buffer)
	cmd.Stderr = errOut

	err := cmd.Run()
	if err != nil {
		return errors.New(string(errOut.Bytes()))
	}

	return nil
}
