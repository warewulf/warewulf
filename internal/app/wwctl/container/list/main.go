package list

import (
	"os"
	"strconv"
	"time"

	"github.com/hpcng/warewulf/internal/pkg/api/container"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	containerInfo, err := container.ContainerList()
	if err != nil {
		wwlog.Error("%s", err)
		return
	}

	tb := tablewriter.NewWriter(os.Stdout)
	tb.SetHeader([]string{"CONTAINER NAME", "NODES", "KERNEL VERSION", "CREATION TIME", "MODIFICATION TIME", "SIZE"})
	tb.SetAutoWrapText(false)
	tb.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
	tb.SetAlignment(tablewriter.ALIGN_LEFT)
	tb.SetCenterSeparator("")
	tb.SetColumnSeparator("")
	tb.SetRowSeparator("")
	tb.SetHeaderLine(false)
	tb.SetBorder(false)
	for i := 0; i < len(containerInfo); i++ {
		createTime := time.Unix(int64(containerInfo[i].CreateDate), 0)
		modTime := time.Unix(int64(containerInfo[i].ModDate), 0)
		tb.Append([]string{
			containerInfo[i].Name,
			strconv.FormatUint(uint64(containerInfo[i].NodeCount), 10),
			containerInfo[i].KernelVersion,
			createTime.Format(time.RFC822),
			modTime.Format(time.RFC822),
			util.ByteToString(int64(containerInfo[i].Size)),
		})
	}
	tb.Render()
	return
}
