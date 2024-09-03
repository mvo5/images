package common_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/osbuild/images/internal/common"
)

func TestDecodeErrorHandling(t *testing.T) {
	bad := bytes.NewBufferString("no json")
	var v interface{}
	err := common.ReadDecodeJSON(bad, &v)
	assert.ErrorContains(t, err, `cannot unmarshal json "no json": invalid character `)
}
