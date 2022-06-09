package ssh

import (
	"bufio"
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/hpcng/warewulf/internal/pkg/batch"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/hpcng/warewulf/pkg/hostlist"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	batchpool := batch.New(FanOut)

	nodeDB, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
		os.Exit(1)
	}

	nodes, err := nodeDB.FindAllNodes()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not get node list: %s\n", err)
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

		if _, ok := node.NetDevs["default"]; !ok {
			fmt.Fprintf(os.Stderr, "%s: Default network device doesn't exist\n", node.Id.Get())
			continue
		}

		if node.NetDevs["default"].Ipaddr.Get() == "" {
			fmt.Fprintf(os.Stderr, "%s: Default network IP address not configured\n", node.Id.Get())
			continue
		}

		nodename := node.Id.Print()
		var command []string

		command = append(command, node.NetDevs["default"].Ipaddr.Get())
		command = append(command, args[1:]...)

		batchpool.Submit(func() {

			if DryRun {
				fmt.Printf("%s: %s %s\n", nodename, SshPath, strings.Join(command, " "))
			} else {

				wwlog.Printf(wwlog.DEBUG, "Sending command to node '%s': %s\n", nodename, command)
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
