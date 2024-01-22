package apiprofile

import (
	"fmt"
	"sort"

	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/node"
	"github.com/hpcng/warewulf/internal/pkg/profile"
)

/*
Returns the formatted list of profiles as string
*/
func ProfileList(ShowOpt *wwapiv1.GetProfileList) (*profile.ProfileListResponse, error) {
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

	resp := &profile.ProfileListResponse{
		Profiles: make(map[string][]profile.ProfileListEntry),
	}
	if ShowOpt.ShowAll || ShowOpt.ShowFullAll {
		for _, p := range profiles {
			fields := p.GetFields(ShowOpt.ShowFullAll)

			var entries []profile.ProfileListEntry
			for _, f := range fields {
				entries = append(entries, &profile.ProfileListLongEntry{
					Field:   f.Field,
					Profile: f.Source,
					Value:   f.Value,
				})
			}

			if vals, ok := resp.Profiles[p.Id.Print()]; ok {
				entries = append(entries, vals...)
			}
			resp.Profiles[p.Id.Print()] = entries
		}
	} else {
		for _, p := range profiles {
			var entries []profile.ProfileListEntry
			entries = append(entries, &profile.ProfileListSimpleEntry{
				CommentDesc: p.Comment.Print(),
			})

			if vals, ok := resp.Profiles[p.Id.Print()]; ok {
				entries = append(entries, vals...)
			}
			resp.Profiles[p.Id.Print()] = entries
		}

	}

	return resp, nil
}
