package disk_test

import (
	"math/rand"
	"testing"

	"github.com/osbuild/images/internal/common"
	"github.com/osbuild/images/internal/testdisk"
	"github.com/osbuild/images/pkg/blueprint"
	"github.com/osbuild/images/pkg/disk"
	"github.com/osbuild/images/pkg/platform"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPartitionTable_GetMountpointSize(t *testing.T) {
	pt := testdisk.MakeFakePartitionTable("/", "/app")

	size, err := pt.GetMountpointSize("/")
	assert.NoError(t, err)
	assert.Equal(t, testdisk.FakePartitionSize, size)

	size, err = pt.GetMountpointSize("/app")
	assert.NoError(t, err)
	assert.Equal(t, testdisk.FakePartitionSize, size)

	// non-existing
	_, err = pt.GetMountpointSize("/custom")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "cannot find mountpoint /custom")
}

func TestPartitionTable_GenerateUUIDs(t *testing.T) {
	pt := disk.PartitionTable{
		Type: "gpt",
		Partitions: []disk.Partition{
			{
				Size:     1 * common.MebiByte,
				Bootable: true,
				Type:     disk.BIOSBootPartitionGUID,
				UUID:     disk.BIOSBootPartitionUUID,
			},
			{
				Size: 2 * common.GibiByte,
				Type: disk.FilesystemDataGUID,
				Payload: &disk.Filesystem{
					// create mixed xfs root filesystem and a btrfs /var partition
					Type:         "xfs",
					Label:        "root",
					Mountpoint:   "/",
					FSTabOptions: "defaults",
					FSTabFreq:    0,
					FSTabPassNo:  0,
				},
			},
			{
				Size: 10 * common.GibiByte,
				Payload: &disk.Btrfs{
					Subvolumes: []disk.BtrfsSubvolume{
						{
							Mountpoint: "/var",
						},
					},
				},
			},
		},
	}

	// Static seed for testing
	/* #nosec G404 */
	rnd := rand.New(rand.NewSource(0))

	pt.GenerateUUIDs(rnd)

	// Check that GenUUID doesn't change already defined UUIDs
	assert.Equal(t, disk.BIOSBootPartitionUUID, pt.Partitions[0].UUID)

	// Check that GenUUID generates fresh UUIDs if not defined prior the call
	assert.Equal(t, "a178892e-e285-4ce1-9114-55780875d64e", pt.Partitions[1].UUID)
	assert.Equal(t, "6e4ff95f-f662-45ee-a82a-bdf44a2d0b75", pt.Partitions[1].Payload.(*disk.Filesystem).UUID)

	// Check that GenUUID generates the same UUID for BTRFS and its subvolumes
	assert.Equal(t, "fb180daf-48a7-4ee0-b10d-394651850fd4", pt.Partitions[2].Payload.(*disk.Btrfs).UUID)
	assert.Equal(t, "fb180daf-48a7-4ee0-b10d-394651850fd4", pt.Partitions[2].Payload.(*disk.Btrfs).Subvolumes[0].UUID)
}

func TestPartitionTable_GenerateUUIDs_VFAT(t *testing.T) {
	pt := disk.PartitionTable{
		Type: "dos",
		Partitions: []disk.Partition{
			{
				Size: 2 * common.GibiByte,
				Type: disk.FilesystemDataGUID,
				Payload: &disk.Filesystem{
					Type:       "vfat",
					Mountpoint: "/boot/efi",
				},
			},
		},
	}

	// Static seed for testing
	/* #nosec G404 */
	rnd := rand.New(rand.NewSource(0))

	pt.GenerateUUIDs(rnd)

	assert.Equal(t, "6e4ff95f", pt.Partitions[0].Payload.(*disk.Filesystem).UUID)
}

func TestNewCustomPartitionTable(t *testing.T) {
	type testCase struct {
		customizations *blueprint.PartitioningCustomization
		options        *disk.CustomPartitionTableOptions
		expected       *disk.PartitionTable
	}

	testCases := map[string]testCase{
		"null": {
			customizations: nil,
			options: &disk.CustomPartitionTableOptions{
				DefaultFSType:      disk.FS_XFS,
				BootMode:           platform.BOOT_HYBRID,
				PartitionTableType: "dos",
			},
			expected: &disk.PartitionTable{
				Type: "dos",
				Size: 202*common.MebiByte + 3*common.GibiByte,
				UUID: "0194fdc2-fa2f-4cc0-81d3-ff12045b73c8",
				Partitions: []disk.Partition{
					{
						Start:    1 * common.MebiByte, // header
						Bootable: true,
						Size:     1 * common.MebiByte,
						Type:     disk.BIOSBootPartitionGUID,
						UUID:     disk.BIOSBootPartitionUUID,
					},
					{
						Start: 2 * common.MebiByte,
						Size:  200 * common.MebiByte,
						Type:  disk.EFISystemPartitionGUID,
						UUID:  disk.EFISystemPartitionUUID,
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
						Start:    202 * common.MebiByte,
						Size:     3 * common.GibiByte,
						Type:     disk.FilesystemDataGUID,
						Bootable: false,
						Payload: &disk.Filesystem{
							Type:         "xfs",
							Label:        "root",
							Mountpoint:   "/",
							FSTabOptions: "defaults",
							UUID:         "6e4ff95f-f662-45ee-a82a-bdf44a2d0b75",
						},
					},
				},
			},
		},
		"plain": {
			customizations: &blueprint.PartitioningCustomization{
				Plain: &blueprint.PlainFilesystemCustomization{
					Filesystems: []blueprint.FilesystemCustomization{
						{
							Mountpoint: "/data",
							MinSize:    20 * common.MebiByte,
							Label:      "data",
							Type:       "ext4",
						},
					},
				},
			},
			options: &disk.CustomPartitionTableOptions{
				DefaultFSType:      disk.FS_XFS,
				BootMode:           platform.BOOT_HYBRID,
				PartitionTableType: "dos",
			},
			expected: &disk.PartitionTable{
				Type: "dos",
				Size: 222*common.MebiByte + 3*common.GibiByte,
				UUID: "0194fdc2-fa2f-4cc0-81d3-ff12045b73c8",
				Partitions: []disk.Partition{
					{
						Start:    1 * common.MebiByte, // header
						Size:     1 * common.MebiByte,
						Bootable: true,
						Type:     disk.BIOSBootPartitionGUID,
						UUID:     disk.BIOSBootPartitionUUID,
					},
					{
						Start: 2 * common.MebiByte,
						Size:  200 * common.MebiByte,
						Type:  disk.EFISystemPartitionGUID,
						UUID:  disk.EFISystemPartitionUUID,
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
						Start:    202 * common.MebiByte,
						Size:     20 * common.MebiByte,
						Type:     disk.FilesystemDataGUID,
						Bootable: false,
						UUID:     "", // partitions on dos PTs don't have UUIDs
						Payload: &disk.Filesystem{
							Type:         "ext4",
							Label:        "data",
							Mountpoint:   "/data",
							UUID:         "6e4ff95f-f662-45ee-a82a-bdf44a2d0b75",
							FSTabOptions: "defaults",
							FSTabFreq:    0,
							FSTabPassNo:  0,
						},
					},
					{
						Start:    222 * common.MebiByte,
						Size:     3 * common.GibiByte,
						Type:     disk.FilesystemDataGUID,
						UUID:     "", // partitions on dos PTs don't have UUIDs
						Bootable: false,
						Payload: &disk.Filesystem{
							Type:         "xfs",
							Label:        "root",
							Mountpoint:   "/",
							UUID:         "fb180daf-48a7-4ee0-b10d-394651850fd4",
							FSTabOptions: "defaults",
							FSTabFreq:    0,
							FSTabPassNo:  0,
						},
					},
				},
			},
		},
		"plain+": {
			customizations: &blueprint.PartitioningCustomization{
				Plain: &blueprint.PlainFilesystemCustomization{
					Filesystems: []blueprint.FilesystemCustomization{
						{
							Mountpoint: "/",
							MinSize:    50 * common.MebiByte,
							Label:      "root",
							Type:       "xfs",
						},
						{
							Mountpoint: "/home",
							MinSize:    20 * common.MebiByte,
							Label:      "home",
							Type:       "ext4",
						},
					},
				},
			},
			options: &disk.CustomPartitionTableOptions{
				DefaultFSType:      disk.FS_EXT4,
				BootMode:           platform.BOOT_HYBRID,
				PartitionTableType: "gpt",
			},
			expected: &disk.PartitionTable{
				Type: "gpt",
				Size: 222*common.MebiByte + 3*common.GibiByte + common.MebiByte, // start + size of last partition + footer

				UUID: "0194fdc2-fa2f-4cc0-81d3-ff12045b73c8",
				Partitions: []disk.Partition{
					{
						Start:    1 * common.MebiByte, // header
						Size:     1 * common.MebiByte,
						Bootable: true,
						Type:     disk.BIOSBootPartitionGUID,
						UUID:     disk.BIOSBootPartitionUUID,
					},
					{
						Start: 2 * common.MebiByte,
						Size:  200 * common.MebiByte,
						Type:  disk.EFISystemPartitionGUID,
						UUID:  disk.EFISystemPartitionUUID,
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
					// root is aligned to the end but not reindexed
					{
						Start:    222 * common.MebiByte,
						Size:     3*common.GibiByte + common.MebiByte - (disk.DefaultSectorSize + (128 * 128)), // grows by 1 grain size (1 MiB) minus the unaligned size of the header to fit the gpt footer
						Type:     disk.FilesystemDataGUID,
						UUID:     "a178892e-e285-4ce1-9114-55780875d64e",
						Bootable: false,
						Payload: &disk.Filesystem{
							Type:         "xfs",
							Label:        "root",
							Mountpoint:   "/",
							FSTabOptions: "defaults",
							UUID:         "6e4ff95f-f662-45ee-a82a-bdf44a2d0b75",
							FSTabFreq:    0,
							FSTabPassNo:  0,
						},
					},
					{
						Start:    202 * common.MebiByte,
						Size:     20 * common.MebiByte,
						Type:     disk.FilesystemDataGUID,
						UUID:     "e2d3d0d0-de6b-48f9-b44c-e85ff044c6b1",
						Bootable: false,
						Payload: &disk.Filesystem{
							Type:         "ext4",
							Label:        "home",
							Mountpoint:   "/home",
							FSTabOptions: "defaults",
							UUID:         "fb180daf-48a7-4ee0-b10d-394651850fd4",
							FSTabFreq:    0,
							FSTabPassNo:  0,
						},
					},
				},
			},
		},
		"lvm": {
			customizations: &blueprint.PartitioningCustomization{
				LVM: &blueprint.LVMCustomization{
					VolumeGroups: []blueprint.VGCustomization{
						{
							Name:    "testvg",
							MinSize: 100 * common.MebiByte,
							LogicalVolumes: []blueprint.LVCustomization{
								{
									Name: "varloglv",
									FilesystemCustomization: blueprint.FilesystemCustomization{
										Mountpoint: "/var/log",
										MinSize:    10 * common.MebiByte,
										Label:      "var-log",
										Type:       "xfs",
									},
								},
								{
									Name: "rootlv",
									FilesystemCustomization: blueprint.FilesystemCustomization{
										Mountpoint: "/",
										MinSize:    50 * common.MebiByte,
										Label:      "root",
										Type:       "xfs",
									},
								},
								{
									Name: "datalv", // untyped logical volume
									FilesystemCustomization: blueprint.FilesystemCustomization{
										Mountpoint: "/data",
										MinSize:    100 * common.MebiByte,
										Label:      "data",
									},
								},
							},
						},
					},
				},
			},
			options: &disk.CustomPartitionTableOptions{
				DefaultFSType: disk.FS_EXT4,
				BootMode:      platform.BOOT_HYBRID,
			},
			expected: &disk.PartitionTable{
				Type: "gpt", // default when unspecified
				UUID: "0194fdc2-fa2f-4cc0-81d3-ff12045b73c8",
				Size: 714*common.MebiByte + 112*common.MebiByte + 3*common.GibiByte + common.MebiByte, // start + size of last partition (VG) + footer
				Partitions: []disk.Partition{
					{
						Start:    1 * common.MebiByte, // header
						Size:     1 * common.MebiByte,
						Bootable: true,
						Type:     disk.BIOSBootPartitionGUID,
						UUID:     disk.BIOSBootPartitionUUID,
					},
					{
						Start: 2 * common.MebiByte,
						Size:  200 * common.MebiByte,
						Type:  disk.EFISystemPartitionGUID,
						UUID:  disk.EFISystemPartitionUUID,
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
						Start:    202 * common.MebiByte,
						Size:     512 * common.MiB,
						Type:     disk.XBootLDRPartitionGUID,
						UUID:     "f83b8e88-3bbf-457a-ab99-c5b252c7429c",
						Bootable: false,
						Payload: &disk.Filesystem{
							Type:         "ext4",
							Label:        "boot",
							Mountpoint:   "/boot",
							FSTabOptions: "defaults",
							UUID:         "6e4ff95f-f662-45ee-a82a-bdf44a2d0b75",
							FSTabFreq:    0,
							FSTabPassNo:  0,
						},
					},
					{
						Start:    714 * common.MebiByte,
						Size:     3*common.GibiByte + 112*common.MebiByte + common.MebiByte - (disk.DefaultSectorSize + (128 * 128)), // the sum of the LVs (rounded to the next 4 MiB extent) grows by 1 grain size (1 MiB) minus the unaligned size of the header to fit the gpt footer
						Type:     disk.LVMPartitionGUID,
						UUID:     "32f3a8ae-b79e-4856-b659-c18f0dcecc77",
						Bootable: false,
						Payload: &disk.LVMVolumeGroup{
							Name:        "testvg",
							Description: "created via lvm2 and osbuild",
							LogicalVolumes: []disk.LVMLogicalVolume{
								{
									Name: "varloglv",
									Size: 10 * common.MebiByte,
									Payload: &disk.Filesystem{
										Label:        "var-log",
										Type:         "xfs",
										Mountpoint:   "/var/log",
										FSTabOptions: "defaults",
										UUID:         "fb180daf-48a7-4ee0-b10d-394651850fd4",
									},
								},
								{
									Name: "rootlv",
									Size: 3 * common.GibiByte,
									Payload: &disk.Filesystem{
										Label:        "root",
										Type:         "xfs",
										Mountpoint:   "/",
										FSTabOptions: "defaults",
										UUID:         "a178892e-e285-4ce1-9114-55780875d64e",
									},
								},
								{
									Name: "datalv",
									Size: 100 * common.MebiByte,
									Payload: &disk.Filesystem{
										Label:        "data",
										Type:         "ext4", // the defaultType
										Mountpoint:   "/data",
										FSTabOptions: "defaults",
										UUID:         "e2d3d0d0-de6b-48f9-b44c-e85ff044c6b1",
									},
								},
							},
						},
					},
				},
			},
		},
		"btrfs": {
			customizations: &blueprint.PartitioningCustomization{
				Btrfs: &blueprint.BtrfsCustomization{
					Volumes: []blueprint.BtrfsVolumeCustomization{
						{
							MinSize: 230 * common.MebiByte,
							Subvolumes: []blueprint.BtrfsSubvolumeCustomization{
								{
									Name:       "subvol/root",
									Mountpoint: "/",
								},
								{
									Name:       "subvol/home",
									Mountpoint: "/home",
								},
								{
									Name:       "subvol/varlog",
									Mountpoint: "/var/log",
								},
							},
						},
					},
				},
			},
			options: &disk.CustomPartitionTableOptions{
				DefaultFSType:      disk.FS_EXT4,
				BootMode:           platform.BOOT_HYBRID,
				PartitionTableType: "gpt",
			},
			expected: &disk.PartitionTable{
				Type: "gpt",
				Size: 714*common.MebiByte + 3*common.GibiByte + common.MebiByte, // start + size of last partition + footer
				UUID: "0194fdc2-fa2f-4cc0-81d3-ff12045b73c8",
				Partitions: []disk.Partition{
					{
						Start:    1 * common.MebiByte, // header
						Size:     1 * common.MebiByte,
						Bootable: true,
						Type:     disk.BIOSBootPartitionGUID,
						UUID:     disk.BIOSBootPartitionUUID,
					},
					{
						Start: 2 * common.MebiByte, // header
						Size:  200 * common.MebiByte,
						Type:  disk.EFISystemPartitionGUID,
						UUID:  disk.EFISystemPartitionUUID,
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
						Start:    202 * common.MebiByte,
						Size:     512 * common.MiB,
						Type:     disk.XBootLDRPartitionGUID,
						UUID:     "a178892e-e285-4ce1-9114-55780875d64e",
						Bootable: false,
						Payload: &disk.Filesystem{
							Type:         "ext4",
							Label:        "boot",
							Mountpoint:   "/boot",
							UUID:         "6e4ff95f-f662-45ee-a82a-bdf44a2d0b75",
							FSTabOptions: "defaults",
							FSTabFreq:    0,
							FSTabPassNo:  0,
						},
					},
					{
						Start:    714 * common.MebiByte,
						Size:     3*common.GibiByte + common.MebiByte - (disk.DefaultSectorSize + (128 * 128)), // grows by 1 grain size (1 MiB) minus the unaligned size of the header to fit the gpt footer
						Type:     disk.FilesystemDataGUID,
						UUID:     "e2d3d0d0-de6b-48f9-b44c-e85ff044c6b1",
						Bootable: false,
						Payload: &disk.Btrfs{
							UUID: "fb180daf-48a7-4ee0-b10d-394651850fd4",
							Subvolumes: []disk.BtrfsSubvolume{
								{
									Name:       "subvol/root",
									Mountpoint: "/",
									UUID:       "fb180daf-48a7-4ee0-b10d-394651850fd4", // same as volume UUID
									Size:       3 * common.GibiByte,
								},
								{
									Name:       "subvol/home",
									Mountpoint: "/home",
									UUID:       "fb180daf-48a7-4ee0-b10d-394651850fd4", // same as volume UUID
								},
								{
									Name:       "subvol/varlog",
									Mountpoint: "/var/log",
									UUID:       "fb180daf-48a7-4ee0-b10d-394651850fd4", // same as volume UUID
								},
							},
						},
					},
				},
			},
		},
	}

	for name := range testCases {
		tc := testCases[name]
		t.Run(name, func(t *testing.T) {
			require := require.New(t)

			// Initialise rng for each test separately, otherwise test run
			// order will affect results
			/* #nosec G404 */
			rnd := rand.New(rand.NewSource(0))
			pt, err := disk.NewCustomPartitionTable(tc.customizations, tc.options, rnd)

			require.NoError(err)
			require.Equal(tc.expected, pt)
		})
	}

}
