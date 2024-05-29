package list

import (
	"strconv"
	"time"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/helper"
	"github.com/warewulf/warewulf/internal/pkg/container"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

var containerList = container.List

func CobraRunE(vars *variables) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
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
		return
	}
}
