// buildconfig is used by the internal tools gen-manifests, build only currently
package buildconfig

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/osbuild/images/pkg/blueprint"
	"github.com/osbuild/images/pkg/distro"
)

type BuildConfig struct {
	Name      string               `json:"name"`
	Blueprint *blueprint.Blueprint `json:"blueprint,omitempty"`
	Options   distro.ImageOptions  `json:"options"`
	Depends   interface{}          `json:"depends,omitempty"` // ignored
}

func tryDecode(dec *json.Decoder, path string, data any) error {
	if err := dec.Decode(&data); err != nil {
		return err
	}
	if dec.More() {
		return fmt.Errorf("multiple configuration objects or extra data found in %q", path)
	}
	return nil
}

func New(path string) (*BuildConfig, error) {
	fp, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer fp.Close()

	dec := json.NewDecoder(fp)
	dec.DisallowUnknownFields()
	var conf BuildConfig
	if err := tryDecode(dec, path, &conf); err != nil {
		fmt.Fprintf(os.Stderr, "WARNING: strict decoding failed, trying non-strict: %s", err)
	}
	dec = json.NewDecoder(fp)
	if err := tryDecode(dec, path, &conf); err != nil {
		return nil, err
	}

	return &conf, nil
}
