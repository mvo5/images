package fsnode

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"

	"github.com/osbuild/images/internal/common"
)

type directoryJSON struct {
	Base             baseFsNode
	EnsureParentDirs bool `json:"ensure_parent_dirs" yaml:"ensure_parent_dirs"`
}

type Directory struct {
	directoryJSON
}

func (d *Directory) IsDir() bool {
	return true
}

func (d *Directory) EnsureParentDirs() bool {
	if d == nil {
		return false
	}
	return d.directoryJSON.EnsureParentDirs
}

func (d *Directory) UnmarshalJSON(data []byte) error {
	var dv directoryJSON
	dec := json.NewDecoder(bytes.NewBuffer(data))
	dec.UseNumber()
	dec.DisallowUnknownFields()
	if err := dec.Decode(&dv); err != nil {
		return err
	}
	fmt.Printf("%+v %q\n", dv, data)
	d.directoryJSON = dv

	return d.validate()

}

func (d *Directory) UnmarshalYAML(unmarshal func(any) error) error {
	return common.UnmarshalYAMLviaJSON(d, unmarshal)
}

// NewDirectory creates a new directory with the given path, mode, user and group.
// user and group can be either a string (user name/group name), an int64 (UID/GID) or nil.
func NewDirectory(path string, mode *os.FileMode, user interface{}, group interface{}, ensureParentDirs bool) (*Directory, error) {
	baseNode, err := newBaseFsNode(path, mode, user, group)

	if err != nil {
		return nil, err
	}

	return &Directory{
		directoryJSON: directoryJSON{
			baseFsNode:       *baseNode,
			EnsureParentDirs: ensureParentDirs,
		},
	}, nil
}
