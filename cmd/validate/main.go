// Command validate checks the events catalog and prints every problem found.
//
// Usage: go run ./cmd/validate [catalog directory, default "."]
package main

import (
	"fmt"
	"io"
	"os"
	"strings"

	"events.uy/data/internal/validate"
)

func main() {
	os.Exit(runArgs(os.Args[1:], os.Stdout, os.Stderr))
}

// runArgs interprets the command line: at most one argument, a directory.
func runArgs(args []string, stdout, stderr io.Writer) int {
	root := "."
	if len(args) > 0 {
		if strings.HasPrefix(args[0], "-") {
			fmt.Fprintln(stderr, "usage: validate [catalog directory]")
			return 2
		}
		root = args[0]
	}
	return run(root, stdout, stderr)
}

func run(root string, stdout, stderr io.Writer) int {
	rep, err := validate.Dir(root)
	if err != nil {
		fmt.Fprintln(stderr, "validate:", err)
		return 2
	}
	for _, p := range rep.Problems {
		fmt.Fprintln(stderr, p)
	}
	if n := len(rep.Problems); n > 0 {
		fmt.Fprintf(stderr, "\n%d problem(s) found in %d event file(s) checked\n", n, rep.Events)
		return 1
	}
	fmt.Fprintf(stdout, "OK: %d event file(s) checked\n", rep.Events)
	return 0
}
