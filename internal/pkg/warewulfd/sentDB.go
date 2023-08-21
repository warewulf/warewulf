package warewulfd

import (
	"crypto/sha256"
	"fmt"
	"io"
	"path"
	"sync"

	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

/*
store the sent files name and its checksum
*/
type SentFiles struct {
	Files     []File `json:"files:"`
	sha256sum [32]byte
	Sha256hex string `json:"sha256"`
	wwinit    bool
}

type File struct {
	FileName  string `json:"file name"`
	sha256sum [32]byte
	Sha256hex string `json:"sha256"`
}

// database for the checksum of sent files
var sentDB map[string]*SentFiles

// mutex for locking the map
var mu sync.Mutex

func init() {
	sentDB = map[string]*SentFiles{}
}

/*
Adds the image with the name to the database
*/
func DBAddImage(node string, fileName string, content io.ReadSeeker) {
	wwlog.Debug("adding file %s for node %s to sentDB", node, fileName)
	hasher := sha256.New()
	sent := File{
		FileName: path.Base(fileName),
	}
	if _, err := io.Copy(hasher, content); err != nil {
		wwlog.SecWarn("couldn't create hash of %s for %s", fileName, node)
		return
	}
	copy(sent.sha256sum[:], hasher.Sum(nil))
	sent.Sha256hex = fmt.Sprintf("%x", sent.sha256sum)
	mu.Lock()
	if _, ok := sentDB[node]; !ok {
		sentDB[node] = new(SentFiles)
	}
	sentDB[node].Files = append(sentDB[node].Files, sent)
	for i := 0; i < sha256.Size; i++ {
		sentDB[node].sha256sum[i] = 1
	}
	for _, sntFile := range sentDB[node].Files {
		sentDB[node].sha256sum = sha256.Sum256(append(sentDB[node].sha256sum[:], sntFile.sha256sum[:]...))
	}
	sentDB[node].Sha256hex = fmt.Sprintf("%x", sentDB[node].sha256sum)
	mu.Unlock()
	hasher.Reset()
}

/*
Get the final sum of all the hashed files
*/
func DBGetSum(node string) (ret [sha256.Size]byte) {
	mu.Lock()
	defer mu.Unlock()
	if sentNode, ok := sentDB[node]; ok {
		ret = sentNode.sha256sum
		return
	}
	ret = [sha256.Size]byte{0}
	return
}

/*
Reset the database for a single node
*/
func DBReset(node string) {
	mu.Lock()
	if _, ok := sentDB[node]; !ok {
		sentDB[node] = new(SentFiles)
	}
	sentDB[node] = new(SentFiles)
	mu.Unlock()
}

/*
Reset the database
*/
func DBResetAll() {
	sentDB = make(map[string]*SentFiles)
}

/*
Get the size of the DB
*/
func DBSize(node string) int {
	mu.Lock()
	if _, ok := sentDB[node]; !ok {
		sentDB[node] = new(SentFiles)
	}
	size := len(sentDB[node].Files)
	mu.Unlock()
	return size
}

/*
Check if wwinit was sent
*/
func DBGetWWinit(node string) bool {
	mu.Lock()
	if _, ok := sentDB[node]; !ok {
		sentDB[node] = new(SentFiles)
	}
	ret := sentDB[node].wwinit
	mu.Unlock()
	return ret
}

/*
Mark that wwinit was sent
*/

func DBWWinitSent(node string) {
	mu.Lock()
	if _, ok := sentDB[node]; !ok {
		sentDB[node] = new(SentFiles)
	}
	sentDB[node].wwinit = true
	mu.Unlock()
}
