package overlaydiff

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultExcludes_ContainsSystemAndUserNoisePaths(t *testing.T) {
	assert.Contains(t, DefaultExcludes, "/boot")
	assert.Contains(t, DefaultExcludes, "/home")
	assert.Contains(t, DefaultExcludes, "/root")
	assert.NotContains(t, DefaultExcludes, "/var/lib")
}
