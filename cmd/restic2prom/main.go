package main

import (
	"bufio"
	"fmt"
	"github.com/juergenhoetzel/restic2prom/internal/metrics"
	"flag"
	"os"
	"os/exec"
	"io/ioutil"
	"strings"
)

func startRestic(prom *metrics.Prom, resticArgs []string) {
	cmd := exec.Command(resticArgs[0], resticArgs[1:]...)
	cmd.Stdin = os.Stdin
	stdoutPipe, _ := cmd.StdoutPipe()
	stdoutReader := bufio.NewReader(stdoutPipe)
	stderrPipe, _ := cmd.StderrPipe()
	stderrReader := bufio.NewReader(stderrPipe)

	if err := cmd.Start(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to start command: %v\n", err)
		os.Exit(1)
	}

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
	var textfile = flag.String("t", "", "node exporter textfile")
	flag.Parse()
	if (!strings.HasSuffix(*textfile, ".prom")) {
		fmt.Fprintf(os.Stderr, "Invalid textfile name '%s' (missing '.prom' suffix)\n", *textfile)
		os.Exit(1)
	}

	repo := os.Getenv("RESTIC_REPOSITORY")
	// poor mans command line parsing of retic command
	if (repo == "") {
		for i, arg := range flag.Args() {
			if arg == "-r" || arg == "--repo" {
				repo = flag.Arg(i+1)
			} else
			if arg == "--repository-file" {
				data, err := ioutil.ReadFile(flag.Arg(i+1));
				if err != nil {
					fmt.Fprintf(os.Stderr, "Invalid repository filename '%s': %s)\n", flag.Arg(i+1),err)
					os.Exit(1);
				}
				repo = strings.TrimSpace(string(data))
			}
		}
	}
	prom := metrics.New(repo, *textfile)
	startRestic(prom, flag.Args())
}
