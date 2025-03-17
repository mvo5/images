package fedora

import (
	"github.com/osbuild/images/internal/common"
	"github.com/osbuild/images/pkg/distro"
	"github.com/osbuild/images/pkg/distro/packagesets"
	"github.com/osbuild/images/pkg/rpmmd"
)

func packageSetLoader(t *imageType) (rpmmd.PackageSet, error) {
	return packagesets.Load(t, "", VersionReplacements())
}

func imageConfigLoader(imageType string) *distro.ImageConfig {
	return common.Must(packagesets.LoadImageConfig("fedora", imageType, VersionReplacements()))
}
