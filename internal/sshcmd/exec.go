package sshcmd

import (
	"errors"
	"os"
	"os/exec"
	"runtime"
	"syscall"
)

func Exec(args []string) error {
	if len(args) == 0 {
		return errors.New("empty command")
	}

	if runtime.GOOS == "windows" {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	}

	binary, err := exec.LookPath(args[0])
	if err != nil {
		return err
	}
	return syscall.Exec(binary, args, os.Environ())
}
