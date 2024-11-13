package distro

import (
	"fmt"
	"path/filepath"
)

func ValidateImageTypeFilename(s string) error {
	if filepath.Base(s) != s {
		return fmt.Errorf("cannot use %q: only relative filenames supported", s)
	}
	if filepath.Clean(s) != s {
		return fmt.Errorf("cannot use %q: only clean filenames supported", s)
	}

	return nil
}
