package manifest_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/osbuild/images/pkg/distro/inputs"
	"github.com/osbuild/images/pkg/manifest"
	"github.com/osbuild/images/pkg/osbuild"
	"github.com/osbuild/images/pkg/runner"
)

func TestTarSerialize(t *testing.T) {
	mani := manifest.New()
	runner := &runner.Linux{}
	repos := &inputs.RepoContainerConfig{}
	build := manifest.NewBuild(&mani, runner, repos, nil)

	// setup
	rawImage := manifest.NewRawImage(build, nil)
	tarPipeline := manifest.NewTar(build, rawImage, "tar-pipeline")
	tarPipeline.SetFilename("filename.tar")
	tarPipeline.Transform = "s/foo/bar"

	// run
	osbuildPipeline := tarPipeline.Serialize()

	// assert
	assert.Equal(t, "tar-pipeline", osbuildPipeline.Name)
	assert.Equal(t, 1, len(osbuildPipeline.Stages))
	tarStage := osbuildPipeline.Stages[0]
	assert.Equal(t, &osbuild.TarStageOptions{
		Filename:  "filename.tar",
		Transform: "s/foo/bar",
	}, tarStage.Options.(*osbuild.TarStageOptions))
}
