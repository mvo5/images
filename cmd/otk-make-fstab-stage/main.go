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

	pt := inp.Const.Internal.PartitionTable
	stage := osbuild.NewFSTabStage(osbuild.NewFSTabStageOptions(pt))

	outputJson, err := json.MarshalIndent(stage, "", "  ")
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
