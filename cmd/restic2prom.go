package main

import (
	"fmt"
	"os/exec"
	"os"
	"github.com/juergenhoetzel/restic2prom/internal/util"
	"strings"
)

func main() {
	args := make([]string, len(os.Args))
	copy(args, os.Args[2:])
	if ! util.Contains(args, "--json") {
		args = append([]string{"--json"}, args...)
	}
	cmd := exec.Command(os.Args[1], args...)
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
