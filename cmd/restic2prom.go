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

func startRestic(prom *metrics.Prom, resticArgs []string) {
	cmd := exec.Command(resticArgs[0], resticArgs[1:]...)
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
	if err != nil {
		if _, ok := err.(*exec.ExitError); ok {
			os.Exit(cmd.ProcessState.ExitCode())
		} else {
			fmt.Fprintf(os.Stderr, "exec '%s' failed\n", strings.Join(os.Args[1:], " "))
			os.Exit(1)
		}
	}
}

func main() {
	// we leverage cobra to get the repository from restic flags or environment variables
	rootCmd := &cobra.Command{
		Use: "restic",
		Run: func(cmd *cobra.Command, args []string) {
			repo, _ := cmd.Flags().GetString("repo")
			textfile, _ := cmd.Flags().GetString("textfile")
			if (!strings.HasSuffix(textfile, ".prom")) {
				fmt.Fprintf(os.Stderr, "Invalid textfile name '%s' (missing '.prom' suffix)\n", textfile)
				os.Exit(1)
			}
			prom := metrics.New(repo, textfile)
			startRestic(prom, os.Args[3:])
		},
		FParseErrWhitelist: cobra.FParseErrWhitelist{
			UnknownFlags: true,
		},
	}
	rootCmd.PersistentFlags().StringP("repo", "r", os.Getenv("RESTIC_REPOSITORY"), "_")
	rootCmd.PersistentFlags().StringP("textfile", "t", "", ".prom output file (required).")
	rootCmd.MarkPersistentFlagRequired("textfile")
	rootCmd.Execute()
}
