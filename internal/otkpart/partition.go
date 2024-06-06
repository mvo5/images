package otkpart

import (
	"github.com/osbuild/images/pkg/disk"
)

type PartitionTable struct {
	Const Const `json:"const"`
}

type Const struct {
	KernelOptsList []string `json:"kernel_opts_list"`
	// we generate this for convenience for otk users, so that they
	// can write, e.g. "filesystem.partition_map.boot.uuid"
	PartitionMap map[string]Partition `json:"partition_map"`
	Internal     Internal             `json:"internal"`

	// XXX: or diskname?
	Filename string `json:"filename"`
}

// "exported" view of partitions, this is an API so only add things here
// that are really needed and unlikely to change
type Partition struct {
	// not a UUID type because fat UUIDs are not compliant
	UUID string `json:"uuid"`
}

type Internal struct {
	PartitionTable *disk.PartitionTable `json:"partition-table"`
}
