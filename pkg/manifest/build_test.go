package manifest

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/osbuild/images/pkg/container"
	"github.com/osbuild/images/pkg/rpmmd"
	"github.com/osbuild/images/pkg/runner"
)

func TestBuildContainerBuildableNo(t *testing.T) {
	repos := []rpmmd.RepoConfig{}
	mf := New()
	runner := &runner.Fedora{Version: 39}

	build := NewBuild(&mf, runner, repos, nil)
	require.NotNil(t, build)

	for _, tc := range []struct {
		packageSpec           []rpmmd.PackageSpec
		containerBuildable    bool
		expectedSELinuxLabels map[string]string
	}{
		// no pkgs means no selinux labels (container build or not)
		{
			[]rpmmd.PackageSpec{},
			false,
			map[string]string{},
		},
		{
			[]rpmmd.PackageSpec{},
			true,
			map[string]string{},
		},
		{
			[]rpmmd.PackageSpec{{Name: "coreutils"}},
			false,
			map[string]string{
				"/usr/bin/cp": "system_u:object_r:install_exec_t:s0",
			},
		},
		{
			[]rpmmd.PackageSpec{{Name: "tar"}},
			false,
			map[string]string{
				"/usr/bin/tar": "system_u:object_r:install_exec_t:s0",
			},
		},
		{
			[]rpmmd.PackageSpec{{Name: "coreutils"}, {Name: "tar"}},
			false,
			map[string]string{
				"/usr/bin/cp":  "system_u:object_r:install_exec_t:s0",
				"/usr/bin/tar": "system_u:object_r:install_exec_t:s0",
			},
		},
		{
			[]rpmmd.PackageSpec{{Name: "coreutils"}},
			true,
			map[string]string{
				"/usr/bin/cp":     "system_u:object_r:install_exec_t:s0",
				"/usr/bin/mount":  "system_u:object_r:install_exec_t:s0",
				"/usr/bin/umount": "system_u:object_r:install_exec_t:s0",
			},
		},
		{
			[]rpmmd.PackageSpec{{Name: "coreutils"}, {Name: "tar"}},
			true,
			map[string]string{
				"/usr/bin/cp":     "system_u:object_r:install_exec_t:s0",
				"/usr/bin/mount":  "system_u:object_r:install_exec_t:s0",
				"/usr/bin/umount": "system_u:object_r:install_exec_t:s0",
				"/usr/bin/tar":    "system_u:object_r:install_exec_t:s0",
			},
		},
	} {
		build.packageSpecs = tc.packageSpec
		build.containerBuildable = tc.containerBuildable

		labels := build.getSELinuxLabels()
		require.Equal(t, labels, tc.expectedSELinuxLabels)
	}
}

func TestNewBuildFromContainerSpecs(t *testing.T) {
	containers := []container.Spec{
		{
			ImageID:   "sha256:511295a6eabcf0ca042017ed9e6561e4facd687b05910c0df56e4b11f4fb24b2",
			LocalName: "ghcr.io/ondrejbudai/bootc:centos",
		},
	}
	mf := New()
	runner := &runner.Fedora{Version: 39}

	build := NewBuildFromContainersSpec(&mf, runner, containers, nil)
	require.NotNil(t, build)

	osbuildPipeline := build.serialize()
	require.Len(t, osbuildPipeline.Stages, 2)
	assert.Equal(t, osbuildPipeline.Stages[0].Type, "org.osbuild.container-deploy")
	// TODO: find a better way to test that the inputs made it
	// into the pipeline than json :/
	json, err := json.Marshal(osbuildPipeline)
	require.Nil(t, err)
	require.Contains(t, string(json), `{"type":"org.osbuild.container-deploy","inputs":{"type":"org.osbuild.containers","origin":"org.osbuild.source","references":{"sha256:511295a6eabcf0ca042017ed9e6561e4facd687b05910c0df56e4b11f4fb24b2":{"name":"ghcr.io/ondrejbudai/bootc:centos"}}}`)
	require.Contains(t, string(json), `{"type":"org.osbuild.selinux","options":{"labels":{"/usr/bin/bootc":"system_u:object_r:install_exec_t:s0","/usr/bin/ostree":"system_u:object_r:install_exec_t:s0"}}}`)
}
