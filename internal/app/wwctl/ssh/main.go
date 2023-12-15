package ssh

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/batch"
	"github.com/warewulf/warewulf/internal/pkg/hostlist"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	batchpool := batch.New(FanOut)

	nodeDB, err := node.New()
	if err != nil {
		wwlog.Error("Could not open node configuration: %s", err)
		os.Exit(1)
	}

	nodes, err := nodeDB.FindAllNodes()
	if err != nil {
		wwlog.Error("Could not get node list: %s", err)
		os.Exit(1)
	}

	if len(args) > 0 {
		nodes = node.FilterByName(nodes, hostlist.Expand(args))
	} else {
		//nolint:errcheck
		cmd.Usage()
		os.Exit(1)
	}
	for _, node := range nodes {
		var primaryNet string
		for netName := range node.NetDevs {
			if node.NetDevs[netName].Primary() {
				primaryNet = netName
				break
			}
		}
		if primaryNet == "" {
			wwlog.Error("%s: Primary network device doesn't exist\n", node.Id())
			continue
		}
		if node.NetDevs[primaryNet].Ipaddr.IsUnspecified() {
			wwlog.Error("%s: Primary network IP address not configured\n", node.Id())
			continue
		}

		nodename := node.Id()
		var command []string

		command = append(command, node.NetDevs[primaryNet].Ipaddr.String())
		command = append(command, args[1:]...)

		batchpool.Submit(func() {

			if DryRun {
				fmt.Printf("%s: %s %s\n", nodename, SshPath, strings.Join(command, " "))
			} else {

				wwlog.Debug("Sending command to node '%s': %s", nodename, command)
				var stdout, stderr bytes.Buffer
				cmd := exec.Command(SshPath, command...)
				cmd.Stdin = os.Stdin
				cmd.Stdout = &stdout
				cmd.Stderr = &stderr
				_ = cmd.Run()

				scan_stdout := bufio.NewScanner(&stdout)
				for scan_stdout.Scan() {
					fmt.Printf("%s: %s\n", nodename, scan_stdout.Text())
				}

				scan_stderr := bufio.NewScanner(&stderr)
				for scan_stderr.Scan() {
					fmt.Fprintf(os.Stderr, "%s: %s\n", nodename, scan_stderr.Text())
				}
			}
			//util.ExecInteractive(SshPath, command...)
			time.Sleep(time.Duration(Sleep) * time.Second)

		})

	}

	batchpool.Run()

	return nil
}
