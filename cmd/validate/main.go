// Command validate checks the events catalog and prints every problem found.
//
// Usage: go run ./cmd/validate [catalog directory, default "."]
package main

import (
	"fmt"
	"io"
	"os"

	"events.uy/data/internal/validate"
)

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	os.Exit(run(root, os.Stdout, os.Stderr))
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
