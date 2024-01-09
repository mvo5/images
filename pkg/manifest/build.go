package manifest

import (
	"github.com/osbuild/images/pkg/container"
	"github.com/osbuild/images/pkg/osbuild"
	"github.com/osbuild/images/pkg/ostree"
	"github.com/osbuild/images/pkg/rpmmd"
	"github.com/osbuild/images/pkg/runner"
)

// A Build represents the build environment for other pipelines. As a
// general rule, tools required to build pipelines are used from the build
// environment, rather than from the pipeline itself. Without a specified
// build environment, the build host's root filesystem would be used, which
// is not predictable nor reproducible. For the purposes of building the
// build pipeline, we do use the build host's filesystem, this means we should
// make minimal assumptions about what's available there.
type Build struct {
	Base

	runner     runner.Runner
	dependents []Pipeline
	repos      []rpmmd.RepoConfig

	containerBuildable bool
}

type BuildOptions struct {
	// ContainerBuildable tweaks the buildroot to be container friendly,
	// i.e. to not rely on an installed osbuild-selinux
	ContainerBuildable bool
}

// NewBuild creates a new build pipeline from the repositories in repos
// and the specified packages.
func NewBuild(m *Manifest, runner runner.Runner, repos []rpmmd.RepoConfig, opts *BuildOptions) *Build {
	if opts == nil {
		opts = &BuildOptions{}
	}

	name := "build"
	pipeline := &Build{
		Base:               NewBase(name, nil),
		runner:             runner,
		dependents:         make([]Pipeline, 0),
		repos:              filterRepos(repos, name),
		containerBuildable: opts.ContainerBuildable,
	}
	m.addPipeline(pipeline)
	return pipeline
}

func (p *Build) addDependent(dep Pipeline) {
	p.dependents = append(p.dependents, dep)
	man := p.Manifest()
	if man == nil {
		panic("cannot add build dependent without a manifest")
	}
	man.addPipeline(dep)
}

func (p *Build) getPackageSetChain(distro Distro) []rpmmd.PackageSet {
	// TODO: make the /usr/bin/cp dependency conditional
	// TODO: make the /usr/bin/xz dependency conditional
	packages := []string{
		"selinux-policy-targeted", // needed to build the build pipeline
		"coreutils",               // /usr/bin/cp - used all over
		"xz",                      // usage unclear
	}

	packages = append(packages, p.runner.GetBuildPackages()...)

	for _, pipeline := range p.dependents {
		packages = append(packages, pipeline.getBuildPackages(distro)...)
	}

	return []rpmmd.PackageSet{
		{
			Include:         packages,
			Repositories:    p.repos,
			InstallWeakDeps: true,
		},
	}
}

func (p *Build) serialize2(inputs *SerializeInputs) (osbuild.Pipeline, *SerializeOutputs) {
	packageSpecs := inputs.packages

	pipeline := p.Base.serialize()
	pipeline.Runner = p.runner.String()

	pipeline.AddStage(osbuild.NewRPMStage(osbuild.NewRPMStageOptions(p.repos), osbuild.NewRpmStageSourceFilesInputs(packageSpecs)))
	pipeline.AddStage(osbuild.NewSELinuxStage(&osbuild.SELinuxStageOptions{
		FileContexts: "etc/selinux/targeted/contexts/files/file_contexts",
		Labels:       p.getSELinuxLabels(packageSpecs),
	},
	))
	outputs := &SerializeOutputs{
		// XXX: manifest.go always use the input packages so maybe
		//      don't need this?
		packages: inputs.packages,
	}

	return pipeline, outputs
}

func (p *Build) getPackageSpecs() []rpmmd.PackageSpec {
	return nil
}

func (p *Build) serializeStart(packages []rpmmd.PackageSpec, _ []container.Spec, _ []ostree.CommitSpec) {
}

func (p *Build) serializeEnd() {
}

func (p *Build) serialize() osbuild.Pipeline {
	return osbuild.Pipeline{}
}

// Returns a map of paths to labels for the SELinux stage based on specific
// packages found in the pipeline.
func (p *Build) getSELinuxLabels(packages []rpmmd.PackageSpec) map[string]string {
	labels := make(map[string]string)
	for _, pkg := range packages {
		switch pkg.Name {
		case "coreutils":
			labels["/usr/bin/cp"] = "system_u:object_r:install_exec_t:s0"
			if p.containerBuildable {
				labels["/usr/bin/mount"] = "system_u:object_r:install_exec_t:s0"
				labels["/usr/bin/umount"] = "system_u:object_r:install_exec_t:s0"
			}
		case "tar":
			labels["/usr/bin/tar"] = "system_u:object_r:install_exec_t:s0"
		}
	}
	return labels
}
