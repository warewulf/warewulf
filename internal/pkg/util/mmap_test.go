package util

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapToFile(t *testing.T) {
	f, err := os.CreateTemp(os.TempDir(), "test-*")
	assert.NoError(t, err)
	defer os.Remove(f.Name())

	mp := &MMap{}
	var buffer bytes.Buffer
	buffer.WriteString("hello world")

	data, err := mp.MapToFile(buffer.Bytes(), f)
	defer func() {
		_ = mp.Unmap(data)
	}()
	assert.NoError(t, err)

	assert.Equal(t, "hello world", string(data))

	fdata, err := io.ReadAll(f)
	assert.NoError(t, err)
	assert.Equal(t, "hello world", string(fdata))
}

func TestMapFromFile(t *testing.T) {
	f, err := os.CreateTemp(os.TempDir(), "test-*")
	assert.NoError(t, err)
	defer os.Remove(f.Name())

	mp := &MMap{}
	_, err = f.WriteString("hello world")
	assert.NoError(t, err)

	data, err := mp.MapFromFile(f)
	assert.NoError(t, err)
	defer func() {
		_ = mp.Unmap(data)
	}()

	assert.Equal(t, "hello world", string(data))
}

func TestMapBidirection(t *testing.T) {
	f, err := os.CreateTemp(os.TempDir(), "test-*")
	assert.NoError(t, err)
	defer os.Remove(f.Name())

	mp := &MMap{}
	var buffer bytes.Buffer
	buffer.WriteString("hello world")

	data, err := mp.MapToFile(buffer.Bytes(), f)
	defer func() {
		_ = mp.Unmap(data)
	}()
	assert.NoError(t, err)
	assert.Equal(t, "hello world", string(data))

	// update file
	_, err = f.Seek(0, 2)
	assert.NoError(t, err)

	_, err = f.WriteString("\n another world")
	assert.NoError(t, err)

	data2, err := mp.MapFromFile(f)
	assert.NoError(t, err)
	defer func() {
		_ = mp.Unmap(data)
	}()

	assert.Equal(t, "hello world\n another world", string(data2))

	// update buffer again
	buffer.Reset()
	_, err = buffer.WriteString("new world")
	assert.NoError(t, err)
	data3, err := mp.MapToFile(buffer.Bytes(), f)
	assert.NoError(t, err)
	defer func() {
		_ = mp.Unmap(data)
	}()

	assert.Equal(t, "new world", string(data3))

	data4, err := io.ReadAll(f)
	assert.NoError(t, err)

	assert.Equal(t, "new world", string(data4))
}

func TestFailureBecauseUnwritableFile(t *testing.T) {
	f, err := os.CreateTemp(os.TempDir(), "test-*")
	assert.NoError(t, err)
	defer os.Remove(f.Name())

	err = os.Chmod(f.Name(), 0440)
	assert.NoError(t, err)

	mp := &MMap{}
	d, err := mp.MapToFile([]byte("test"), f)
	assert.Error(t, err)
	assert.Nil(t, d)
	assert.Contains(t, err.Error(), "file is not writable")
}

func TestFailureBecauseUnreadableFile(t *testing.T) {
	f, err := os.CreateTemp(os.TempDir(), "test-*")
	assert.NoError(t, err)
	defer os.Remove(f.Name())

	err = os.Chmod(f.Name(), 0300)
	assert.NoError(t, err)

	mp := &MMap{}
	d, err := mp.MapFromFile(f)
	assert.Error(t, err)
	assert.Nil(t, d)
	assert.Contains(t, err.Error(), "file is not readable")
}
