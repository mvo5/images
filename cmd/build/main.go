// Standalone executable for building a test image.
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/osbuild/images/internal/buildconfig"
	"github.com/osbuild/images/internal/cmdutil"
	"github.com/osbuild/images/pkg/arch"
	"github.com/osbuild/images/pkg/distrofactory"
	"github.com/osbuild/images/pkg/manifest"
	"github.com/osbuild/images/pkg/manifestgen"
	"github.com/osbuild/images/pkg/osbuild"
	"github.com/osbuild/images/pkg/reporegistry"
	"github.com/osbuild/images/pkg/rhsm/facts"
	"github.com/osbuild/images/pkg/rpmmd"
)

func save(ms manifest.OSBuildManifest, fpath string) error {
	b, err := json.MarshalIndent(ms, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal data for %q: %w", fpath, err)
	}
	b = append(b, '\n') // add new line at end of file
	fp, err := os.Create(fpath)
	if err != nil {
		return fmt.Errorf("failed to create output file %q: %w", fpath, err)
	}
	defer fp.Close()
	if _, err := fp.Write(b); err != nil {
		return fmt.Errorf("failed to write output file %q: %w", fpath, err)
	}
	return nil
}

func u(s string) string {
	return strings.Replace(s, "-", "_", -1)
}

func run() error {
	// common args
	var outputDir, osbuildStore, rpmCacheRoot, repositories string
	flag.StringVar(&outputDir, "output", ".", "artifact output directory")
	flag.StringVar(&osbuildStore, "store", ".osbuild", "osbuild store for intermediate pipeline trees")
	flag.StringVar(&rpmCacheRoot, "rpmmd", "/tmp/rpmmd", "rpm metadata cache directory")
	flag.StringVar(&repositories, "repositories", "test/data/repositories", "path to repository file or directory")

	// osbuild checkpoint arg
	var checkpoints cmdutil.MultiValue
	flag.Var(&checkpoints, "checkpoints", "comma-separated list of pipeline names to checkpoint (passed to osbuild --checkpoint)")

	// image selection args
	var distroName, imgTypeName, configFile string
	flag.StringVar(&distroName, "distro", "", "distribution (required)")
	flag.StringVar(&imgTypeName, "type", "", "image type name (required)")
	flag.StringVar(&configFile, "config", "", "build config file (required)")

	flag.Parse()

	if distroName == "" || imgTypeName == "" || configFile == "" {
		flag.Usage()
		os.Exit(1)
	}

	distroFac := distrofactory.NewDefault()

	config, err := buildconfig.New(configFile)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(outputDir, 0777); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	distribution := distroFac.GetDistro(distroName)
	if distribution == nil {
		return fmt.Errorf("invalid or unsupported distribution: %q", distroName)
	}

	archName := arch.Current().String()
	arch, err := distribution.GetArch(archName)
	if err != nil {
		return fmt.Errorf("invalid arch name %q for distro %q: %w", archName, distroName, err)
	}

	buildName := fmt.Sprintf("%s-%s-%s-%s", u(distroName), u(archName), u(imgTypeName), u(config.Name))
	buildDir := filepath.Join(outputDir, buildName)
	if err := os.MkdirAll(buildDir, 0777); err != nil {
		return fmt.Errorf("failed to create target directory: %w", err)
	}

	imgType, err := arch.GetImageType(imgTypeName)
	if err != nil {
		return fmt.Errorf("invalid image type %q for distro %q and arch %q: %w", imgTypeName, distroName, archName, err)
	}

	var reporeg *reporegistry.RepoRegistry
	var customRepos []rpmmd.RepoConfig
	if st, err := os.Stat(repositories); err == nil && st.Mode().IsRegular() {
		repoConfig, err := rpmmd.LoadRepositoriesFromFile(repositories)
		if err != nil {
			return fmt.Errorf("failed to load repositories from %q: %w", repositories, err)
		}
		customRepos = repoConfig[archName]
	} else {
		reporeg, err = reporegistry.New([]string{repositories})
		if err != nil {
			return fmt.Errorf("failed to load repositories from %q: %w", repositories, err)
		}
	}

	fmt.Printf("Generating manifest for %s: ", config.Name)
	var mf bytes.Buffer
	manifestOpts := manifestgen.Options{
		Output:      &mf,
		Cachedir:    filepath.Join(rpmCacheRoot, archName+distribution.Name()),
		CustomRepos: customRepos,
	}
	// add RHSM fact to detect changes
	config.Options.Facts = &facts.ImageOptions{
		APIType: facts.TEST_APITYPE,
	}

	mg, err := manifestgen.New(reporeg, &manifestOpts)
	if err != nil {
		return err
	}
	// allow WARNINGS here
	if err := mg.Generate(config.Blueprint, distribution, imgType, arch, &config.Options); err != nil {
		return err
	}
	fmt.Print("DONE\n")

	manifestPath := filepath.Join(buildDir, "manifest.json")
	if err := save(mf.Bytes(), manifestPath); err != nil {
		return err
	}

	fmt.Printf("Building manifest: %s\n", manifestPath)

	jobOutput := filepath.Join(outputDir, buildName)
	_, err = osbuild.RunOSBuild(mf.Bytes(), osbuildStore, jobOutput, imgType.Exports(), checkpoints, nil, false, os.Stderr)
	if err != nil {
		return err
	}

	fmt.Printf("Jobs done. Results saved in\n%s\n", outputDir)
	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}
}
