package api

import (
	"context"
	"fmt"

	"dario.cat/mergo"
	"github.com/swaggest/usecase"
	"github.com/swaggest/usecase/status"
	"github.com/warewulf/warewulf/internal/pkg/image"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func getProfiles() usecase.Interactor {
	u := usecase.NewInteractor(func(ctx context.Context, _ struct{}, output *map[string]*node.Profile) error {
		wwlog.Debug("api.getProfiles()")
		if registry, err := node.New(); err != nil {
			return err
		} else {
			*output = registry.NodeProfiles
			return nil
		}
	})
	u.SetTitle("Get profiles")
	u.SetDescription("Get all node profiles.")
	u.SetTags("Profile")
	return u
}

func getProfileByID() usecase.Interactor {
	type getProfileByIDInput struct {
		ID string `path:"id" required:"true" description:"ID of profile to get"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input getProfileByIDInput, output *node.Profile) error {
		wwlog.Debug("api.getProfileByID(ID:%v)", input.ID)
		if registry, err := node.New(); err != nil {
			return err
		} else {
			if profile, err := registry.GetProfile(input.ID); err != nil {
				return status.Wrap(fmt.Errorf("profile not found: %v (%v)", input.ID, err), status.NotFound)
			} else {
				*output = profile
				return nil
			}
		}
	})
	u.SetTitle("Get a profile")
	u.SetDescription("Get a node profile by its ID.")
	u.SetTags("Profile")
	u.SetExpectedErrors(status.NotFound)
	return u
}

func addProfile() usecase.Interactor {
	type addProfileInput struct {
		ID      string       `path:"id" required:"true" description:"ID of profile to add"`
		Profile node.Profile `json:"profile" required:"true" description:"Field values in JSON format for added profile"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input addProfileInput, output *node.Profile) error {
		wwlog.Debug("api.addProfile(ID:%v, Profile:%+v)", input.ID, input.Profile)
		if registry, err := node.New(); err != nil {
			return err
		} else {
			for _, profile := range input.Profile.Profiles {
				if _, ok := registry.NodeProfiles[profile]; !ok {
					return status.Wrap(fmt.Errorf("profile '%s' does not exist", profile), status.InvalidArgument)
				}
			}
			if input.Profile.ImageName != "" && !image.ValidSource(input.Profile.ImageName) {
				return status.Wrap(fmt.Errorf("image '%s' does not exist", input.Profile.ImageName), status.InvalidArgument)
			}
			for _, overlay_ := range input.Profile.SystemOverlay {
				if !overlay.GetOverlay(overlay_).Exists() {
					return status.Wrap(fmt.Errorf("overlay '%s' does not exist", overlay_), status.InvalidArgument)
				}
			}
			for _, overlay_ := range input.Profile.RuntimeOverlay {
				if !overlay.GetOverlay(overlay_).Exists() {
					return status.Wrap(fmt.Errorf("overlay '%s' does not exist", overlay_), status.InvalidArgument)
				}
			}
			registry.NodeProfiles[input.ID] = &input.Profile
			if err := registry.Persist(); err != nil {
				return err
			}
			warewulfd.Reload()
			*output = *(registry.NodeProfiles[input.ID])
			return nil
		}
	})
	u.SetTitle("Add a profile")
	u.SetDescription("Add a new node profile.")
	u.SetTags("Profile")

	return u
}

func updateProfile() usecase.Interactor {
	type updateProfileInput struct {
		ID      string       `path:"id" required:"true" description:"ID of profile to update"`
		Profile node.Profile `json:"profile" required:"true" description:"Field values in JSON format to update on profile"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input updateProfileInput, output *node.Profile) error {
		wwlog.Debug("api.updateProfile(ID:%v, Profile:%+v)", input.ID, input.Profile)
		if registry, err := node.New(); err != nil {
			return err
		} else {
			for _, profile := range input.Profile.Profiles {
				if _, ok := registry.NodeProfiles[profile]; !ok {
					return status.Wrap(fmt.Errorf("profile '%s' does not exist", profile), status.InvalidArgument)
				}
			}
			if input.Profile.ImageName != "" && !image.ValidSource(input.Profile.ImageName) {
				return status.Wrap(fmt.Errorf("image '%s' does not exist", input.Profile.ImageName), status.InvalidArgument)
			}
			for _, overlay_ := range input.Profile.SystemOverlay {
				if !overlay.GetOverlay(overlay_).Exists() {
					return status.Wrap(fmt.Errorf("overlay '%s' does not exist", overlay_), status.InvalidArgument)
				}
			}
			for _, overlay_ := range input.Profile.RuntimeOverlay {
				if !overlay.GetOverlay(overlay_).Exists() {
					return status.Wrap(fmt.Errorf("overlay '%s' does not exist", overlay_), status.InvalidArgument)
				}
			}
			if profilePtr, err := registry.GetProfilePtr(input.ID); err != nil {
				return status.Wrap(err, status.NotFound)
			} else {
				if err := mergo.MergeWithOverwrite(profilePtr, &input.Profile); err != nil {
					return err
				}
				if err := registry.Persist(); err != nil {
					return err
				}
				warewulfd.Reload()
				*output = *profilePtr
				return nil
			}
		}
	})
	u.SetTitle("Update a profile")
	u.SetDescription("Update an existing node profile.")
	u.SetTags("Profile")

	return u
}

func deleteProfile() usecase.Interactor {
	type deleteProfileInput struct {
		ID string `path:"id" required:"true" description:"ID of profile to delete"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input deleteProfileInput, output *node.Profile) error {
		wwlog.Debug("api.deleteProfile(ID:%v)", input.ID)
		if registry, err := node.New(); err != nil {
			return err
		} else {
			if profile, ok := registry.NodeProfiles[input.ID]; ok {
				*output = *profile
			}

			nodesCount := len(registry.ListNodesUsingProfile(input.ID))
			profilesCount := len(registry.ListProfilesUsingProfile(input.ID))
			if nodesCount > 0 || profilesCount > 0 {
				return status.Wrap(fmt.Errorf(
					"profile '%s' is in use by %v nodes and %v profiles", input.ID, nodesCount, profilesCount),
					status.InvalidArgument)
			}

			if err := registry.DelProfile(input.ID); err != nil {
				return err
			}

			if err := registry.Persist(); err != nil {
				return err
			}

			warewulfd.Reload()
			return nil
		}
	})
	u.SetTitle("Delete a profile")
	u.SetDescription("Delete an existing node profile.")
	u.SetTags("Profile")

	return u
}
