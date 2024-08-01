package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/osbuild/images/pkg/osbuild"

	"github.com/osbuild/images/internal/otkdisk"
)

type Input = otkdisk.Data

func run(r io.Reader, w io.Writer) error {
	var inp Input
	if err := json.NewDecoder(r).Decode(&inp); err != nil {
		return err
	}

	// rhel7 uses PTSgdisk, if we ever need to support this, make this
	// configurable
	partTool := osbuild.PTSfdisk
	stages, err := osbuild.GenImagePrepareStages(inp.Const.Internal.PartitionTable, inp.Const.Filename, partTool)
	if err != nil {
		return fmt.Errorf("cannot make partition stages: %w", err)
	}

	outputJson, err := json.MarshalIndent(stages, "", "  ")
	if err != nil {
		return fmt.Errorf("cannot marshal response: %w", err)
	}
	fmt.Fprintf(w, "%s\n", outputJson)
	return nil
}

func main() {
	if err := run(os.Stdin, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err.Error())
		os.Exit(1)
	}
}
