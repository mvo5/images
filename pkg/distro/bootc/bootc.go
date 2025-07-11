package bootc

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/osbuild/images/internal/common"
	"github.com/osbuild/images/pkg/arch"
	"github.com/osbuild/images/pkg/bib/container"
	"github.com/osbuild/images/pkg/bib/osinfo"
	"github.com/osbuild/images/pkg/distro"
	"github.com/osbuild/images/pkg/distro/generic"
	"github.com/osbuild/images/pkg/dnfjson"
	"github.com/osbuild/images/pkg/imagefilter"
	"github.com/osbuild/images/pkg/manifestgen"
	"github.com/osbuild/images/pkg/rpmmd"
	"github.com/osbuild/images/pkg/sbom"
)

func depsolverFuncFor(solver *dnfjson.Solver) manifestgen.DepsolveFunc {
	return func(cacheDir string, depsolveWarningsOutput io.Writer, packageSets map[string][]rpmmd.PackageSet, d distro.Distro, arch string) (map[string]dnfjson.DepsolveResult, error) {
		// XXX: duplicated from manifestgen extract a common
		// helper
		depsolvedSets := make(map[string]dnfjson.DepsolveResult)
		for name, pkgSet := range packageSets {
			res, err := solver.Depsolve(pkgSet, sbom.StandardTypeSpdx)
			if err != nil {
				return nil, fmt.Errorf("error depsolving: %w", err)
			}
			depsolvedSets[name] = *res
		}
		return depsolvedSets, nil
	}
}

func FromRef(bootcRef, imgTypeStr, archStr string) (*imagefilter.Result, manifestgen.DepsolveFunc, string, error) {
	cnt, err := container.New(bootcRef)
	if err != nil {
		return nil, nil, "", err
	}
	defer cnt.Stop()

	info, err := osinfo.Load(cnt.Root())
	if err != nil {
		return nil, nil, "", err
	}

	nameVer := fmt.Sprintf("bootc-%s-%s", info.OSRelease.ID, info.OSRelease.VersionID)
	distro, err := generic.NewDistroForBootc(nameVer, "./bootc")
	if err != nil {
		return nil, nil, "", err
	}
	a, err := distro.GetArch(archStr)
	if err != nil {
		return nil, nil, "", err
	}
	imgType, err := a.GetImageType(imgTypeStr)
	if err != nil {
		return nil, nil, "", err
	}

	// XXX: someone needs to clean this up :( lifecycle as it is
	// is a mess
	tmpdir, err := os.MkdirTemp("", "bootc-repos")
	if err != nil {
		return nil, nil, "", err
	}
	fakeRepoContent := fmt.Sprintf(`{"%s": [{"uri": "http://fake"}]}`, archStr)
	if err := os.WriteFile(filepath.Join(tmpdir, nameVer+".json"), []byte(fakeRepoContent), 0644); err != nil {
		return nil, nil, "", err
	}

	// XXX: add build container
	if err := cnt.InitDNF(); err != nil {
		return nil, nil, "", err
	}
	solver, err := cnt.NewContainerSolver("", common.Must(arch.FromString(archStr)), info)
	if err != nil {
		return nil, nil, "", err
	}

	return &imagefilter.Result{distro, a, imgType}, depsolverFuncFor(solver), tmpdir, nil
}
