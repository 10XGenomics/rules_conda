package main

import (
	"os/exec"
	"syscall"
)

func makeCmd(conda, condaPath string, args ...string) *exec.Cmd {
	cmd := exec.Command(conda, args...)
	cmd.Env = []string{"PATH=" + condaPath}
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGTERM,
	}
	return cmd
}
