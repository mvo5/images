package rhel10

// This file defines package sets that are used by more than one image type.

import (
	"github.com/osbuild/images/pkg/distro"
	"github.com/osbuild/images/pkg/distro/rhel"
	"github.com/osbuild/images/pkg/rpmmd"
)

func packageSetLoader(t *rhel.ImageType) (map[string]rpmmd.PackageSet, error) {
	return t.ImageTypeYAML.GetPackageSets(t.Arch().Distro().Name(), t.Arch().Name())
}

func imageConfig(d *rhel.Distribution, archName, imageType string) *distro.ImageConfig {
	it := d.ImageTypes[imageType]
	return it.GetImageConfig(d.Name(), archName)
}
