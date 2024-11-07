package manifest

import (
	"fmt"

	"github.com/osbuild/images/pkg/container"
	"github.com/osbuild/images/pkg/osbuild"
	"github.com/osbuild/images/pkg/ostree"
	"github.com/osbuild/images/pkg/platform"
	"github.com/osbuild/images/pkg/rpmmd"
)

// OSBootc represents the filesystem tree of the target image for a bootc based
// system.
type OSBootc struct {
	Base
	platform platform.Platform

	containers     []container.SourceSpec
	containerSpecs []container.Spec

	KernelOptionsAppend []string
	WSLConfig           *osbuild.WSLConfStageOptions
}

func NewOSBootc(buildPipeline Build, containers []container.SourceSpec, platform platform.Platform) *OSBootc {
	p := &OSBootc{
		Base:     NewBase("os", buildPipeline),
		platform: platform,

		containers: containers,
	}
	buildPipeline.addDependent(p)
	return p
}

func (p *OSBootc) getContainerSources() []container.SourceSpec {
	return p.containers
}

func (p *OSBootc) getContainerSpecs() []container.Spec {
	return p.containerSpecs
}

func (p *OSBootc) serializeStart(_ []rpmmd.PackageSpec, containerSpecs []container.Spec, _ []ostree.CommitSpec, _ []rpmmd.RepoConfig) {
	if len(p.containerSpecs) > 0 {
		panic("double call to serializeStart()")
	}
	p.containerSpecs = containerSpecs
}

func (p *OSBootc) serializeEnd() {
	if len(p.containerSpecs) == 0 {
		panic("serializeEnd() call when serialization not in progress")
	}
	p.containerSpecs = nil
}

func (p *OSBootc) serialize() osbuild.Pipeline {
	pipeline := p.Base.serialize()

	if len(p.containerSpecs) != 1 {
		panic(fmt.Errorf("expected a single container input got %v", p.containerSpecs))
	}

	opts := &osbuild.BootcInstallToFilesystemOptions{
		Kargs: p.KernelOptionsAppend,
	}
	inputs := osbuild.ContainerDeployInputs{
		Images: osbuild.NewContainersInputForSingleSource(p.containerSpecs[0]),
	}
	st, err := osbuild.NewBootcInstallToFilesystemStage(opts, inputs, nil, nil, p.platform)
	if err != nil {
		panic(err)
	}
	pipeline.AddStage(st)

	if p.WSLConfig != nil {
		pipeline.AddStage(osbuild.NewWSLConfStage(p.WSLConfig))
	}

	// XXX: add support for users/groups/selinux, see raw_bootc

	return pipeline
}
