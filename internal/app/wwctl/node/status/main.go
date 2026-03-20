package nodestatus

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/hostlist"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
	"golang.org/x/term"
)

type nodeStatus struct {
	NodeName string `json:"node name"`
	Stage    string `json:"stage"`
	Sent     string `json:"sent"`
	Ipaddr   string `json:"ipaddr"`
	Lastseen int64  `json:"last seen"`
	Security string `json:"security"`
}

const (
	formatStrHdr = "%-15s %-15s %-25s %-10s %-10s\n"
	formatStr    = "%-15s %-15s %-25s %-10d %-10s\n"
)

func displayStage(stage string) string {
	switch stage {
	case "efiboot":
		return "EFI"
	case "grub":
		return "GRUB"
	case "ipxe":
		return "IPXE"
	case "kernel":
		return "KERNEL"
	case "image":
		return "IMAGE"
	case "system":
		return "SYSTEM OVERLAY"
	case "runtime":
		return "RUNTIME OVERLAY"
	case "initramfs":
		return "INITRAMFS"
	default:
		return strings.ToUpper(stage)
	}
}

func CobraRunE(cmd *cobra.Command, args []string) (err error) {

	controller := warewulfconf.Get()

	if controller.Ipaddr == "" {
		return fmt.Errorf("warewulf Server IP Address is not properly configured")

	}

	for {
		var elipsis bool
		var height int
		var count int
		rightnow := time.Now().Unix()

		statusURL := fmt.Sprintf("http://%s:%d/status", controller.Ipaddr, controller.Warewulf.Port)
		wwlog.Verbose("Connecting to: %s", statusURL)

		resp, err := http.Get(statusURL)
		if err != nil {
			return fmt.Errorf("could not connect to Warewulf server: %w", err)
		}

		var wwNodeStatus struct {
			Nodes map[string]*nodeStatus `json:"nodes"`
		}
		err = json.NewDecoder(resp.Body).Decode(&wwNodeStatus)
		_ = resp.Body.Close()
		if err != nil {
			return fmt.Errorf("could not decode JSON: %w", err)
		}

		if SetWatch {
			fmt.Print("\033[H\033[2J")
			_, height, err = term.GetSize(0)
			if err != nil {
				wwlog.Warn("Could not get terminal height, using 24")
				height = 24
			}
		}

		fmt.Printf(formatStrHdr, "NODENAME", "STAGE", "SENT", "LASTSEEN", "SECURITY")
		fmt.Printf("%s\n", strings.Repeat("=", 80))

		wwlog.Verbose("Building sort index")
		var statuses []*nodeStatus
		if len(args) > 0 {
			nodeList := hostlist.Expand(args)
			for _, v := range wwNodeStatus.Nodes {
				for _, name := range nodeList {
					if v.NodeName == name {
						statuses = append(statuses, v)
						break
					}
				}
			}
		} else {
			for _, v := range wwNodeStatus.Nodes {
				statuses = append(statuses, v)
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
			wwlog.Debug("Reversing sort order")
			sort.Slice(statuses, func(i, j int) bool {
				return statuses[i].NodeName > statuses[j].NodeName
			})

		} else {
			sort.Slice(statuses, func(i, j int) bool {
				return statuses[i].NodeName < statuses[j].NodeName
			})
		}

		wwlog.Debug("Printing results")
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
					color.Red(formatStr, o.NodeName, displayStage(o.Stage), o.Sent, rightnow-o.Lastseen, o.Security)
				} else if rightnow-o.Lastseen >= int64(controller.Warewulf.UpdateInterval+5) {
					color.Yellow(formatStr, o.NodeName, displayStage(o.Stage), o.Sent, rightnow-o.Lastseen, o.Security)
				} else {
					fmt.Printf(formatStr, o.NodeName, displayStage(o.Stage), o.Sent, rightnow-o.Lastseen, o.Security)
				}
			} else {
				color.HiBlack(formatStrHdr, o.NodeName, "--", "--", "--", "--")
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
