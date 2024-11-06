package main_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/osbuild/images/cmd/image-builder"
	"github.com/osbuild/images/pkg/distrofactory"
	"github.com/osbuild/images/pkg/imagefilter"
)

func TestFitleredResultFormatter(t *testing.T) {
	fac := distrofactory.NewTestDefault()

	for _, tc := range []struct {
		formatter             string
		distro, arch, imgType string
		expectsOutput         string
	}{
		{"", "test-distro-1", "test_arch3", "qcow2", `test-distro-1 arch:test_arch3 type:qcow2` + "\n"},
		{"text", "test-distro-1", "test_arch3", "qcow2", `test-distro-1 arch:test_arch3 type:qcow2` + "\n"},
		{"json", "test-distro-1", "test_arch3", "qcow2", `[{"distro":{"name":"test-distro-1"},"arch":{"name":"test_arch3"},"image_type":{"name":"qcow2"}}]` + "\n"},
	} {
		// XXX: it would be nice if TestDistro would support constructing
		// like GetDistro("rhel-8.1:i386,amd64:ami,qcow2") instead of
		// the current very static setup
		di := fac.GetDistro(tc.distro)
		require.NotNil(t, di)
		ar, err := di.GetArch(tc.arch)
		require.NoError(t, err)
		im, err := ar.GetImageType(tc.imgType)
		require.NoError(t, err)

		var buf bytes.Buffer
		res := []imagefilter.Result{
			{Distro: di, Arch: ar, ImgType: im},
		}
		fmter, err := main.NewFilteredResultFormatter(tc.formatter)
		require.NoError(t, err)
		err = fmter.Output(&buf, res)
		assert.NoError(t, err)
		assert.Equal(t, tc.expectsOutput, buf.String(), tc)
	}
}
