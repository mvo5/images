package manifest_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/osbuild/images/pkg/manifest"
	"github.com/osbuild/images/pkg/runner"
)

func TestPipelineExports(t *testing.T) {
	var mf manifest.Manifest

	build := manifest.NewBuild(&mf, &runner.Linux{}, nil, nil)

	manifest.NewSubscription(build, nil)
	manifest.NewXZ(build, nil)
	assert.Equal(t, []string{"xz"}, mf.Exports())
}
