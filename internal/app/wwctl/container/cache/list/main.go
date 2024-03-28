package cachelist

import (
	"context"

	"github.com/opencontainers/umoci"
	"github.com/spf13/cobra"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(vars *variables) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		/*
			containerInfo, err := containerList()
			if err != nil {
				wwlog.Error("%s", err)
				return
			}

			ph := helper.NewPrintHelper([]string{"CONTAINER NAME", "NODES", "KERNEL VERSION", "CREATION TIME", "MODIFICATION TIME", "SIZE"})
			for i := 0; i < len(containerInfo); i++ {
				createTime := time.Unix(int64(containerInfo[i].CreateDate), 0)
				modTime := time.Unix(int64(containerInfo[i].ModDate), 0)
				ph.Append([]string{
					containerInfo[i].Name,
					strconv.FormatUint(uint64(containerInfo[i].NodeCount), 10),
					containerInfo[i].KernelVersion,
					createTime.Format(time.RFC822),
					modTime.Format(time.RFC822),
					util.ByteToString(int64(containerInfo[i].Size)),
				})
			}
			ph.Render()
		*/
		eng, err := umoci.OpenLayout(warewulfconf.Get().Warewulf.DataStore + "/oci")
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		if !vars.allblobs {
			refs, err := eng.ListReferences(ctx)
			if err != nil {
				return err
			}
			for _, ref := range refs {
				wwlog.Output("ref: %v\n", ref)
			}
		} else {
			blobs, err := eng.ListBlobs(ctx)
			if err != nil {
				return err
			}
			for _, blob := range blobs {
				wwlog.Output(string(blob))
			}
		}

		return
	}
}
