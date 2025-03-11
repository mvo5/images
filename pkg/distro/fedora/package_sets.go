package fedora

import (
	"github.com/osbuild/images/pkg/distro/yamlloader"
	"github.com/osbuild/images/pkg/rpmmd"
)

func packageSetLoader(t *imageType) rpmmd.PackageSet {
	return yamlloader.LoadPkgSet(t, "", VersionReplacements())
}
