package nodestatus

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"sort"
	"strconv"
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

// statusHTTPTimeout bounds how long "wwctl node status" waits on the
// warewulfd /status endpoint. The request would otherwise use
// http.DefaultClient, which has no timeout, so an unreachable or
// firewalled server would hang the command indefinitely.
const statusHTTPTimeout = 30 * time.Second

// statusURL builds the warewulfd /status endpoint URL from the server
// configuration. It prefers the IPv4 server address (ipaddr) and falls
// back to the IPv6 address (ipaddr6), so that node status works on
// IPv6-only servers where ipaddr is left unset. net.JoinHostPort
// brackets IPv6 literals correctly (e.g. [2001:db8::1]:9873).
//
// The IPv4-first order preserves the previous node status behavior,
// which only ever used ipaddr. Address selection is currently duplicated
// across the tree (wwclient prefers ipaddr6; warewulfd is request-driven);
// consolidating it behind a shared config method is left as a follow-up.
func statusURL(controller *warewulfconf.WarewulfYaml) (string, error) {
	serverAddr := controller.Ipaddr
	if serverAddr == "" {
		serverAddr = controller.Ipaddr6
	}

	if serverAddr == "" {
		confFile := controller.GetWarewulfConf()
		if confFile == "" {
			confFile = "warewulf.conf"
		}
		return "", fmt.Errorf("warewulf server address is not configured: set ipaddr or ipaddr6 in %s", confFile)
	}

	hostPort := net.JoinHostPort(serverAddr, strconv.Itoa(controller.Warewulf.Port))
	return fmt.Sprintf("http://%s/status", hostPort), nil
}

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	controller := warewulfconf.Get()

	endpoint, err := statusURL(controller)
	if err != nil {
		return err
	}

	client := &http.Client{Timeout: statusHTTPTimeout}

	for {
		var elipsis bool
		var height int
		var count int
		rightnow := time.Now().Unix()

		wwlog.Verbose("Connecting to: %s", endpoint)

		resp, err := client.Get(endpoint)
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
