package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSubStrings(t *testing.T) {
	assert.Equal(t, SubString("", 0), "", "")
	assert.Equal(t, SubString("", 3), "", "")
	assert.Equal(t, SubString("", -1), "", "")
	assert.Equal(t, SubString("b", 0), "", "")
	assert.Equal(t, SubString("b", -1), "b", "")
	assert.Equal(t, SubString("b", 1), "b", "")
	assert.Equal(t, SubString("b", 2), "b", "")
	assert.Equal(t, SubString("batman", -1), "batman", "")
	assert.Equal(t, SubString("batman", 0), "", "")
	assert.Equal(t, SubString("batman", 1), "b", "")
	assert.Equal(t, SubString("batman", 2), "ba", "")
	assert.Equal(t, SubString("batman", 5), "batma", "")
	assert.Equal(t, SubString("batman", 6), "batman", "")
	assert.Equal(t, SubString("batman", 15), "batman", "")
}
