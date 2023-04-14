package nodestatus

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	apinode "github.com/hpcng/warewulf/internal/pkg/api/node"
	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	warewulfconf "github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/hpcng/warewulf/pkg/hostlist"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {

	controller := warewulfconf.New()

	if controller.Ipaddr == "" {
		return fmt.Errorf("warewulf Server IP Address is not properly configured")

	}

	for {
		var elipsis bool
		var height int
		var count int
		rightnow := time.Now().Unix()

		var nodeStatusResponse *wwapiv1.NodeStatusResponse
		nodeStatusResponse, err = apinode.NodeStatus([]string{})
		if err != nil {
			return err
		}

		if SetWatch {
			fmt.Print("\033[H\033[2J")
			_, height, err = term.GetSize(0)
			if err != nil {
				wwlog.Warn("Could not get terminal height, using 24")
				height = 24
			}
		}

		fmt.Printf("%-20s %-20s %-25s %-10s\n", "NODENAME", "STAGE", "SENT", "LASTSEEN (s)")
		fmt.Printf("%s\n", strings.Repeat("=", 80))

		wwlog.Verbose("Building sort index")
		var statuses []*wwapiv1.NodeStatus
		if len(args) > 0 {
			nodeList := hostlist.Expand(args)
			for i := 0; i < len(nodeStatusResponse.NodeStatus); i++ {
				for j := 0; j < len(nodeList); j++ {
					if nodeStatusResponse.NodeStatus[i].NodeName == nodeList[j] {
						statuses = append(statuses, nodeStatusResponse.NodeStatus[i])
						break
					}
				}
			}
		} else {
			for i := 0; i < len(nodeStatusResponse.NodeStatus); i++ {
				statuses = append(statuses, nodeStatusResponse.NodeStatus[i])
			}
		}

		wwlog.Verbose("Sorting index")
		if SetSortLast {
			sort.Slice(statuses, func(i, j int) bool {
				if statuses[i].Lastseen > statuses[j].Lastseen {
					return true
				} else if statuses[i].Lastseen < statuses[j].Lastseen {
					return false
				} else {
					return statuses[i].NodeName < statuses[j].NodeName
				}
			})
		} else if SetSortReverse {
			wwlog.Verbose("Reversing sort order")
			sort.Slice(statuses, func(i, j int) bool {
				return statuses[i].NodeName > statuses[j].NodeName
			})

		} else {
			sort.Slice(statuses, func(i, j int) bool {
				return statuses[i].NodeName < statuses[j].NodeName
			})
		}

		wwlog.Verbose("Printing results")
		for i := 0; i < len(statuses); i++ {
			o := statuses[i]
			if SetTime > 0 && o.Lastseen < SetTime {
				continue
			}

			if o.Lastseen > 0 {
				if SetUnknown {
					continue
				}
				if rightnow-o.Lastseen >= int64(controller.Warewulf.UpdateInterval*2) {
					color.Red("%-20s %-20s %-25s %-10d\n", o.NodeName, o.Stage, o.Sent, rightnow-o.Lastseen)
				} else if rightnow-o.Lastseen >= int64(controller.Warewulf.UpdateInterval+5) {
					color.Yellow("%-20s %-20s %-25s %-10d\n", o.NodeName, o.Stage, o.Sent, rightnow-o.Lastseen)
				} else {
					fmt.Printf("%-20s %-20s %-25s %-10d\n", o.NodeName, o.Stage, o.Sent, rightnow-o.Lastseen)
				}
			} else {
				color.HiBlack("%-20s %-20s %-25s %-10s\n", o.NodeName, "--", "--", "--")
			}
			if count+4 >= height && SetWatch {
				if count+1 != len(statuses) {
					elipsis = true
				}
				break
			}
			count++
		}

		if SetWatch {
			if elipsis {
				fmt.Printf("... ")
			}
			time.Sleep(time.Duration(SetUpdate) * time.Millisecond)
		} else {
			break
		}
	}
	return
}
