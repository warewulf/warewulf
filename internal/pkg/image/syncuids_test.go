package image

import (
	"bytes"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func writeTempFile(t *testing.T, input string) string {
	tempFile, createTempError := os.CreateTemp("", "syncuids-*")
	assert.NoError(t, createTempError)
	_, writeError := tempFile.Write([]byte(input))
	assert.NoError(t, writeError)
	assert.NoError(t, tempFile.Sync())
	return tempFile.Name()
}

func makeSyncDB(t *testing.T, hostInput string, imageInput string) syncDB {
	hostFileName := writeTempFile(t, hostInput)
	defer os.Remove(hostFileName)
	imageFileName := writeTempFile(t, imageInput)
	defer os.Remove(imageFileName)
	db := make(syncDB)
	var err error
	err = db.readFromHost(hostFileName)
	assert.NoError(t, err)
	err = db.readFromimage(imageFileName)
	assert.NoError(t, err)
	return db
}

func Test_readFromHost_single(t *testing.T) {
	hostFileName := writeTempFile(t, `testuser:x:1001:1001::/home/testuser:/bin/bash`)
	defer os.Remove(hostFileName)

	db := make(syncDB)
	err := db.readFromHost(hostFileName)
	assert.NoError(t, err)

	assert.Len(t, db, 1)
	assert.Equal(t, 1001, db["testuser"].HostID)
	assert.Equal(t, -1, db["testuser"].imageID)
}

func Test_readFromHost_multiple(t *testing.T) {
	hostFileName := writeTempFile(t, `
testuser1:x:1001:1001::/home/testuser:/bin/bash
testuser2:x:1002:1002::/home/testuser:/bin/bash
`)
	defer os.Remove(hostFileName)

	db := make(syncDB)
	err := db.readFromHost(hostFileName)
	assert.NoError(t, err)

	assert.Len(t, db, 2)
	assert.Equal(t, 1001, db["testuser1"].HostID)
	assert.Equal(t, -1, db["testuser1"].imageID)
	assert.Equal(t, 1002, db["testuser2"].HostID)
	assert.Equal(t, -1, db["testuser2"].imageID)
}

func Test_readFromimage_single(t *testing.T) {
	imageFileName := writeTempFile(t, `testuser:x:1001:1001::/home/testuser:/bin/bash`)
	defer os.Remove(imageFileName)

	db := make(syncDB)
	err := db.readFromimage(imageFileName)
	assert.NoError(t, err)

	assert.Len(t, db, 1)
	assert.Equal(t, 1001, db["testuser"].imageID)
	assert.Equal(t, -1, db["testuser"].HostID)
}

func Test_readFromimage_multiple(t *testing.T) {
	imageFileName := writeTempFile(t, `
testuser1:x:1001:1001::/home/testuser:/bin/bash
testuser2:x:1002:1002::/home/testuser:/bin/bash
`)
	defer os.Remove(imageFileName)

	db := make(syncDB)
	err := db.readFromimage(imageFileName)
	assert.NoError(t, err)
	assert.Len(t, db, 2)
	assert.Equal(t, 1001, db["testuser1"].imageID)
	assert.Equal(t, -1, db["testuser1"].HostID)
	assert.Equal(t, 1002, db["testuser2"].imageID)
	assert.Equal(t, -1, db["testuser2"].HostID)
}

func Test_readFromBoth_multiple(t *testing.T) {
	imageFileName := writeTempFile(t, `
testuser1:x:1001:1001::/home/testuser:/bin/bash
testuser2:x:1002:1002::/home/testuser:/bin/bash
`)
	defer os.Remove(imageFileName)

	hostFileName := writeTempFile(t, `
testuser1:x:2001:2001::/home/testuser:/bin/bash
testuser3:x:2003:2003::/home/testuser:/bin/bash
`)
	defer os.Remove(hostFileName)

	db := make(syncDB)
	var err error
	err = db.readFromimage(imageFileName)
	assert.NoError(t, err)
	err = db.readFromHost(hostFileName)
	assert.NoError(t, err)
	assert.Len(t, db, 3)
	assert.Equal(t, 1001, db["testuser1"].imageID)
	assert.Equal(t, 2001, db["testuser1"].HostID)
	assert.Equal(t, 1002, db["testuser2"].imageID)
	assert.Equal(t, -1, db["testuser2"].HostID)
	assert.Equal(t, -1, db["testuser3"].imageID)
	assert.Equal(t, 2003, db["testuser3"].HostID)
}

func Test_checkConflicts_empty(t *testing.T) {
	db := makeSyncDB(t, "", "")
	assert.NoError(t, db.checkConflicts())
}

func Test_checkConflicts_single(t *testing.T) {
	db := makeSyncDB(t, "", "testuser:x:1001:1001::/home/testuser:/bin/bash")
	assert.NoError(t, db.checkConflicts())
}

func Test_checkConflicts_match(t *testing.T) {
	db := makeSyncDB(t,
		"testuser:x:1001:1001::/home/testuser:/bin/bash",
		"testuser:x:1001:1001::/home/testuser:/bin/bash")
	assert.NoError(t, db.checkConflicts())
}

func Test_checkConflicts_conflict(t *testing.T) {
	db := makeSyncDB(t,
		"testuser2:x:1001:1001::/home/testuser:/bin/bash",
		"testuser1:x:1001:1001::/home/testuser:/bin/bash")
	assert.Error(t, db.checkConflicts())
}

func Test_getOnlyimageLines(t *testing.T) {
	imageFileName := writeTempFile(t, `
testuser1:x:1001:1001::/home/testuser:/bin/bash
testuser2:x:1002:1002::/home/testuser:/bin/bash
`)
	defer os.Remove(imageFileName)

	hostFileName := writeTempFile(t, `
testuser1:x:2001:2001::/home/testuser:/bin/bash
testuser3:x:2003:2003::/home/testuser:/bin/bash
`)
	defer os.Remove(hostFileName)

	db := make(syncDB)
	var err error
	err = db.readFromimage(imageFileName)
	assert.NoError(t, err)
	err = db.readFromHost(hostFileName)
	assert.NoError(t, err)

	lines, err := db.getOnlyimageLines(imageFileName)
	assert.NoError(t, err)

	assert.Len(t, lines, 1)
	assert.Equal(t, lines[0], "testuser2:x:1002:1002::/home/testuser:/bin/bash")
}

func Test_needsSync_empty(t *testing.T) {
	db := makeSyncDB(t, "", "")
	assert.False(t, db.needsSync())
}

func Test_needsSync_imageOnly(t *testing.T) {
	db := makeSyncDB(t, "", `
testuser1:x:1001:1001::/home/testuser:/bin/bash
testuser2:x:1002:1002::/home/testuser:/bin/bash`)

	assert.False(t, db.needsSync())
}

func Test_needsSync_hostOnly(t *testing.T) {
	db := makeSyncDB(t, `
testuser1:x:1001:1001::/home/testuser:/bin/bash
testuser2:x:1002:1002::/home/testuser:/bin/bash`, "")

	assert.True(t, db.needsSync())
}

func Test_needsSync_match(t *testing.T) {
	db := makeSyncDB(t,
		"testuser:x:1001:1001::/home/testuser:/bin/bash",
		"testuser:x:1001:1001::/home/testuser:/bin/bash")

	assert.False(t, db.needsSync())
}

func Test_needsSync_differ(t *testing.T) {
	db := makeSyncDB(t,
		`
testuser1:x:2001:2001::/home/testuser:/bin/bash
testuser3:x:2003:2003::/home/testuser:/bin/bash`,
		`
testuser1:x:1001:1001::/home/testuser:/bin/bash
testuser2:x:1002:1002::/home/testuser:/bin/bash`)

	assert.True(t, db.needsSync())
}

func Test_onlyHost(t *testing.T) {
	db := makeSyncDB(t, "testuser1:x:2001:2001::/home/testuser:/bin/bash", "")
	entry := db["testuser1"]
	assert.True(t, entry.inHost())
	assert.False(t, entry.inimage())
	assert.True(t, entry.onlyHost())
	assert.False(t, entry.onlyimage())
	assert.False(t, entry.match())
	assert.False(t, entry.differ())
}

func Test_onlyimage(t *testing.T) {
	db := makeSyncDB(t, "", "testuser1:x:2001:2001::/home/testuser:/bin/bash")
	entry := db["testuser1"]
	assert.False(t, entry.inHost())
	assert.True(t, entry.inimage())
	assert.False(t, entry.onlyHost())
	assert.True(t, entry.onlyimage())
	assert.False(t, entry.match())
	assert.False(t, entry.differ())
}

func Test_match(t *testing.T) {
	db := makeSyncDB(t,
		"testuser1:x:2001:2001::/home/testuser:/bin/bash",
		"testuser1:x:2001:2001::/home/testuser:/bin/bash")
	entry := db["testuser1"]
	assert.True(t, entry.inHost())
	assert.True(t, entry.inimage())
	assert.False(t, entry.onlyHost())
	assert.False(t, entry.onlyimage())
	assert.True(t, entry.match())
	assert.False(t, entry.differ())
}

func Test_differ(t *testing.T) {
	db := makeSyncDB(t,
		"testuser1:x:1001:1001::/home/testuser:/bin/bash",
		"testuser1:x:2001:2001::/home/testuser:/bin/bash")
	entry := db["testuser1"]
	assert.True(t, entry.inHost())
	assert.True(t, entry.inimage())
	assert.False(t, entry.onlyHost())
	assert.False(t, entry.onlyimage())
	assert.False(t, entry.match())
	assert.True(t, entry.differ())
}

func Test_malformed_passwd(t *testing.T) {
	hostInput := `"testuser1:x:1001:1001::/home/testuser:/bin/bash"
	asdf`
	hostFileName := writeTempFile(t, hostInput)
	defer os.Remove(hostFileName)
	db := make(syncDB)
	err := db.readFromHost(hostFileName)
	assert.NoError(t, err)
}

func Test_network_passwd(t *testing.T) {
	buf := new(bytes.Buffer)
	wwlog.SetLogWriter(buf)
	hostInput := `testuser1:x:1001:1001::/home/testuser:/bin/bash
+::::::
-::::::`
	hostFileName := writeTempFile(t, hostInput)
	defer os.Remove(hostFileName)
	db := make(syncDB)
	err := db.readFromHost(hostFileName)
	assert.NotContains(t, buf.String(), "parse error")
	assert.NoError(t, err)
}
