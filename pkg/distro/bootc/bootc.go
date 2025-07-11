package bootc

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/osbuild/images/pkg/bib/container"
	"github.com/osbuild/images/pkg/bib/osinfo"
	"github.com/osbuild/images/pkg/distro/generic"
	"github.com/osbuild/images/pkg/imagefilter"
)

func FromRef(bootcRef, imgTypeStr, archStr string) (*imagefilter.Result, string, error) {
	cnt, err := container.New(bootcRef)
	if err != nil {
		return nil, "", err
	}
	defer cnt.Stop()

	info, err := osinfo.Load(cnt.Root())
	if err != nil {
		return nil, "", err
	}

	nameVer := fmt.Sprintf("bootc-%s-%s", info.OSRelease.ID, info.OSRelease.VersionID)
	distro, err := generic.NewDistroForBootc(nameVer, "./bootc")
	if err != nil {
		return nil, "", err
	}
	a, err := distro.GetArch(archStr)
	if err != nil {
		return nil, "", err
	}
	imgType, err := a.GetImageType(imgTypeStr)
	if err != nil {
		return nil, "", err
	}

	// XXX: someone needs to clean this up :( lifecycle as it is
	// is a mess
	tmpdir, err := os.MkdirTemp("", "bootc-repos")
	if err != nil {
		return nil, "", err
	}
	fakeRepoContent := fmt.Sprintf(`{"%s": [{"uri": "http://fake"}]}`, archStr)
	if err := os.WriteFile(filepath.Join(tmpdir, nameVer+".json"), []byte(fakeRepoContent), 0644); err != nil {
		return nil, "", err
	}

	return &imagefilter.Result{distro, a, imgType}, tmpdir, nil
}
