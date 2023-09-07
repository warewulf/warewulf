package apiprofile

import (
	"fmt"
	"sort"

	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
)

/*
Returns the formatted list of profiles as string
*/
func ProfileList(ShowOpt *wwapiv1.GetProfileList) (profileList wwapiv1.ProfileList, err error) {
	profileList.Output = []string{}
	nodeDB, err := node.New()
	if err != nil {
		wwlog.Error("Could not open node configuration: %s", err)
		return
	}

	profiles, err := nodeDB.FindAllProfiles()
	if err != nil {
		wwlog.Error("Could not find all profiles: %s", err)
		return
	}
	profiles = node.FilterByName(profiles, ShowOpt.Profiles)
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Id.Get() < profiles[j].Id.Get()
	})
	if ShowOpt.ShowAll || ShowOpt.ShowFullAll {
		for _, p := range profiles {
			profileList.Output = append(profileList.Output,
				fmt.Sprintf("%s=%s=%s=%s", "PROFILE", "FIELD", "PROFILE", "VALUE"))
			fields := p.GetFields(ShowOpt.ShowFullAll)
			for _, f := range fields {
				profileList.Output = append(profileList.Output,
					fmt.Sprintf("%s=%s=%s=%s", p.Id.Print(), f.Field, f.Source, f.Value))
			}
		}
	} else {
		profileList.Output = append(profileList.Output,
			fmt.Sprintf("%s=%s", "PROFILE NAME", "COMMENT/DESCRIPTION"))

		for _, profile := range profiles {
			profileList.Output = append(profileList.Output,
				fmt.Sprintf("%s=%s", profile.Id.Print(), profile.Comment.Print()))
		}
	}
	return
}
