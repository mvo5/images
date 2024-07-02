package main_test

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	main "github.com/osbuild/images/cmd/build"
)

func makeFakeBinary(t *testing.T, name, content string) {
	tmpdir := t.TempDir()
	t.Setenv("PATH", tmpdir+":"+os.Getenv("PATH"))

	err := os.WriteFile(filepath.Join(tmpdir, name), []byte(content), 0755)
	assert.NoError(t, err)
}

func TestSmokeNoArgs(t *testing.T) {
	fakeStderr := bytes.NewBuffer(nil)
	err := main.Run(os.Stdout, fakeStderr, nil)
	assert.ErrorContains(t, err, "need distro/image/config command line arguments")
	// usage info is printed
	assert.Contains(t, fakeStderr.String(), "Usage of ")
}

func TestSmokeRun(t *testing.T) {
	makeFakeBinary(t, "osbuild", "")
	makeFakeBinary(t, "osbuild-depsolve-dnf", `echo {}`)
	os.Chdir("../..")

	tmpdir := t.TempDir()
	fakeConfigPath := filepath.Join(tmpdir, "config.json")
	err := os.WriteFile(fakeConfigPath, []byte("{}"), 0644)
	assert.NoError(t, err)
	args := []string{
		"-config", fakeConfigPath,
		"-distro", "rhel-9.1",
		"-image", "ami",
	}
	err = main.Run(os.Stdout, os.Stderr, args)
	assert.NoError(t, err)
}
