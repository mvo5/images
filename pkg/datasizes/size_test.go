package datasizes_test

import (
	"encoding/json"
	"testing"

	"github.com/BurntSushi/toml"
	"github.com/stretchr/testify/assert"

	"github.com/osbuild/images/pkg/datasizes"
)

func TestSizeUnmarshalTOMLUnhappy(t *testing.T) {
	cases := []struct {
		name  string
		input string
		err   string
	}{
		{
			name:  "wrong datatype",
			input: `size = true`,
			err:   `toml: line 1 (last key "size"): TOML unmarshal: error decoding size: failed to convert value "true" to number`,
		},
		{
			name:  "wrong unit",
			input: `size = "20 KG"`,
			err:   `toml: line 1 (last key "size"): TOML unmarshal: error decoding size: unknown data size units in string: 20 KG`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var v struct {
				Size datasizes.Size `toml:"size"`
			}
			err := toml.Unmarshal([]byte(tc.input), &v)
			assert.EqualError(t, err, tc.err, tc.input)
		})
	}
}

func TestSizeUnmarshalJSONUnhappy(t *testing.T) {
	cases := []struct {
		name  string
		input string
		err   string
	}{
		{
			name:  "misize nor string nor int",
			input: `{"size": true}`,
			err:   `JSON unmarshal: error decoding size: failed to convert value "true" to number`,
		},
		{
			name:  "misize not parseable",
			input: `{"size": "20 KG"}`,
			err:   `JSON unmarshal: error decoding size: unknown data size units in string: 20 KG`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			var v struct {
				Size datasizes.Size `json:"size"`
			}
			err := json.Unmarshal([]byte(tc.input), &v)
			assert.EqualError(t, err, tc.err, tc.input)
		})
	}
}
