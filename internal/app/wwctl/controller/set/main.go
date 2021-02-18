package set

import (
	"fmt"
	"os"

	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/manifoldco/promptui"
	"github.com/spf13/cobra"
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

	if len(args) == 0 {
		args = append(args, "localhost")
	}

	if SetAll == true {
		controllers, err = nodeDB.FindAllControllers()
		if err != nil {
			wwlog.Printf(wwlog.ERROR, "%s\n", err)
			os.Exit(1)
		}

	} else {
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
	}

	for _, c := range controllers {

		if SetComment != "" {
			wwlog.Printf(wwlog.VERBOSE, "Controller: %s, Setting Comment: %s\n", c.Id, SetComment)

			c.Comment = SetComment
			err := nodeDB.ControllerUpdate(c)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
		if SetFqdn != "" {
			wwlog.Printf(wwlog.VERBOSE, "Controller: %s, Setting FQDN: %s\n", c.Id, SetFqdn)

			c.Fqdn = SetFqdn
			err := nodeDB.ControllerUpdate(c)
			if err != nil {
				wwlog.Printf(wwlog.ERROR, "%s\n", err)
				os.Exit(1)
			}
		}
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
