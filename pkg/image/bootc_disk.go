package image

import (
	"fmt"
	"math/rand"

	"github.com/osbuild/images/pkg/artifact"
	"github.com/osbuild/images/pkg/container"
	"github.com/osbuild/images/pkg/disk"
	"github.com/osbuild/images/pkg/manifest"
	"github.com/osbuild/images/pkg/platform"
	"github.com/osbuild/images/pkg/runner"
)

type BootcDiskImage struct {
	Base

	Platform       platform.Platform
	PartitionTable *disk.PartitionTable

	OSName string

	Filename string

	ContainerSource *container.SourceSpec

	Compression string
}

func NewBootcDiskImage(container container.SourceSpec) *BootcDiskImage {
	return &BootcDiskImage{
		Base:            NewBase("bootc-raw-image"),
		ContainerSource: &container,
		OSName:          "default",
	}
}

func (img *BootcDiskImage) InstantiateManifestFromContainers(m *manifest.Manifest,
	containers []container.SourceSpec,
	runner runner.Runner,
	rng *rand.Rand) (*artifact.Artifact, error) {

	buildPipeline := manifest.NewBuildFromContainer(m, runner, containers, &manifest.BuildOptions{ContainerBuildable: true})
	buildPipeline.Checkpoint()

	// don't support compressing non-raw images
	imgFormat := img.Platform.GetImageFormat()
	if imgFormat == platform.FORMAT_UNSET {
		// treat unset as raw for this check
		imgFormat = platform.FORMAT_RAW
	}
	if imgFormat != platform.FORMAT_RAW && img.Compression != "" {
		panic(fmt.Sprintf("no compression is allowed with %q format for %q", imgFormat, img.name))
	}

	osPipeline := manifest.NewBootcDeployment(buildPipeline, img.ContainerSource, img.OSName, img.Platform)
	osPipeline.PartitionTable = img.PartitionTable
	baseImage := manifest.NewRawBootcImage(buildPipeline, containers, osPipeline, img.Platform)

	switch imgFormat {
	case platform.FORMAT_QCOW2:
		// qcow2 runs without a build pipeline directly from "bib"
		qcow2Pipeline := manifest.NewQCOW2(nil, baseImage)
		qcow2Pipeline.Compat = img.Platform.GetQCOW2Compat()
		qcow2Pipeline.SetFilename(img.Filename)
		return qcow2Pipeline.Export(), nil
	}

	switch img.Compression {
	case "xz":
		compressedImage := manifest.NewXZ(buildPipeline, baseImage)
		compressedImage.SetFilename(img.Filename)
		return compressedImage.Export(), nil
	case "":
		baseImage.SetFilename(img.Filename)
		return baseImage.Export(), nil
	default:
		panic(fmt.Sprintf("unsupported compression type %q on %q", img.Compression, img.name))
	}
}
