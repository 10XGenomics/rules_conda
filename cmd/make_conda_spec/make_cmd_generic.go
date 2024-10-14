//go:build !linux

package main

import (
	"os/exec"
)

func makeCmd(conda, condaPath string, args ...string) *exec.Cmd {
	cmd := exec.Command(conda, args...)
	cmd.Env = []string{"PATH=" + condaPath}
	return cmd
}
