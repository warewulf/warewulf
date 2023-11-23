package apinode

import (
	"encoding/hex"

	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

func Hash() *wwapiv1.NodeDBHash {
	config, err := node.New()
	if err != nil {
		wwlog.Warn("couldb't read config")
	}
	hash := config.Hash()
	return &wwapiv1.NodeDBHash{
		Hash: hex.EncodeToString(hash[:]),
	}
}
