package rpmmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPackageSpecGetEVRA(t *testing.T) {
	specs := []PackageSpec{
		{
			Name:    "tmux",
			Epoch:   0,
			Version: "3.3a",
			Release: "3.fc38",
			Arch:    "x86_64",
		},
		{
			Name:    "grub2",
			Epoch:   1,
			Version: "2.06",
			Release: "94.fc38",
			Arch:    "noarch",
		},
	}

	assert.Equal(t, "3.3a-3.fc38.x86_64", specs[0].GetEVRA())
	assert.Equal(t, "1:2.06-94.fc38.noarch", specs[1].GetEVRA())
}

func TestPackageSpecGetNEVRA(t *testing.T) {
	specs := []PackageSpec{
		{
			Name:    "tmux",
			Epoch:   0,
			Version: "3.3a",
			Release: "3.fc38",
			Arch:    "x86_64",
		},
		{
			Name:    "grub2",
			Epoch:   1,
			Version: "2.06",
			Release: "94.fc38",
			Arch:    "noarch",
		},
	}

	assert.Equal(t, "tmux-3.3a-3.fc38.x86_64", specs[0].GetNEVRA())
	assert.Equal(t, "grub2-1:2.06-94.fc38.noarch", specs[1].GetNEVRA())
}

func TestRepoConfigMarshalEmpty(t *testing.T) {
	repoCfg := &RepoConfig{}
	js, _ := json.Marshal(repoCfg)
	assert.Equal(t, string(js), `{}`)
}

func TestLoadRepositoryFromFileSmoke(t *testing.T) {
	repos, err := LoadRepositoriesFromFile("../../test/data/repositories/centos-10.json")
	assert.NoError(t, err)
	assert.True(t, len(repos) > 1)
}

func TestLoadRepositoryFromFileRedirect(t *testing.T) {
	reposOrig, err := LoadRepositoriesFromFile("../../test/data/repositories/centos-10.json")
	assert.NoError(t, err)

	// centos-stream-10.json just contains "alias": "centos-10"
	reposAlias, err := LoadRepositoriesFromFile("../../test/data/repositories/centos-stream-10.json")
	assert.NoError(t, err)
	assert.Equal(t, reposOrig, reposAlias)
}

func TestLoadRepositoryFromFileBad(t *testing.T) {
	badAliasPath := filepath.Join(t.TempDir(), "bad-alias.json")
	err := os.WriteFile(badAliasPath, []byte(`{"alias":"foo","unrelated":"stuff"}`), 0644)
	require.NoError(t, err)

	_, err = LoadRepositoriesFromFile(badAliasPath)
	assert.EqualError(t, err, "alias must be the only entry in a repo file")
}
