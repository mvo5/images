package rhel9

import (
	"github.com/osbuild/images/pkg/distro/rhel"
	"github.com/osbuild/images/pkg/rpmmd"
)

func packageSetLoader(t *rhel.ImageType) (map[string]rpmmd.PackageSet, error) {
	return t.ImageTypeYAML.GetPackageSets(t.Arch().Distro().Name(), t.Arch().Name())
}
