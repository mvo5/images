package manifest

import (
	"fmt"

	"github.com/osbuild/images/pkg/osbuild"
)

type BootcCustomize struct {
	Base

	imagePipeline *RawBootcImage
}

func NewBootcCustomize(buildPipeline Build, imagePipeline *RawBootcImage) *BootcCustomize {

	p := &BootcCustomize{
		Base: NewBase("bootc-customize", buildPipeline),

		imagePipeline: imagePipeline,
	}
	buildPipeline.addDependent(p)
	return p
}

func findRootMntName(mounts []osbuild.Mount) string {
	for _, mnt := range mounts {
		if mnt.Target == "/" {
			return mnt.Name
		}
	}
	return ""
}

func (p *BootcCustomize) serialize() osbuild.Pipeline {
	pipeline := p.Base.serialize()

	mkdirStage := osbuild.NewMkdirStage(&osbuild.MkdirStageOptions{
		Paths: []osbuild.MkdirStagePath{
			{
				Path: "/etc",
			},
		},
	})
	pipeline.AddStage(mkdirStage)

	fstabOptions := osbuild.NewFSTabStageOptions(p.imagePipeline.PartitionTable)
	fstabStage := osbuild.NewFSTabStage(fstabOptions)
	pipeline.AddStage(fstabStage)

	// XXX: use osbuild.GenCopyFSTreeOptions() here instead?
	devices, mounts := osbuild.GenBootupdDevicesMounts(p.imagePipeline.Filename(), p.imagePipeline.PartitionTable)
	rootfsMntName := findRootMntName(mounts)
	if rootfsMntName == "" {
		panic(fmt.Sprintf("cannot find root mount in %v", mounts))
	}
	opts := &osbuild.CopyStageOptions{
		Paths: []osbuild.CopyStagePath{
			{
				From: "tree:///etc/fstab",
				To:   fmt.Sprintf("mount://%s/etc/fstab", rootfsMntName),
			},
		},
	}
	inputs := osbuild.NewPipelineTreeInputs("image", p.imagePipeline.Name())
	copyStage := osbuild.NewCopyStage(opts, inputs, devices, mounts)
	pipeline.AddStage(copyStage)

	return pipeline
}
