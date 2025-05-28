package rhel10

import (
	"github.com/osbuild/images/pkg/disk"
	"github.com/osbuild/images/pkg/distro/rhel"
)

func defaultBasePartitionTables(t *rhel.ImageType) (disk.PartitionTable, bool) {
	partitionTable := t.ImageTypeYAML.PartitionTable(t.Arch().Distro().Name(), t.Arch().Name())
	if partitionTable == nil {
		return disk.PartitionTable{}, false
	}

	return *partitionTable, true
}
