package manifest

import (
	"github.com/osbuild/images/pkg/osbuild"
	"github.com/osbuild/images/pkg/platform"
)

// A RawBootcImage represents a raw bootc image file which can be booted in a
// hypervisor.
type RawBootcImage struct {
	Base
	treePipeline *BootcDeployment
	filename     string
	platform     platform.Platform
}

func (p RawBootcImage) Filename() string {
	return p.filename
}

func (p *RawBootcImage) SetFilename(filename string) {
	p.filename = filename
}

func NewRawBootcImage(buildPipeline Build, treePipeline *BootcDeployment, platform platform.Platform) *RawBootcImage {
	p := &RawBootcImage{
		Base:         NewBase("image", buildPipeline),
		treePipeline: treePipeline,
		filename:     "disk.img",
		platform:     platform,
	}
	buildPipeline.addDependent(p)
	return p
}

func (p *RawBootcImage) getBuildPackages(Distro) []string {
	packages := p.platform.GetBuildPackages()
	packages = append(packages, p.platform.GetPackages()...)
	packages = append(packages, p.treePipeline.PartitionTable.GetBuildPackages()...)
	packages = append(packages,
		"rpm-ostree",

		// these should be defined on the platform
		"dracut-config-generic",
		"efibootmgr",
	)
	return packages
}

func (p *RawBootcImage) serialize() osbuild.Pipeline {
	pipeline := p.Base.serialize()

	pt := p.treePipeline.PartitionTable
	if pt == nil {
		panic("no partition table in live image")
	}

	for _, stage := range osbuild.GenImagePrepareStages(pt, p.Filename(), osbuild.PTSfdisk) {
		pipeline.AddStage(stage)
	}

	devices, mounts := osbuild.GenBootupdDevicesMounts(p.Filename(), p.treePipeline.PartitionTable)
	st, err := osbuild.NewBootcInstallToFilesystemStage(devices, mounts)
	if err != nil {
		panic(err)
	}
	pipeline.AddStage(st)

	for _, stage := range osbuild.GenImageFinishStages(pt, p.Filename()) {
		pipeline.AddStage(stage)
	}

	return pipeline
}
