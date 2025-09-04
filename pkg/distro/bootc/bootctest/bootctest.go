package bootctest

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"github.com/osbuild/images/internal/randutil"
)

func makeOsRelease(buildDir string) error {
	osRelease := `
NAME="fake-name"
ID="fake-id"
VERSION_ID="1"
`

	osReleasePath := filepath.Join(buildDir, "etc/os-release")
	if err := os.MkdirAll(filepath.Dir(osReleasePath), 0755); err != nil {
		return err
	}
	//nolint:gosec
	return os.WriteFile(osReleasePath, []byte(osRelease), 0644)
}

func makeUsrBinInstall(buildDir string) error {
	installToml := `
[install]
filesystem = [
    { mountpoint = "/", type = "xfs", size = "10 GiB" },
    { mountpoint = "/boot", type = "ext4", size = "1 GiB" },
]
`

	installTomlPath := filepath.Join(buildDir, "usr/lib/bootc/install/99-fedora-install.toml")
	if err := os.MkdirAll(filepath.Dir(installTomlPath), 0755); err != nil {
		return err
	}
	//nolint:gosec
	return os.WriteFile(installTomlPath, []byte(installToml), 0644)
}

func makeFakeBinaries(buildDir string) error {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		return fmt.Errorf("cannot find caller")
	}
	currentDir := filepath.Dir(currentFile)

	fakeBootcPath := filepath.Join(buildDir, "usr/bin/bootc")
	if err := os.MkdirAll(filepath.Dir(fakeBootcPath), 0755); err != nil {
		return err
	}
	cmd := exec.Command(
		"go", "build",
		"-o", fakeBootcPath,
		filepath.Join(currentDir, "./exe"),
	)
	cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("error running go: %w\n%s", err, output)
	}

	fakeSleepPath := filepath.Join(buildDir, "usr/bin/sleep")
	return os.Symlink("bootc", fakeSleepPath)
}

func makeContainerfile(buildDir string) error {
	var fakeBootcCnt = `
FROM scratch
COPY etc /etc
COPY usr/bin /usr/bin
COPY usr/lib/bootc/install /usr/lib/bootc/install 
`

	cntFilePath := filepath.Join(buildDir, "Containerfile")
	//nolint:gosec
	return os.WriteFile(cntFilePath, []byte(fakeBootcCnt), 0644)
}

func makeFakeContainerImage(buildDir, purpose string) (cntTag string, cleanup func() error, err error) {
	imgTag := fmt.Sprintf("image-builder-test-%s-%s", purpose, randutil.String(10, randutil.AsciiLower))
	output, err := exec.Command(
		"podman", "build",
		"-f", filepath.Join(buildDir, "Containerfile"),
		"-t", imgTag,
	).CombinedOutput()
	if err != nil {
		return "", nil, fmt.Errorf("error running podman: %w\n%s", err, output)
	}
	// add cleanup
	cleanup = func() error {
		output, err := exec.Command("podman", "image", "rm", imgTag).CombinedOutput()
		if err != nil {
			return fmt.Errorf("cannot run podman cleanup: %w\n%s", err, output)
		}
		return nil
	}

	return fmt.Sprintf("localhost/%s", imgTag), cleanup, nil
}

func NewFakeContainer(purpose string) (imgTag string, cleanup func() error, err error) {
	buildDir, err := os.MkdirTemp("", "fake-bootc-cnt")
	if err != nil {
		return "", nil, err
	}

	// XXX: allow adding test specific content
	if err := makeContainerfile(buildDir); err != nil {
		return "", nil, err
	}
	if err := makeFakeBinaries(buildDir); err != nil {
		return "", nil, err
	}
	// XXX: make os-release content configurable
	if err := makeOsRelease(buildDir); err != nil {
		return "", nil, err
	}
	if err := makeUsrBinInstall(buildDir); err != nil {
		return "", nil, err
	}

	cntTag, cleanup, err := makeFakeContainerImage(buildDir, purpose)
	if err != nil {
		return "", nil, err
	}
	return cntTag, cleanup, err
}
