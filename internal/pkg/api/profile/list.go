package apiprofile

import (
	"fmt"
	"sort"

	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/node"
)

/*
Returns the formatted list of profiles as string
*/
func ProfileList(ShowOpt *wwapiv1.GetProfileList) (*wwapiv1.ProfileListResponse, error) {
	nodeDB, err := node.New()
	if err != nil {
		return nil, fmt.Errorf("could not open node configuration: %s", err)
	}

	profiles, err := nodeDB.FindAllProfiles()
	if err != nil {
		return nil, fmt.Errorf("could not find all profiles: %s", err)
	}
	profiles = node.FilterByName(profiles, ShowOpt.Profiles)
	sort.Slice(profiles, func(i, j int) bool {
		return profiles[i].Id.Get() < profiles[j].Id.Get()
	})

	var entries []*wwapiv1.ProfileListEntry
	if ShowOpt.ShowAll || ShowOpt.ShowFullAll {
		for _, p := range profiles {
			fields := p.GetFields(ShowOpt.ShowFullAll)
			for _, f := range fields {
				entries = append(entries, &wwapiv1.ProfileListEntry{
					ProfileEntry: &wwapiv1.ProfileListEntry_ProfileFull{
						ProfileFull: &wwapiv1.ProfileListFull{
							ProfileName: p.Id.Print(),
							Field:       f.Field,
							Source:      f.Source,
							Value:       f.Value,
						},
					},
				})
			}
		}
	} else {
		for _, p := range profiles {
			entries = append(entries, &wwapiv1.ProfileListEntry{
				ProfileEntry: &wwapiv1.ProfileListEntry_ProfileSimple{
					ProfileSimple: &wwapiv1.ProfileListSimple{
						ProfileName: p.Id.Print(),
						Comment:     p.Comment.Print(),
					},
				},
			})
		}
	}

	return &wwapiv1.ProfileListResponse{Profiles: entries}, nil
}
