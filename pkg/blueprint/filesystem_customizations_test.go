package blueprint_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"

	"github.com/osbuild/images/pkg/blueprint"
	"github.com/osbuild/images/pkg/pathpolicy"
)

// ensure all fields that are supported are filled here
var allFieldsFsc = blueprint.FilesystemCustomization{
	Mountpoint: "/data",
	MinSize:    1234567890,
}

func TestFilesystemCustomizationMarshalUnmarshalTOML(t *testing.T) {
	b, err := toml.Marshal(allFieldsFsc)
	assert.NoError(t, err)

	var fsc blueprint.FilesystemCustomization
	err = toml.Unmarshal(b, &fsc)
	assert.NoError(t, err)
	assert.Equal(t, fsc, allFieldsFsc)
}

func TestFilesystemCustomizationMarshalUnmarshalJSON(t *testing.T) {
	b, err := json.Marshal(allFieldsFsc)
	assert.NoError(t, err)

	var fsc blueprint.FilesystemCustomization
	err = json.Unmarshal(b, &fsc)
	assert.NoError(t, err)
	assert.Equal(t, fsc, allFieldsFsc)
}

func TestFilesystemCustomizationUnmarshalTOMLUnhappy(t *testing.T) {
	cases := []struct {
		name  string
		input string
		err   string
	}{
		{
			name: "mountpoint not string",
			input: `mountpoint = 42
			minsize = 42`,
			err: "toml: line 0: TOML unmarshal: mountpoint must be string, got 42 of type int64",
		},
		{
			name: "misize nor string nor int",
			input: `mountpoint="/"
			minsize = true`,
			err: "toml: line 0: TOML unmarshal: minsize must be integer or string, got true of type bool",
		},
		{
			name: "misize not parseable",
			input: `mountpoint="/"
			minsize = "20 KG"`,
			err: "toml: line 0: TOML unmarshal: minsize is not valid filesystem size (unknown data size units in string: 20 KG)",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var fsc blueprint.FilesystemCustomization
			err := toml.Unmarshal([]byte(c.input), &fsc)
			assert.EqualError(t, err, c.err)
		})
	}
}

func TestFilesystemCustomizationUnmarshalJSONUnhappy(t *testing.T) {
	cases := []struct {
		name  string
		input string
		err   string
	}{
		{
			name:  "mountpoint not string",
			input: `{"mountpoint": 42, "minsize": 42}`,
			err:   "JSON unmarshal: mountpoint must be string, got 42 of type float64",
		},
		{
			name:  "misize nor string nor int",
			input: `{"mountpoint":"/", "minsize": true}`,
			err:   "JSON unmarshal: minsize must be float64 number or string, got true of type bool",
		},
		{
			name:  "misize not parseable",
			input: `{ "mountpoint": "/", "minsize": "20 KG"}`,
			err:   "JSON unmarshal: minsize is not valid filesystem size (unknown data size units in string: 20 KG)",
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			var fsc blueprint.FilesystemCustomization
			err := json.Unmarshal([]byte(c.input), &fsc)
			assert.EqualError(t, err, c.err)
		})
	}
}

func TestCheckMountpointsPolicy(t *testing.T) {
	policy := pathpolicy.NewPathPolicies(map[string]pathpolicy.PathPolicy{
		"/": {Exact: true},
	})

	mps := []blueprint.FilesystemCustomization{
		{Mountpoint: "/foo"},
		{Mountpoint: "/boot/"},
	}

	expectedErr := `The following errors occurred while setting up custom mountpoints:
path "/foo" is not allowed
path "/boot/" must be canonical`
	err := blueprint.CheckMountpointsPolicy(mps, policy)
	assert.EqualError(t, err, expectedErr)
}

func TestPartitioningValidation(t *testing.T) {
	type testCase struct {
		partitioning *blueprint.PartitioningCustomization
		expected     error
	}

	testCases := map[string]testCase{
		"null": {
			partitioning: nil,
			expected:     nil,
		},
		"happy-plain": {
			partitioning: &blueprint.PartitioningCustomization{
				Plain: &blueprint.PlainFilesystemCustomization{
					Filesystems: []blueprint.FilesystemCustomization{
						{
							Mountpoint: "/data",
						},
					},
				},
			},
			expected: nil,
		},
		"happy-plain+btrfs": {
			partitioning: &blueprint.PartitioningCustomization{
				Plain: &blueprint.PlainFilesystemCustomization{
					Filesystems: []blueprint.FilesystemCustomization{
						{
							Mountpoint: "/data",
						},
					},
				},
				Btrfs: &blueprint.BtrfsCustomization{
					Volumes: []blueprint.BtrfsVolumeCustomization{
						{
							Subvolumes: []blueprint.BtrfsSubvolumeCustomization{
								{
									Name:       "root",
									Mountpoint: "/",
								},
							},
						},
					},
				},
			},
			expected: nil,
		},
		"happy-plain+lvm": {
			partitioning: &blueprint.PartitioningCustomization{
				Plain: &blueprint.PlainFilesystemCustomization{
					Filesystems: []blueprint.FilesystemCustomization{
						{
							Mountpoint: "/data",
						},
					},
				},
				LVM: &blueprint.LVMCustomization{
					VolumeGroups: []blueprint.VGCustomization{
						{
							LogicalVolumes: []blueprint.LVCustomization{
								{
									FilesystemCustomization: blueprint.FilesystemCustomization{Mountpoint: "/"},
								},
							},
						},
					},
				},
			},
			expected: nil,
		},
		"unhappy-btrfs+lvm": {
			partitioning: &blueprint.PartitioningCustomization{
				Plain: &blueprint.PlainFilesystemCustomization{
					Filesystems: []blueprint.FilesystemCustomization{
						{
							Mountpoint: "/data",
						},
					},
				},
				Btrfs: &blueprint.BtrfsCustomization{
					Volumes: []blueprint.BtrfsVolumeCustomization{
						{
							Subvolumes: []blueprint.BtrfsSubvolumeCustomization{
								{
									Mountpoint: "/backup",
								},
							},
						},
					},
				},
				LVM: &blueprint.LVMCustomization{
					VolumeGroups: []blueprint.VGCustomization{
						{
							LogicalVolumes: []blueprint.LVCustomization{
								{
									FilesystemCustomization: blueprint.FilesystemCustomization{Mountpoint: "/"},
								},
							},
						},
					},
				},
			},
			expected: fmt.Errorf("btrfs and lvm partitioning cannot be combined"),
		},
		"unhappy-plain-dupes": {
			partitioning: &blueprint.PartitioningCustomization{
				Plain: &blueprint.PlainFilesystemCustomization{
					Filesystems: []blueprint.FilesystemCustomization{
						{
							Mountpoint: "/data",
						},
						{
							Mountpoint: "/",
						},
						{
							Mountpoint: "/home",
						},
						{
							Mountpoint: "/data",
						},
					},
				},
			},
			expected: fmt.Errorf("duplicate mountpoint \"/data\" in partitioning customizations"),
		},
		"unhappy-plain+btrfs-dupes": {
			partitioning: &blueprint.PartitioningCustomization{
				Plain: &blueprint.PlainFilesystemCustomization{
					Filesystems: []blueprint.FilesystemCustomization{
						{
							Mountpoint: "/data",
						},
					},
				},
				Btrfs: &blueprint.BtrfsCustomization{
					Volumes: []blueprint.BtrfsVolumeCustomization{
						{
							Subvolumes: []blueprint.BtrfsSubvolumeCustomization{
								{
									Name:       "root",
									Mountpoint: "/",
								},
								{
									Name:       "home",
									Mountpoint: "/home",
								},
								{
									Name:       "data",
									Mountpoint: "/data",
								},
							},
						},
					},
				},
			},
			expected: fmt.Errorf("duplicate mountpoint \"/data\" in partitioning customizations"),
		},
		"unhappy-plain+lvm-dupes": {
			partitioning: &blueprint.PartitioningCustomization{
				Plain: &blueprint.PlainFilesystemCustomization{
					Filesystems: []blueprint.FilesystemCustomization{
						{
							Mountpoint: "/dupydupe",
						},
						{
							Mountpoint: "/data",
						},
					},
				},
				LVM: &blueprint.LVMCustomization{
					VolumeGroups: []blueprint.VGCustomization{
						{
							LogicalVolumes: []blueprint.LVCustomization{
								{
									FilesystemCustomization: blueprint.FilesystemCustomization{Mountpoint: "/"},
								},
								{
									FilesystemCustomization: blueprint.FilesystemCustomization{Mountpoint: "/home"},
								},
								{
									FilesystemCustomization: blueprint.FilesystemCustomization{Mountpoint: "/dupydupe"},
								},
							},
						},
					},
				},
			},
			expected: fmt.Errorf("duplicate mountpoint \"/dupydupe\" in partitioning customizations"),
		},
		"unhappy-multibtrfs": {
			partitioning: &blueprint.PartitioningCustomization{
				Plain: &blueprint.PlainFilesystemCustomization{
					Filesystems: []blueprint.FilesystemCustomization{
						{
							Mountpoint: "/data",
						},
					},
				},
				Btrfs: &blueprint.BtrfsCustomization{
					Volumes: []blueprint.BtrfsVolumeCustomization{
						{
							Subvolumes: []blueprint.BtrfsSubvolumeCustomization{
								{
									Name:       "root",
									Mountpoint: "/",
								},
							},
						},
						{
							Subvolumes: []blueprint.BtrfsSubvolumeCustomization{
								{
									Name:       "home",
									Mountpoint: "/home",
								},
							},
						},
					},
				},
			},
			expected: fmt.Errorf("multiple btrfs volumes are not yet supported"),
		},
		"unhappy-multivg": {
			partitioning: &blueprint.PartitioningCustomization{
				Plain: &blueprint.PlainFilesystemCustomization{
					Filesystems: []blueprint.FilesystemCustomization{
						{
							Mountpoint: "/data",
						},
					},
				},
				LVM: &blueprint.LVMCustomization{
					VolumeGroups: []blueprint.VGCustomization{
						{
							LogicalVolumes: []blueprint.LVCustomization{
								{
									FilesystemCustomization: blueprint.FilesystemCustomization{Mountpoint: "/"},
								},
							},
						},
						{
							LogicalVolumes: []blueprint.LVCustomization{
								{
									FilesystemCustomization: blueprint.FilesystemCustomization{Mountpoint: "/var/log"},
								},
							},
						},
					},
				},
			},
			expected: fmt.Errorf("multiple LVM volume groups are not yet supported"),
		},
		"unhappy-emptymp": {
			partitioning: &blueprint.PartitioningCustomization{
				Plain: &blueprint.PlainFilesystemCustomization{
					Filesystems: []blueprint.FilesystemCustomization{
						{},
					},
				},
			},
			expected: fmt.Errorf("filesystem with empty mountpoint in partitioning customizations"),
		},
		"unhappy-emptymp-btrfs": {
			partitioning: &blueprint.PartitioningCustomization{
				Btrfs: &blueprint.BtrfsCustomization{
					Volumes: []blueprint.BtrfsVolumeCustomization{
						{
							Subvolumes: []blueprint.BtrfsSubvolumeCustomization{
								{
									Name:       "test",
									Mountpoint: "/test",
								},
								{
									Name:       "test2",
									Mountpoint: "",
								},
							},
						},
					},
				},
			},
			expected: fmt.Errorf("btrfs subvolume with empty mountpoint in partitioning customizations"),
		},
		"unhappy-emptymp-lvm": {
			partitioning: &blueprint.PartitioningCustomization{
				LVM: &blueprint.LVMCustomization{
					VolumeGroups: []blueprint.VGCustomization{
						{
							LogicalVolumes: []blueprint.LVCustomization{
								{
									Name: "testlv",
									FilesystemCustomization: blueprint.FilesystemCustomization{
										Mountpoint: "/stuff",
									},
								},
								{
									Name: "testlv2",
									FilesystemCustomization: blueprint.FilesystemCustomization{
										Mountpoint: "",
									},
								},
							},
						},
					},
				},
			},
			expected: fmt.Errorf("logical volume with empty mountpoint in partitioning customizations"),
		},
		"unhappy-dupesubvolname": {
			partitioning: &blueprint.PartitioningCustomization{
				Btrfs: &blueprint.BtrfsCustomization{
					Volumes: []blueprint.BtrfsVolumeCustomization{
						{
							Subvolumes: []blueprint.BtrfsSubvolumeCustomization{
								{
									Name:       "root",
									Mountpoint: "/",
								},
								{
									Name:       "root",
									Mountpoint: "/root",
								},
							},
						},
					},
				},
			},
			expected: fmt.Errorf("duplicate btrfs subvolume name \"root\" in partitioning customizations"),
		},
		"unhappy-dupelvname": {
			partitioning: &blueprint.PartitioningCustomization{
				LVM: &blueprint.LVMCustomization{
					VolumeGroups: []blueprint.VGCustomization{
						{
							LogicalVolumes: []blueprint.LVCustomization{
								{
									Name: "testlv",
									FilesystemCustomization: blueprint.FilesystemCustomization{
										Mountpoint: "/stuff",
									},
								},
								{
									Name: "testlv",
									FilesystemCustomization: blueprint.FilesystemCustomization{
										Mountpoint: "/stuff2",
									},
								},
							},
						},
					},
				},
			},
			expected: fmt.Errorf("duplicate lvm logical volume name \"testlv\" in volume group \"\" in partitioning customizations"),
		},
		"unhappy-emptyname-btrfs": {
			partitioning: &blueprint.PartitioningCustomization{
				Btrfs: &blueprint.BtrfsCustomization{
					Volumes: []blueprint.BtrfsVolumeCustomization{
						{
							Subvolumes: []blueprint.BtrfsSubvolumeCustomization{
								{
									Name:       "test",
									Mountpoint: "/test",
								},
								{
									Name:       "",
									Mountpoint: "/test2",
								},
							},
						},
					},
				},
			},
			expected: fmt.Errorf("btrfs subvolume with empty name in partitioning customizations"),
		},
	}

	for name := range testCases {
		tc := testCases[name]
		t.Run(name, func(t *testing.T) {
			err := tc.partitioning.Validate()
			assert.Equal(t, tc.expected, err)
		})
	}
}
