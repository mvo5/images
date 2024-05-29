package main

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"os"

	"github.com/osbuild/images/internal/buildconfig"
	"github.com/osbuild/images/internal/cmdutil"
	"github.com/osbuild/images/internal/common"
	"github.com/osbuild/images/pkg/disk"
	"github.com/osbuild/images/pkg/osbuild"
)

var basePt = disk.PartitionTable{
	UUID: "D209C89E-EA5E-4FBD-B161-B461CCE297E0",
	Type: "gpt",
	Partitions: []disk.Partition{
		{
			Size:     1 * common.MebiByte,
			Bootable: true,
			Type:     disk.BIOSBootPartitionGUID,
			UUID:     disk.BIOSBootPartitionUUID,
		},
		{
			Size: 200 * common.MebiByte,
			Type: disk.EFISystemPartitionGUID,
			UUID: disk.EFISystemPartitionUUID,
			Payload: &disk.Filesystem{
				Type:         "vfat",
				UUID:         disk.EFIFilesystemUUID,
				Mountpoint:   "/boot/efi",
				Label:        "EFI-SYSTEM",
				FSTabOptions: "defaults,uid=0,gid=0,umask=077,shortname=winnt",
				FSTabFreq:    0,
				FSTabPassNo:  2,
			},
		},
		{
			Size: 500 * common.MebiByte,
			Type: disk.FilesystemDataGUID,
			UUID: disk.FilesystemDataUUID,
			Payload: &disk.Filesystem{
				Type:         "ext4",
				Mountpoint:   "/boot",
				Label:        "boot",
				FSTabOptions: "defaults",
				FSTabFreq:    0,
				FSTabPassNo:  0,
			},
		},
		{
			Size: 2 * common.GibiByte,
			Type: disk.FilesystemDataGUID,
			UUID: disk.RootPartitionUUID,
			Payload: &disk.Filesystem{
				Type:         "ext4",
				Label:        "root",
				Mountpoint:   "/",
				FSTabOptions: "defaults",
				FSTabFreq:    0,
				FSTabPassNo:  0,
			},
		},
	},
}

type OtkGenPartitionInput struct {
	Options    *OtkPartOptions `json:"options"`
	Partitions []*OtkPartition `json:"partitions"`
}

type OtkPartOptions struct {
	Uefi *OtkPartUEFI `json:"uefi"`
	Bios bool         `json:"bios"`
	// XXX: enum?
	Type string `json:"type"`
	Size string `json:"size"`

	SectorSize uint64 `json:"sector_size"`
}

type OtkPartUEFI struct {
	Size string `json:"size"`
}

type OtkPartition struct {
	Name       string `json:"name"`
	Mountpoint string `json:"mountpoint"`
	Label      string `json:"label"`
	Size       string `json:"size"`
	Type       string `json:"type"`

	// TODO: add sectorlvm,luks, see https://github.com/achilleas-k/images/pull/2#issuecomment-2136025471
	// also add "uuid", "freq", more(?) so that users can override calculcated
	// values in a controlled way
}

type otkGenPartitionsOutput struct {
	PartitionTable *disk.PartitionTable `json:"internal-partition-table"`
	KernelOptsList []string             `json:"kernel_opts_list"`
}

func run(r io.Reader) (*otkGenPartitionsOutput, error) {
	var genPartInput OtkGenPartitionInput
	if err := json.NewDecoder(r).Decode(&genPartInput); err != nil {
		return nil, err
	}

	rngSeed, err := cmdutil.SeedArgFor(&buildconfig.BuildConfig{}, "", "", "")
	if err != nil {
		return nil, err
	}
	source := rand.NewSource(rngSeed)
	// math/rand is good enough in this case
	/* #nosec G404 */
	rng := rand.New(source)

	pt, err := disk.NewPartitionTable(&basePt, nil, 0, disk.DefaultPartitioningMode, nil, rng)
	if err != nil {
		return nil, err
	}

	kernelOptions := osbuild.GenImageKernelOptions(pt)
	otkPart := &otkGenPartitionsOutput{
		PartitionTable: pt,
		KernelOptsList: kernelOptions,
	}

	return otkPart, nil
}

func main() {
	output, err := run(os.Stdin)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v", err.Error())
		os.Exit(1)
	}
	fmt.Println(output)
}
