package common

import (
	"encoding/json"
	"fmt"
	"io"
)

// ReadDecodeJSON reads the entire reader and decodes the json, on error
// the full json is part of the error message.
func ReadDecodeJSON(r io.Reader, v any) error {
	b, err := io.ReadAll(r)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(b, &v); err != nil {
		return fmt.Errorf("cannot unmarshal json %q: %w", b, err)
	}
	return nil
}
