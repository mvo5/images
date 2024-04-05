package osbuild_test

import (
	"encoding/json"
	"testing"

	"github.com/osbuild/images/pkg/osbuild"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBindMountDeploymentSerialized(t *testing.T) {
	mntStage := osbuild.NewBindMount("some-name", "mount://", "tree://")
	json, err := json.MarshalIndent(mntStage, "", "  ")
	require.Nil(t, err)
	assert.Equal(t, string(json), `
{
  "name": "some-name",
  "type": "org.osbuild.bind",
  "target": "tree://",
  "options": {
    "source": "mount://"
  }
}`[1:])
}
