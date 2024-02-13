package manifest

import (
	"fmt"
	"os"

	"github.com/osbuild/images/internal/common"
	"github.com/osbuild/images/pkg/container"
	"github.com/osbuild/images/pkg/disk"
	"github.com/osbuild/images/pkg/osbuild"
	"github.com/osbuild/images/pkg/ostree"
	"github.com/osbuild/images/pkg/platform"
	"github.com/osbuild/images/pkg/rpmmd"
)

// BootcDeployment represents a deployment via bootc
type BootcDeployment struct {
	Base

	// XXX: do we need this?
	osName string

	// containerSource represents the source that will be used to retrieve the
	// bootc container for this pipeline.
	containerSource *container.SourceSpec

	// containerSpec is the resolved bootc container that will be
	// deployed in this pipeline.
	containerSpec *container.Spec

	platform       platform.Platform
	PartitionTable *disk.PartitionTable
}

// NewBootcDeployment creates a pipeline for a bootc deployment from a
// container
func NewBootcDeployment(buildPipeline Build,
	container *container.SourceSpec,
	osName string,
	platform platform.Platform) *BootcDeployment {

	p := &BootcDeployment{
		Base:            NewBase("bootc-deployment", buildPipeline),
		containerSource: container,
		osName:          osName,
		platform:        platform,
	}
	buildPipeline.addDependent(p)
	return p
}

func (p *BootcDeployment) getContainerSpecs() []container.Spec {
	if p.containerSpec == nil {
		return []container.Spec{}
	}
	return []container.Spec{*p.containerSpec}
}

func (p *BootcDeployment) getContainerSources() []container.SourceSpec {
	if p.containerSource == nil {
		return []container.SourceSpec{}
	}
	return []container.SourceSpec{
		*p.containerSource,
	}
}

func (p *BootcDeployment) serializeStart(packages []rpmmd.PackageSpec, containers []container.Spec, commits []ostree.CommitSpec) {
	if p.containerSpec != nil {
		panic("double call to serializeStart()")
	}

	if len(containers) != 1 {
		panic(fmt.Sprintf("pipeline %s requires exactly one container (have containers: %v)", p.Name(), containers))
	}
	p.containerSpec = &containers[0]
}

func (p *BootcDeployment) serializeEnd() {
	if p.containerSpec == nil {
		panic("serializeEnd() call when serialization not in progress")
	}
	p.containerSpec = nil
}

func (p *BootcDeployment) serialize() osbuild.Pipeline {
	if p.containerSpec == nil {
		panic("serialization not started")
	}

	pipeline := p.Base.serialize()

	pipeline.AddStage(osbuild.NewMkdirStage(&osbuild.MkdirStageOptions{
		Paths: []osbuild.MkdirStagePath{
			{
				Path: "/boot/efi",
				Mode: common.ToPtr(os.FileMode(0700)),
			},
		},
	}))
	fstabOptions := osbuild.NewFSTabStageOptions(p.PartitionTable)
	fstabStage := osbuild.NewFSTabStage(fstabOptions)
	pipeline.AddStage(fstabStage)

	return pipeline
}
