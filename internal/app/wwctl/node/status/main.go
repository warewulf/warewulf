package nodestatus

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/hpcng/warewulf/internal/pkg/warewulfconf"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"golang.org/x/term"
)

type allStatus struct {
	Nodes map[string]*NodeStatus `json:"nodes"`
}

type NodeStatus struct {
	Stage    string `json:"stage"`
	Sent     string `json:"sent"`
	Ipaddr   string `json:"ipaddr"`
	Lastseen int64  `json:"last seen"`
}

func CobraRunE(cmd *cobra.Command, args []string) error {

	controller, err := warewulfconf.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}

	if controller.Ipaddr == "" {
		wwlog.Printf(wwlog.ERROR, "The Warewulf Server IP Address is not properly configured\n")
		os.Exit(1)
	}

	statusURL := fmt.Sprintf("http://%s:%d/status", controller.Ipaddr, controller.Warewulf.Port)

	for {
		var elipsis bool
		var height int
		var count int
		rightnow := time.Now().Unix()

		wwlog.Printf(wwlog.VERBOSE, "Connecting to: %s\n", statusURL)

		resp, err := http.Get(statusURL)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Could not connect to Warewulf server: %s\n", err)
			os.Exit(1)
		}
		defer resp.Body.Close()

		decoder := json.NewDecoder(resp.Body)
		var nodeStatus allStatus

		err = decoder.Decode(&nodeStatus)
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "Could not decode JSON: %s\n", err)
			os.Exit(1)
		}

		if SetWatch {
			fmt.Print("\033[H\033[2J")
			_, height, err = term.GetSize(0)
			if err != nil {
				wwlog.Printf(wwlog.WARN, "Could not get terminal height, using 24\n")
				height = 24
			}
		}

		fmt.Printf("%-20s %-20s %-25s %-10s\n", "NODENAME", "STATUS", "SENT", "LASTSEEN (s)")
		fmt.Printf("%s\n", strings.Repeat("=", 80))

		keys := make([]string, 0, len(nodeStatus.Nodes))
		for k := range nodeStatus.Nodes {
			keys = append(keys, k)
		}

		sort.Strings(keys)

		for _, id := range keys {
			if nodeStatus.Nodes[id].Lastseen > 0 {
				fmt.Printf("%-20s %-20s %-25s %-10d\n", id, nodeStatus.Nodes[id].Stage, nodeStatus.Nodes[id].Sent, rightnow-nodeStatus.Nodes[id].Lastseen)
			} else {
				fmt.Printf("%-20s %-20s %-25s %-10s\n", id, "--", "--", "--")
			}
			if count+4 >= height && SetWatch {
				if count+1 != len(keys) {
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

	return nil
}
