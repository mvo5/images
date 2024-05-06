package image_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/osbuild/images/pkg/container"
	"github.com/osbuild/images/pkg/image"
	"github.com/osbuild/images/pkg/manifest"
	"github.com/osbuild/images/pkg/platform"
	"github.com/osbuild/images/pkg/runner"
)

var fakeCntSource = container.SourceSpec{
	Source: "source-spec",
	Name:   "name",
}

var fakeRng = rand.New(rand.NewSource(0)) // nolint:gosec

func TestAnacondaContainerInstallerNew(t *testing.T) {
	img := image.NewAnacondaContainerInstaller(fakeCntSource, "some/ref")
	require.NotNil(t, img)
	assert.Equal(t, img.Base.Name(), "container-installer")
}

func TestAnacondaContainerInstallerManifestWithUsers(t *testing.T) {
	img := image.NewAnacondaContainerInstaller(fakeCntSource, "some/ref")
	img.Platform = &platform.X86{}

	m := &manifest.Manifest{}
	runi := &runner.Fedora{}
	_, err := img.InstantiateManifest(m, nil, runi, fakeRng)
	require.Nil(t, err)

	for _, p := range m.Pipelines() {
		fmt.Printf("%v\n", p)
	}

}
