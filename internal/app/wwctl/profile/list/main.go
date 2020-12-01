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

	profiles, err := nodeDB.FindAllProfiles()
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "Could not find all nodes: %s\n", err)
		os.Exit(1)
	}

	for _, group := range profiles {
		v := reflect.ValueOf(group)
		typeOfS := v.Type()
		fmt.Printf("################################################################################\n")
		for i := 0; i< v.NumField(); i++ {
			fmt.Printf("%-25s %s = %v\n", group.Id, typeOfS.Field(i).Name, v.Field(i).Interface())
		}
	}

	return nil
}