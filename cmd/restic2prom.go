package main

import (
	"bufio"
	"fmt"
	"github.com/juergenhoetzel/restic2prom/metrics"
	"github.com/spf13/cobra"
	"os"
	"os/exec"
	"strings"
)

func main() {
	args := make([]string, len(os.Args))
	var prom *metrics.Prom
	// we leverage cobra to get the repository from restic flags or environment variables
	rootCmd := &cobra.Command{
		Use: "restic",
		Run: func(cmd *cobra.Command, args []string) {
			repo, _ := cmd.Flags().GetString("repo")
			// FIXME
			prom = metrics.New(repo, "/tmp/out.prom")
		},
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
	}
	rootCmd.PersistentFlags().StringP("repo", "r", os.Getenv("RESTIC_REPOSITORY"), "_")
	rootCmd.Execute()
	copy(args, os.Args[2:])
	cmd := exec.Command(os.Args[1], args...)
	cmd.Stdin = os.Stdin
	stdoutPipe, _ := cmd.StdoutPipe()
	stdoutReader := bufio.NewReader(stdoutPipe)
	stderrPipe, _ := cmd.StderrPipe()
	stderrReader := bufio.NewReader(stderrPipe)

	cmd.Start()

	go prom.CollectStdout(stdoutReader)
	go prom.CollectStderr(stderrReader)

	err := cmd.Wait()
	if !prom.WriteToTextFile() {
		fmt.Fprintf(os.Stderr, "Did not receive JSON metrics. Missed to add '--json' flag?\n")
	}
	// FIXME: Write metrics in error case?
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			os.Exit(cmd.ProcessState.ExitCode())
		} else {
			fmt.Fprintf(os.Stderr, "exec '%s' failed\n", strings.Join(os.Args[1:], " "))
			os.Exit(1)
		}
	}
}
