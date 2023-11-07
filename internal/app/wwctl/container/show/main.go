package show

import (
	"github.com/hpcng/warewulf/internal/pkg/api/container"
	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"

	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {

	csp := &wwapiv1.ContainerShowParameter{
		ContainerName: args[0],
	}

	var r *wwapiv1.ContainerShowResponse
	r, err = container.ContainerShow(csp)
	if err != nil {
		return
	}

	if !ShowAll {
		wwlog.Info("%s\n", r.Rootfs)
	} else {
		kernelVersion := r.KernelVersion
		if kernelVersion == "" {
			kernelVersion = "not found"
		}
		wwlog.Info("Name: %s\n", r.Name)
		wwlog.Info("KernelVersion: %s\n", kernelVersion)
		wwlog.Info("Rootfs: %s\n", r.Rootfs)
		wwlog.Info("Nr nodes: %d\n", len(r.Nodes))
		wwlog.Info("Nodes: %s\n", r.Nodes)
	}
	return
}
