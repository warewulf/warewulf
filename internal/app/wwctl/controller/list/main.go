package list

import "C"
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

	controllers, err := nodeDB.FindAllControllers()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not find all nodes: %s\n", err)
		os.Exit(1)
	}


	if ShowAll == true {
		for _, controller := range controllers {
			v := reflect.ValueOf(controller)
			typeOfS := v.Type()
			fmt.Printf("################################################################################\n")
			for i := 0; i< v.NumField(); i++ {
				fmt.Printf("%-25s %s = %v\n", controller.Id, typeOfS.Field(i).Name, v.Field(i).Interface())
			}
		}
	} else {
		fmt.Printf("%-22s\n", "CONTROLLER NAME")
		for _, c := range controllers {
			fmt.Printf("%-22s\n", c.Id)
		}
	}


	return nil
}