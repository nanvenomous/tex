package cmd

import (
	"os"
	"os/exec"
)

func editor(flPth string) error {
	cmd := exec.Command(editorEnvVar, flPth)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

type tmpFileFunc func(tmpFl *os.File) error

func withTempFile(pattern string, tff tmpFileFunc) error {
	tmpFl, err := os.CreateTemp("", pattern)
	if err != nil {
		return err
	}
	defer func() {
		tmpFlName := tmpFl.Name()
		err := tmpFl.Close()
		if err != nil {
			panic(err)
		}
		err = os.Remove(tmpFlName)
		if err != nil {
			panic(err)
		}
	}()
	return tff(tmpFl)
}
