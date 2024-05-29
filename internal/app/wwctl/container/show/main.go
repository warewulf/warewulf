package show

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/container"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {

	r, err := container.Show(&container.ShowParameter{
		Name: args[0],
	})
	if err != nil {
		return
	}

	if !ShowAll {
		fmt.Printf("%s\n", r.Rootfs)
	} else {
		kernelVersion := r.KernelVersion
		if kernelVersion == "" {
			kernelVersion = "not found"
		}
		fmt.Printf("Name: %s\n", r.Name)
		fmt.Printf("KernelVersion: %s\n", kernelVersion)
		fmt.Printf("Rootfs: %s\n", r.Rootfs)
		fmt.Printf("Nr nodes: %d\n", len(r.Nodes))
		fmt.Printf("Nodes: %s\n", r.Nodes)
	}
	return
}
