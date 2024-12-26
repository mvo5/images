package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToPtr(t *testing.T) {
	var valueInt int = 42
	gotInt := ToPtr(valueInt)
	assert.Equal(t, valueInt, *gotInt)

	var valueBool bool = true
	gotBool := ToPtr(valueBool)
	assert.Equal(t, valueBool, *gotBool)

	var valueUint64 uint64 = 1
	gotUint64 := ToPtr(valueUint64)
	assert.Equal(t, valueUint64, *gotUint64)

	var valueStr string = "the-greatest-test-value"
	gotStr := ToPtr(valueStr)
	assert.Equal(t, valueStr, *gotStr)
}

func TestUnrefOrDefault(t *testing.T) {
	for _, tc := range []any{
		7, "foo", true, false, 3.14, &struct{ e float32 }{2.718},
	} {
		assert.Equal(t, tc, UnrefOrDefault(&tc))
	}
}

func TestUnrefOr(t *testing.T) {
	var intP *int
	i := 2

	assert.Equal(t, 1, UnrefOr(intP, 1))
	assert.Equal(t, 2, UnrefOr(&i, 1))
}
