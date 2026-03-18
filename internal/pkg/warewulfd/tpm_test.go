package warewulfd

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/tpm"
)

func TestTPMLogStore(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()

	// Initialize the configuration so paths are available
	_ = env.Configure()
	conf := warewulfconf.Get()

	nodeId := "testnode"
	tpmPath := filepath.Join(conf.Paths.OverlayProvisiondir(), nodeId, "tpm.json")

	// Test NewTPMLogStore creation
	store, err := NewTPMLogStore(nodeId)
	assert.NoError(t, err)
	assert.NotNil(t, store)

	// Verify tpm.json is created with initial empty quote and modified time
	assert.FileExists(t, tpmPath)
	data, err := os.ReadFile(tpmPath)
	assert.NoError(t, err)

	var initialQuote tpm.Quote
	err = json.Unmarshal(data, &initialQuote)
	assert.NoError(t, err)
	assert.Empty(t, initialQuote.EKCert)
	assert.NotZero(t, initialQuote.Modified)

	// Test Save (Quote)
	newQuote := tpm.Quote{
		ID:     nodeId,
		EKCert: "dummy_cert",
	}
	err = store.Save(newQuote)
	assert.NoError(t, err)

	data, err = os.ReadFile(tpmPath)
	assert.NoError(t, err)
	var savedQuote tpm.Quote
	err = json.Unmarshal(data, &savedQuote)
	assert.NoError(t, err)
	assert.Equal(t, "dummy_cert", savedQuote.EKCert)
	assert.Equal(t, nodeId, savedQuote.ID)
	assert.True(t, savedQuote.Modified.After(initialQuote.Modified) || savedQuote.Modified.Equal(initialQuote.Modified))

	// Test SaveChallenge
	challenge := tpm.Challenge{
		Secret: []byte("my_secret"),
		ID:     nodeId,
	}
	err = store.SaveChallenge(challenge)
	assert.NoError(t, err)

	data, err = os.ReadFile(tpmPath)
	assert.NoError(t, err)
	err = json.Unmarshal(data, &savedQuote)
	assert.NoError(t, err)
	assert.NotNil(t, savedQuote.Challenge)
	assert.Equal(t, []byte("my_secret"), savedQuote.Challenge.Secret)

	// Test GetSecret
	secret := store.GetSecret()
	assert.Equal(t, "my_secret", secret)

	// Test Update (Adding log entries)
	filename1 := "image1.img"
	checksum1 := fmt.Sprintf("%x", sha256.Sum256([]byte("data1")))
	err = store.Update(filename1, checksum1)
	assert.NoError(t, err)

	// Update with file reading (empty checksum)
	testFile := env.GetPath("test.img")
	err = os.WriteFile(testFile, []byte("test_data"), 0644)
	assert.NoError(t, err)
	err = store.Update(testFile, "")
	assert.NoError(t, err)

	data, err = os.ReadFile(tpmPath)
	assert.NoError(t, err)
	err = json.Unmarshal(data, &savedQuote)
	assert.NoError(t, err)

	assert.Len(t, savedQuote.SentLog, 2)
	assert.Equal(t, filename1, savedQuote.SentLog[0].Filename)
	assert.Equal(t, checksum1, savedQuote.SentLog[0].Checksum)
	assert.Equal(t, testFile, savedQuote.SentLog[1].Filename)
	expectedChecksum := fmt.Sprintf("%x", sha256.Sum256([]byte("test_data")))
	assert.Equal(t, expectedChecksum, savedQuote.SentLog[1].Checksum)

	// Test Update with duplicate entry (should not add duplicate)
	err = store.Update(filename1, checksum1)
	assert.NoError(t, err)
	data, err = os.ReadFile(tpmPath)
	assert.NoError(t, err)
	err = json.Unmarshal(data, &savedQuote)
	assert.NoError(t, err)
	assert.Len(t, savedQuote.SentLog, 2)

	// Test ClearLogs
	err = store.ClearLogs()
	assert.NoError(t, err)

	data, err = os.ReadFile(tpmPath)
	assert.NoError(t, err)
	savedQuote = tpm.Quote{}
	err = json.Unmarshal(data, &savedQuote)
	assert.NoError(t, err)
	assert.Empty(t, savedQuote.SentLog)
}
