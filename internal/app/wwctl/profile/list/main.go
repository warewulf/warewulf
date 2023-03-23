package list

import (
	"strings"

	"github.com/hpcng/warewulf/internal/app/wwctl/helper"
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

	if len(profileInfo.Output) > 0 {
		ph := helper.NewPrintHelper(strings.Split(profileInfo.Output[0], "="))
		for _, val := range profileInfo.Output[1:] {
			ph.Append(strings.Split(val, "="))
		}
		ph.Render()
	}
	return
}
