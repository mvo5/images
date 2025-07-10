package bootc

import (
	"github.com/osbuild/images/pkg/distro/generic"
	"github.com/osbuild/images/pkg/imagefilter"
)

func FromRef(bootcRef, imgTypeStr, archStr string) (*imagefilter.Result, error) {
	// XXX: get by actually introspecting
	nameVer := "bootc-1"
	distro, err := generic.NewDistroForBootc(nameVer, "./bootc")
	if err != nil {
		return nil, err
	}
	a, err := distro.GetArch(archStr)
	if err != nil {
		return nil, err
	}
	imgType, err := a.GetImageType(imgTypeStr)
	if err != nil {
		return nil, err
	}

	return &imagefilter.Result{distro, a, imgType}, nil
}
