package power

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/batch"
	"github.com/warewulf/warewulf/internal/pkg/bmc"
	"github.com/warewulf/warewulf/internal/pkg/hostlist"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

type Action struct {
	Name   string
	BmcCmd string
	Short  string
	Long   string
}

var Actions = []Action{
	{"on", "PowerOn", "Power on the given node(s)", "This command will power on a set of nodes specified by PATTERN."},
	{"off", "PowerOff", "Power off the given node(s)", "This command will shutdown power to a set of nodes specified by PATTERN."},
	{"cycle", "PowerCycle", "Power cycle the given node(s)", "This command cycles power for a set of nodes specified by PATTERN."},
	{"reset", "PowerReset", "Issue a reset to node(s)", "This command will issue a reset to a set of nodes specified by PATTERN."},
	{"soft", "PowerSoft", "Gracefully shuts down the given node(s)", "This command uses the operating system to shut down the set of nodes specified by PATTERN."},
	{"status", "PowerStatus", "Show power status for the given node(s)", "This command displays the power status of a set of nodes specified by PATTERN."},
}

func lookupAction(name string) (Action, bool) {
	for _, action := range Actions {
		if action.Name == name {
			return action, true
		}
	}
	return Action{}, false
}

func actionNames() []string {
	names := make([]string, len(Actions))
	for i, action := range Actions {
		names[i] = action.Name
	}
	return names
}

func validateAction(cmd *cobra.Command, args []string) error {
	if _, ok := lookupAction(args[0]); !ok {
		return fmt.Errorf("invalid action %q, must be one of: %s", args[0], strings.Join(actionNames(), ", "))
	}
	return nil
}

func CobraRunE(vars *variables) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		var returnErr error = nil

		if err := validateAction(cmd, args); err != nil {
			return err
		}
		action, _ := lookupAction(args[0])

		nodeDB, err := node.New()
		if err != nil {
			return fmt.Errorf("could not open node configuration: %s", err)
		}

		nodes, err := nodeDB.FindAllNodes()
		if err != nil {
			return fmt.Errorf("could not get node list: %s", err)
		}

		nodes = node.FilterNodeListByName(nodes, hostlist.Expand(args[1:]))
		if len(nodes) == 0 {
			return fmt.Errorf("no nodes found")
		}

		batchpool := batch.New(vars.Fanout)
		jobcount := len(nodes)
		results := make(chan bmc.TemplateStruct, jobcount)

		for _, node := range nodes {
			if node.Ipmi == nil || node.Ipmi.Ipaddr == nil || node.Ipmi.Ipaddr.IsUnspecified() {
				wwlog.Error("%s: No IPMI IP address", node.Id())
				returnErr = fmt.Errorf("one or more nodes have no IPMI IP address")
				continue
			}
			ipmiCmd := bmc.TemplateStruct{
				IpmiConf: *node.Ipmi,
				ShowOnly: vars.Showcmd,
			}
			batchpool.Submit(func() {
				//nolint:errcheck
				ipmiCmd.Command(action.BmcCmd)
				results <- ipmiCmd
			})
		}

		batchpool.Run()

		close(results)

		for result := range results {
			out, err := result.Result()
			if err != nil {
				wwlog.Error("%s: %s", result.Ipaddr, out)
				returnErr = err
				continue
			}
			wwlog.Info("%s: %s\n", result.Ipaddr, out)
		}

		return returnErr
	}
}
