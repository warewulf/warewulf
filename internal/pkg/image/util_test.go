package image

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_ValidName(t *testing.T) {
	for name, valid := range map[string]bool{
		"rockylinux-9": true,
		"rocky:9.5":    true,
		"a.b":          true,
		"":             false,
		".":            false,
		"..":           false,
		"../x":         false,
		"with space":   false,
	} {
		assert.Equal(t, valid, ValidName(name), name)
	}
}
