package list

import (
	"fmt"

	apiprofile "github.com/hpcng/warewulf/internal/pkg/api/profile"

	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	req := wwapiv1.GetProfileList{
		ShowAll:  ShowAll,
		Profiles: args,
	}
	profileInfo, err := apiprofile.ProfileList(&req)
	if err != nil {
		return
	}
	for _, str := range profileInfo.Output {
		fmt.Printf("%s\n", str)
	}
	return
}
