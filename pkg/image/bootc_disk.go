package image

import (
	"fmt"
	"math/rand"
	"path/filepath"
	"strings"

	"github.com/osbuild/images/pkg/artifact"
	"github.com/osbuild/images/pkg/container"
	"github.com/osbuild/images/pkg/customizations/fsnode"
	"github.com/osbuild/images/pkg/disk"
	"github.com/osbuild/images/pkg/manifest"
	"github.com/osbuild/images/pkg/osbuild"
	"github.com/osbuild/images/pkg/platform"
	"github.com/osbuild/images/pkg/rpmmd"
	"github.com/osbuild/images/pkg/runner"
)

type BootcDiskImage struct {
	Base

	Platform       platform.Platform
	PartitionTable *disk.PartitionTable

	Filename    string
	Compression string

	// Control the VPC subformat use of force_size
	VPCForceSize *bool

	ContainerSource      *container.SourceSpec
	BuildContainerSource *container.SourceSpec

	// Customizations
	OSCustomizations manifest.OSCustomizations
}

func NewBootcDiskImage(container container.SourceSpec, buildContainer container.SourceSpec) *BootcDiskImage {
	return &BootcDiskImage{
		Base:                 NewBase("bootc-raw-image"),
		ContainerSource:      &container,
		BuildContainerSource: &buildContainer,
	}
}

func (img *BootcDiskImage) InstantiateManifest(m *manifest.Manifest,
	repos []rpmmd.RepoConfig,
	runner runner.Runner,
	rng *rand.Rand) (*artifact.Artifact, error) {

	return nil, fmt.Errorf("internal error: BootcDiskImage  only supported InstantiateManifestFromContainers")
}

func (img *BootcDiskImage) InstantiateManifestFromContainers(m *manifest.Manifest,
	containers []container.SourceSpec,
	runner runner.Runner,
	rng *rand.Rand) (*artifact.Artifact, error) {

	policy := img.OSCustomizations.SElinux
	if img.OSCustomizations.BuildSElinux != "" {
		policy = img.OSCustomizations.BuildSElinux
	}

	var copyFilesFrom map[string][]string
	var ensureDirs []*fsnode.Directory

	if *img.ContainerSource != *img.BuildContainerSource {
		// If we're using a different build container from the target container then we copy
		// the bootc customization file directories from the target container. This includes the
		// bootc install customization, and /usr/lib/ostree/prepare-root.conf which configures
		// e.g. composefs and fs-verity setup.
		//
		// To ensure that these copies never fail we also create the source and target
		// directories as needed.

		pipelineName := "target"
		// files to copy have slash at end to copy directory contents, not directory itself
		copyFiles := []string{"/usr/lib/bootc/install/", "/usr/lib/ostree/"}
		ensureDirPaths := []string{"/usr/lib/bootc/install", "/usr/lib/ostree"}

		copyFilesFrom = map[string][]string{pipelineName: copyFiles}
		for _, path := range ensureDirPaths {
			// Note: Mode/User/Group must be nil here to make  GenDirectoryNodesStages use dirExistOk
			dir, err := fsnode.NewDirectory(path, nil, nil, nil, true)
			if err != nil {
				return nil, err
			}
			ensureDirs = append(ensureDirs, dir)
		}

		targetContainers := []container.SourceSpec{*img.ContainerSource}
		targetBuildPipeline := manifest.NewBuildFromContainer(m, runner, targetContainers,
			&manifest.BuildOptions{
				PipelineName:       pipelineName,
				ContainerBuildable: true,
				SELinuxPolicy:      policy,
				EnsureDirs:         ensureDirs,
			})
		targetBuildPipeline.Checkpoint()
	}

	buildContainers := []container.SourceSpec{*img.BuildContainerSource}
	buildPipeline := manifest.NewBuildFromContainer(m, runner, buildContainers,
		&manifest.BuildOptions{
			ContainerBuildable: true,
			SELinuxPolicy:      policy,
			CopyFilesFrom:      copyFilesFrom,
			EnsureDirs:         ensureDirs,
		})

	buildPipeline.Checkpoint()

	// In the bootc flow, we reuse the host container context for tools;
	// this is signified by passing nil to the below pipelines.
	// The reason is that we use the bootc container as the buildroot
	// and because it is bootc we cannot install extra build tools.
	var hostPipeline manifest.Build

	rawImagePipeline := manifest.NewRawBootcImage(buildPipeline, containers, img.Platform)
	rawImagePipeline.PartitionTable = img.PartitionTable
	rawImagePipeline.Users = img.OSCustomizations.Users
	rawImagePipeline.Groups = img.OSCustomizations.Groups
	rawImagePipeline.Files = img.OSCustomizations.Files
	rawImagePipeline.Directories = img.OSCustomizations.Directories
	rawImagePipeline.KernelOptionsAppend = img.OSCustomizations.KernelOptionsAppend
	rawImagePipeline.SELinux = img.OSCustomizations.SElinux
	rawImagePipeline.MountUnits = true // always use mount units for bootc disk images

	// XXX: duplicated with disk.go, ostree_disk.go, etc
	var imagePipeline manifest.FilePipeline
	switch img.Platform.GetImageFormat() {
	case platform.FORMAT_RAW:
		imagePipeline = rawImagePipeline
	case platform.FORMAT_QCOW2:
		qcow2Pipeline := manifest.NewQCOW2(hostPipeline, rawImagePipeline)
		qcow2Pipeline.Compat = img.Platform.GetQCOW2Compat()
		imagePipeline = qcow2Pipeline
	case platform.FORMAT_VAGRANT_LIBVIRT:
		qcow2Pipeline := manifest.NewQCOW2(hostPipeline, rawImagePipeline)
		qcow2Pipeline.Compat = img.Platform.GetQCOW2Compat()

		vagrantPipeline := manifest.NewVagrant(hostPipeline, qcow2Pipeline)

		tarPipeline := manifest.NewTar(hostPipeline, vagrantPipeline, "archive")
		tarPipeline.Format = osbuild.TarArchiveFormatUstar

		imagePipeline = tarPipeline
	case platform.FORMAT_VHD:
		vpcPipeline := manifest.NewVPC(hostPipeline, rawImagePipeline)
		vpcPipeline.ForceSize = img.VPCForceSize
		imagePipeline = vpcPipeline
	case platform.FORMAT_VMDK:
		imagePipeline = manifest.NewVMDK(hostPipeline, rawImagePipeline)
	case platform.FORMAT_OVA:
		vmdkPipeline := manifest.NewVMDK(hostPipeline, rawImagePipeline)
		ovfPipeline := manifest.NewOVF(hostPipeline, vmdkPipeline)
		tarPipeline := manifest.NewTar(hostPipeline, ovfPipeline, "archive")
		tarPipeline.Format = osbuild.TarArchiveFormatUstar
		tarPipeline.SetFilename(img.Filename)
		extLess := strings.TrimSuffix(img.Filename, filepath.Ext(img.Filename))
		// The .ovf descriptor needs to be the first file in the archive
		tarPipeline.Paths = []string{
			fmt.Sprintf("%s.ovf", extLess),
			fmt.Sprintf("%s.mf", extLess),
			fmt.Sprintf("%s.vmdk", extLess),
		}
		imagePipeline = tarPipeline
	case platform.FORMAT_GCE:
		// NOTE(akoutsou): temporary workaround; filename required for GCP
		// TODO: define internal raw filename on image type
		rawImagePipeline.SetFilename("disk.raw")
		tarPipeline := newGCETarPipelineForImg(buildPipeline, rawImagePipeline, "archive")
		imagePipeline = tarPipeline
	default:
		panic(fmt.Errorf("invalid image format %q for image kind", img.Platform.GetImageFormat()))
	}

	compressionPipeline := GetCompressionPipeline(img.Compression, hostPipeline, imagePipeline)
	compressionPipeline.SetFilename(img.Filename)

	return compressionPipeline.Export(), nil
}
