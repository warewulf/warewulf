package syncuser

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func TestConflictMessages(t *testing.T) {
	for _, tc := range []struct{ file, suffix, id string }{
		{"passwd", ":x:1234:100::/:/bin/sh", "UID"},
		{"group", ":x:1234:", "GID"},
	} {
		t.Run(tc.file, func(t *testing.T) {
			env := testenv.New(t)
			defer env.RemoveAll()
			imageEtc := filepath.Join(testenv.WWChrootdir, "sample", "rootfs", "etc")
			for _, file := range []string{"passwd", "group"} {
				var host, image string
				if file == tc.file {
					host = "host-name" + tc.suffix + "\n"
					image = "image-name" + tc.suffix + "\n"
				}
				env.WriteFile(filepath.Join(testenv.Sysconfdir, file), host)
				env.WriteFile(filepath.Join(imageEtc, file), image)
			}
			var warnings bytes.Buffer
			wwlog.SetLogWriterErr(&warnings)
			defer wwlog.SetLogWriterErr(os.Stderr)
			cmd := GetCommand()
			cmd.SetArgs([]string{"--write=false", "sample"})
			assert.EqualError(t, cmd.Execute(), "error in synchronize: "+tc.id+"(1234) collision: host(host-name) image(image-name)")
			assert.Contains(t, warnings.String(), tc.id+" 1234")
		})
	}
}
