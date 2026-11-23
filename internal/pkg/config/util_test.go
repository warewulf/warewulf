package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetInterfaceForIP(t *testing.T) {
	assert.Equal(t, "lo", GetInterfaceForIP("::1"))
	assert.Equal(t, "lo", GetInterfaceForIP("127.0.0.1"))
	assert.Empty(t, GetInterfaceForIP("2001:db8::1"))
	assert.Empty(t, GetInterfaceForIP("not-an-address"))
	assert.Empty(t, GetInterfaceForIP(""))
}
