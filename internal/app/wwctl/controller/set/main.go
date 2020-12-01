package set

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
	"os"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	var err error
	var controllers []node.ControllerInfo

	nodeDB, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
		os.Exit(1)
	}

	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Cloud not get nodeList: %s\n", err)
		os.Exit(1)
	}

	if SetAll == true {
		var tmp []node.ControllerInfo
		tmp, err = nodeDB.FindAllControllers()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}

		for _, c := range tmp {
			controllers = append(controllers, c)
		}

	} else if len(args) > 0 {
		var tmp []node.ControllerInfo
		tmp, err = nodeDB.FindAllControllers()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}

		for _, a := range args {
			for _, c := range tmp {
				if c.Id == a {
					controllers = append(controllers, c)
				}
			}
		}

	} else {
		cmd.Usage()
		os.Exit(1)
	}

	for _, c := range controllers {

		if SetIpaddr != "" {
			wwlog.Printf(wwlog.VERBOSE, "Controller: %s, Setting IP Addr to: %s\n", c.Id, SetIpaddr)

			c.Ipaddr = SetIpaddr
			err := nodeDB.ControllerUpdate(c)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}

	}

	if len(controllers) > 0 {
		q := fmt.Sprintf("Are you sure you want to modify %d group(s)", len(controllers))

		prompt := promptui.Prompt{
			Label:     q,
			IsConfirm: true,
		}

		result, _ := prompt.Run()

		if result == "y" || result == "yes" {
			nodeDB.Persist()
		}

	} else {
		fmt.Printf("No controllers found\n")
	}

	return nil
}
