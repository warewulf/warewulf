package tpm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetCommand(t *testing.T) {
	cmd := GetCommand()
	assert.NotNil(t, cmd)
	assert.Equal(t, "tpm COMMAND [OPTIONS]", cmd.Use)

	subCommands := cmd.Commands()
	assert.Len(t, subCommands, 3)

	found := make(map[string]bool)
	for _, sub := range subCommands {
		found[sub.Name()] = true
	}

	assert.True(t, found["list"])
	assert.True(t, found["verify"])
	assert.True(t, found["check"])
}
