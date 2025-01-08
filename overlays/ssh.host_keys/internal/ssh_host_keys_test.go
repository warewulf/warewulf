package ssh_host_keys

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/warewulf/warewulf/internal/app/wwctl/overlay/show"
	"github.com/warewulf/warewulf/internal/pkg/testenv"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func Test_ssh_host_keysOverlay(t *testing.T) {
	env := testenv.New(t)
	defer env.RemoveAll()
	env.ImportFile("etc/warewulf/nodes.conf", "nodes.conf")
	env.ImportFile("var/lib/warewulf/overlays/ssh.host_keys/rootfs/etc/ssh/ssh_host_dsa_key.pub.ww", "../rootfs/etc/ssh/ssh_host_dsa_key.pub.ww")
	env.ImportFile("var/lib/warewulf/overlays/ssh.host_keys/rootfs/etc/ssh/ssh_host_dsa_key.ww", "../rootfs/etc/ssh/ssh_host_dsa_key.ww")
	env.ImportFile("var/lib/warewulf/overlays/ssh.host_keys/rootfs/etc/ssh/ssh_host_ecdsa_key.pub.ww", "../rootfs/etc/ssh/ssh_host_ecdsa_key.pub.ww")
	env.ImportFile("var/lib/warewulf/overlays/ssh.host_keys/rootfs/etc/ssh/ssh_host_ecdsa_key.ww", "../rootfs/etc/ssh/ssh_host_ecdsa_key.ww")
	env.ImportFile("var/lib/warewulf/overlays/ssh.host_keys/rootfs/etc/ssh/ssh_host_ed25519_key.pub.ww", "../rootfs/etc/ssh/ssh_host_ed25519_key.pub.ww")
	env.ImportFile("var/lib/warewulf/overlays/ssh.host_keys/rootfs/etc/ssh/ssh_host_ed25519_key.ww", "../rootfs/etc/ssh/ssh_host_ed25519_key.ww")
	env.ImportFile("var/lib/warewulf/overlays/ssh.host_keys/rootfs/etc/ssh/ssh_host_rsa_key.pub.ww", "../rootfs/etc/ssh/ssh_host_rsa_key.pub.ww")
	env.ImportFile("var/lib/warewulf/overlays/ssh.host_keys/rootfs/etc/ssh/ssh_host_rsa_key.ww", "../rootfs/etc/ssh/ssh_host_rsa_key.ww")
	env.WriteFile("etc/warewulf/keys/ssh_host_dsa_key.pub", `dsa pubkey sentinel`)
	env.WriteFile("etc/warewulf/keys/ssh_host_dsa_key", `dsa key sentinel`)
	env.WriteFile("etc/warewulf/keys/ssh_host_ecdsa_key.pub", `ecdsa pubkey sentinel`)
	env.WriteFile("etc/warewulf/keys/ssh_host_ecdsa_key", `ecdsa key sentinel`)
	env.WriteFile("etc/warewulf/keys/ssh_host_ed25519_key.pub", `ed25519 pubkey sentinel`)
	env.WriteFile("etc/warewulf/keys/ssh_host_ed25519_key", `ed25519 key sentinel`)
	env.WriteFile("etc/warewulf/keys/ssh_host_rsa_key.pub", `rsa pubkey sentinel`)
	env.WriteFile("etc/warewulf/keys/ssh_host_rsa_key", `rsa key sentinel`)

	tests := []struct {
		name string
		args []string
		log  string
	}{
		{
			name: "ssh.host_keys:dsa pub",
			args: []string{"--render", "node1", "ssh.host_keys", "etc/ssh/ssh_host_dsa_key.pub.ww"},
			log:  ssh_host_dsa_key_pub,
		},
		{
			name: "ssh.host_keys:dsa",
			args: []string{"--render", "node1", "ssh.host_keys", "etc/ssh/ssh_host_dsa_key.ww"},
			log:  ssh_host_dsa_key,
		},
		{
			name: "ssh.host_keys:ecdsa pub",
			args: []string{"--render", "node1", "ssh.host_keys", "etc/ssh/ssh_host_ecdsa_key.pub.ww"},
			log:  ssh_host_ecdsa_key_pub,
		},
		{
			name: "ssh.host_keys:ecdsa",
			args: []string{"--render", "node1", "ssh.host_keys", "etc/ssh/ssh_host_ecdsa_key.ww"},
			log:  ssh_host_ecdsa_key,
		},
		{
			name: "ssh.host_keys:rsa pub",
			args: []string{"--render", "node1", "ssh.host_keys", "etc/ssh/ssh_host_rsa_key.pub.ww"},
			log:  ssh_host_rsa_key_pub,
		},
		{
			name: "ssh.host_keys:dsa",
			args: []string{"--render", "node1", "ssh.host_keys", "etc/ssh/ssh_host_rsa_key.ww"},
			log:  ssh_host_rsa_key,
		},
		{
			name: "ssh.host_keys:ed25519 pub",
			args: []string{"--render", "node1", "ssh.host_keys", "etc/ssh/ssh_host_ed25519_key.pub.ww"},
			log:  ssh_host_ed25519_key_pub,
		},
		{
			name: "ssh.host_keys:ed25519",
			args: []string{"--render", "node1", "ssh.host_keys", "etc/ssh/ssh_host_ed25519_key.ww"},
			log:  ssh_host_ed25519_key,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := show.GetCommand()
			cmd.SetArgs(tt.args)
			stdout := bytes.NewBufferString("")
			stderr := bytes.NewBufferString("")
			logbuf := bytes.NewBufferString("")
			cmd.SetOut(stdout)
			cmd.SetErr(stderr)
			wwlog.SetLogWriter(logbuf)
			err := cmd.Execute()
			assert.NoError(t, err)
			assert.Empty(t, stdout.String())
			assert.Empty(t, stderr.String())
			assert.Equal(t, tt.log, logbuf.String())
		})
	}
}

const ssh_host_dsa_key_pub string = `backupFile: true
writeFile: true
Filename: etc/ssh/ssh_host_dsa_key.pub
dsa pubkey sentinel
`

const ssh_host_dsa_key string = `backupFile: true
writeFile: true
Filename: etc/ssh/ssh_host_dsa_key
dsa key sentinel
`

const ssh_host_ecdsa_key_pub string = `backupFile: true
writeFile: true
Filename: etc/ssh/ssh_host_ecdsa_key.pub
ecdsa pubkey sentinel
`

const ssh_host_ecdsa_key string = `backupFile: true
writeFile: true
Filename: etc/ssh/ssh_host_ecdsa_key
ecdsa key sentinel
`

const ssh_host_ed25519_key_pub string = `backupFile: true
writeFile: true
Filename: etc/ssh/ssh_host_ed25519_key.pub
ed25519 pubkey sentinel
`

const ssh_host_ed25519_key string = `backupFile: true
writeFile: true
Filename: etc/ssh/ssh_host_ed25519_key
ed25519 key sentinel
`

const ssh_host_rsa_key_pub string = `backupFile: true
writeFile: true
Filename: etc/ssh/ssh_host_rsa_key.pub
rsa pubkey sentinel
`

const ssh_host_rsa_key string = `backupFile: true
writeFile: true
Filename: etc/ssh/ssh_host_rsa_key
rsa key sentinel
`
