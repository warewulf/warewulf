package list

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"github.com/spf13/cobra"
	"os"
	"reflect"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	nodeDB, err := node.New()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not open node configuration: %s\n", err)
		os.Exit(1)
	}

	groups, err := nodeDB.FindAllGroups()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not find all nodes: %s\n", err)
		os.Exit(1)
	}


	if ShowAll == true {
		for _, group := range groups {
			v := reflect.ValueOf(group)
			typeOfS := v.Type()
			fmt.Printf("################################################################################\n")
			for i := 0; i< v.NumField(); i++ {
				fmt.Printf("%-25s %s = %v\n", group.Id, typeOfS.Field(i).Name, v.Field(i).Interface())
			}
		}
	} else {
		fmt.Printf("%-22s %-16s %-30s %-30s %-16s\n", "GROUP NAME", "DOMAINNAME", "KERNEL VERSION", "VNFS IMAGE", "RUNTIME OVERLAY")
		for _, g := range groups {
			fmt.Printf("%-22s %-16s %-30s %-30s %-16s\n", g.GroupName, g.DomainName, g.KernelVersion, g.Vnfs, g.RuntimeOverlay)
		}
	}


	return nil
}
