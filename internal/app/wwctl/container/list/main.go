package list

import (
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/warewulf/warewulf/internal/app/wwctl/table"
	apicontainer "github.com/warewulf/warewulf/internal/pkg/api/container"
	"github.com/warewulf/warewulf/internal/pkg/container"
	"github.com/warewulf/warewulf/internal/pkg/util"
)

var containerList = apicontainer.ContainerList

func CobraRunE(vars *variables) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		t := table.New(cmd.OutOrStdout())
		showSize := vars.size || vars.chroot || vars.compressed
		if showSize || vars.full || vars.kernel {
			containerInfo, err := containerList()
			if err != nil {
				return err
			}
			if vars.full {
				t.AddHeader("CONTAINER NAME", "NODES", "KERNEL VERSION", "CREATION TIME", "MODIFICATION TIME", "SIZE")
				for i := 0; i < len(containerInfo); i++ {
					createTime := time.Unix(int64(containerInfo[i].CreateDate), 0)
					modTime := time.Unix(int64(containerInfo[i].ModDate), 0)
					sz := util.ByteToString(int64(containerInfo[i].ImgSize))
					if vars.compressed {
						sz = util.ByteToString(int64(containerInfo[i].ImgSizeComp))
					}
					if vars.chroot {
						sz = util.ByteToString(int64(containerInfo[i].Size))
					}
					t.AddLine(
						containerInfo[i].Name,
						strconv.FormatUint(uint64(containerInfo[i].NodeCount), 10),
						containerInfo[i].KernelVersion,
						createTime.Format(time.RFC822),
						modTime.Format(time.RFC822),
						sz,
					)
				}
			} else if vars.kernel {
				t.AddHeader("CONTAINER NAME", "NODES", "KERNEL VERSION")
				for i := 0; i < len(containerInfo); i++ {
					t.AddLine(
						containerInfo[i].Name,
						strconv.FormatUint(uint64(containerInfo[i].NodeCount), 10),
						containerInfo[i].KernelVersion,
					)
				}
			} else if showSize {
				t.AddHeader("CONTAINER NAME", "NODES", "SIZE")
				for i := 0; i < len(containerInfo); i++ {
					sz := util.ByteToString(int64(containerInfo[i].ImgSize))
					if vars.compressed {
						sz = util.ByteToString(int64(containerInfo[i].ImgSizeComp))
					}
					if vars.chroot {
						sz = util.ByteToString(int64(containerInfo[i].Size))
					}

					t.AddLine(
						containerInfo[i].Name,
						strconv.FormatUint(uint64(containerInfo[i].NodeCount), 10),
						sz,
					)
				}
			}
		} else {
			t.AddHeader("CONTAINER NAME")
			list, _ := container.ListSources()
			for _, cont := range list {
				t.AddLine(cont)
			}
		}
		t.Print()
		return
	}
}
