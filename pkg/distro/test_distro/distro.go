package test_distro

import (
	"github.com/osbuild/images/pkg/distro/generic"
)

const (
	// The Test Distro name base. It can't be used to get a distro.Distro
	// instance from the DistroFactory(), because it does not include any
	// release version.
	TestDistroNameBase = "test-distro"

	// An ID string for a Test Distro instance with release version 1.
	TestDistro1Name = TestDistroNameBase + "-1"

	TestArchName  = "x86_64"
	TestArch2Name = "aarch64"
	TestArch3Name = "s390x"

	TestImageTypeName   = "test_type"
	TestImageType2Name  = "test_type2"
	TestImageTypeOSTree = "test_ostree_type"

	// added for cloudapi tests
	TestImageTypeAmi            = "ami"
	TestImageTypeGce            = "gce"
	TestImageTypeVhd            = "vhd"
	TestImageTypeEdgeCommit     = "rhel-edge-commit"
	TestImageTypeEdgeInstaller  = "rhel-edge-installer"
	TestImageTypeImageInstaller = "image-installer"
	TestImageTypeQcow2          = "qcow2"
	TestImageTypeVmdk           = "vmdk"
)

// for compat with the exiting clients
var DistroFactory = generic.DistroFactory
