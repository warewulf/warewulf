package apiprofile

import (
	"fmt"
	"sort"

	"github.com/warewulf/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/warewulf/warewulf/internal/pkg/node"
)

/*
Returns the formatted list of profiles as string
*/
func ProfileList(ShowOpt *wwapiv1.GetProfileList) (profileList wwapiv1.ProfileList, err error) {
	profileList.Output = []string{}
	nodeDB, err := node.New()
	if err != nil {
		return
	}
	profiles, err := nodeDB.FindAllProfiles()
	if err != nil {
		return
	}
	profiles = node.FilterByName(profiles, ShowOpt.Profiles)
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Id() < profiles[j].Id()
	})
	if ShowOpt.ShowAll || ShowOpt.ShowFullAll {
		for _, p := range profiles {
			profileList.Output = append(profileList.Output,
				fmt.Sprintf("%s=%s=%s=%s", "PROFILE", "FIELD", "PROFILE", "VALUE"))
			fields := nodeDB.GetFields(p, ShowOpt.ShowFullAll)
			for _, f := range fields {
				profileList.Output = append(profileList.Output,
					fmt.Sprintf("%s=%s=%s=%s", p.Id(), f.Field, f.Source, f.Value))
			}
		}
	} else {
		profileList.Output = append(profileList.Output,
			fmt.Sprintf("%s=%s", "PROFILE NAME", "COMMENT/DESCRIPTION"))

		for _, profile := range profiles {
			profileList.Output = append(profileList.Output,
				fmt.Sprintf("%s=%s", profile.Id(), node.PrintVal(profile.Comment)))
		}
	}
	return
}
