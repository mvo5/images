package osbuild_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/osbuild/images/pkg/container"
	"github.com/osbuild/images/pkg/osbuild"
)

func TestContainersDeployStageRenders(t *testing.T) {
	inputs := osbuild.NewContainersInputForSources([]container.Spec{
		{
			ImageID: "id-0",
			Source:  "registry.example.org/reg/img",
		},
	})
	stage, err := osbuild.NewContainerDeployStage(inputs)
	require.NotNil(t, stage)
	require.Nil(t, err)

	assert.Equal(t, stage.Type, "org.osbuild.container-deploy")
	assert.Equal(t, stage.Inputs.(osbuild.ContainerDeployInputs).Images, inputs)
}

func TestContainersDeployStageInputsValidate(t *testing.T) {
	type testCase struct {
		inputs osbuild.ContainerDeployInputs
		err    string
	}

	testCases := map[string]testCase{
		"empty": {
			inputs: osbuild.ContainerDeployInputs{},
			err:    "stage requires exactly 1 input container (got nil References)",
		},
		"nil": {
			inputs: osbuild.ContainerDeployInputs{
				Images: osbuild.ContainersInput{
					References: nil,
				},
			},
			err: "stage requires exactly 1 input container (got nil References)",
		},
		"zero": {
			inputs: osbuild.ContainerDeployInputs{
				Images: osbuild.NewContainersInputForSources([]container.Spec{}),
			},
			err: "stage requires exactly 1 input container (got 0)",
		},
		"one": {
			inputs: osbuild.ContainerDeployInputs{
				Images: osbuild.NewContainersInputForSources([]container.Spec{
					{
						ImageID: "id-0",
						Source:  "registry.example.org/reg/img",
					},
				}),
			},
		},
		"two": {
			inputs: osbuild.ContainerDeployInputs{
				Images: osbuild.NewContainersInputForSources([]container.Spec{
					{
						ImageID: "id-1",
						Source:  "registry.example.org/reg/img-one",
					},
					{
						ImageID: "id-2",
						Source:  "registry.example.org/reg/img-two",
					},
				}),
			},
			err: "stage requires exactly 1 input container (got 2)",
		},
	}
	for name := range testCases {
		tc := testCases[name]
		t.Run(name, func(t *testing.T) {
			stage, err := osbuild.NewContainerDeployStage(tc.inputs.Images)
			if tc.err == "" {
				require.NotNil(t, stage)
				require.Nil(t, err)
			} else {
				require.Nil(t, stage)
				assert.ErrorContains(t, err, tc.err)
			}
		})
	}
}
