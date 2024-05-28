package main

import (
	"fmt"
	"os"

	"github.com/osbuild/images/pkg/disk"
)

func run(args []string) error {

	return nil
}

func main() {
	if err := run(os.Args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
