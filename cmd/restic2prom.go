package main

import (
	"fmt"
	"os/exec"
	"os"
	"strings"
)

func main() {
	cmd := exec.Command(os.Args[1], os.Args[2:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Start()
	err := cmd.Wait()
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			os.Exit(cmd.ProcessState.ExitCode())
		} else {
			fmt.Fprintf(os.Stderr, "exec '%s' failed\n", strings.Join(os.Args[1:], " ") )
			os.Exit(1)
		}
	}
}
