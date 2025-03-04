package api

import (
	"context"
	"fmt"

	"dario.cat/mergo"
	"github.com/swaggest/usecase"
	"github.com/swaggest/usecase/status"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd/daemon"
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
		registry, regErr := node.New()
		if regErr != nil {
			return regErr
		}

		registry.NodeProfiles[input.ID] = &input.Profile
		persistErr := registry.Persist()
		if persistErr != nil {
			return persistErr
		}

		*output = *(registry.NodeProfiles[input.ID])

		_ = daemon.DaemonReload()
		return nil
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
		nodeDB, err := node.New()
		if err != nil {
			return fmt.Errorf("failed to initialize nodeDB, err: %w", err)
		}
		profilePtr, err := nodeDB.GetProfilePtr(input.ID)
		if err != nil {
			return fmt.Errorf("failed to retrieve profile by its id, err: %w", err)
		}
		err = mergo.MergeWithOverwrite(profilePtr, &input.Profile)
		if err != nil {
			return err
		}

		err = nodeDB.Persist()
		if err != nil {
			return err
		}

		*output = *profilePtr

		_ = daemon.DaemonReload()
		return nil
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
		registry, regErr := node.New()
		if regErr != nil {
			return regErr
		}

		profile, ok := registry.NodeProfiles[input.ID]
		if !ok {
			return fmt.Errorf("profile '%s' does not exist", input.ID)
		}

		nodesCount := 0
		for _, n := range registry.Nodes {
			if util.InSlice(n.Profiles, input.ID) {
				nodesCount++
			}
		}

		profilesCount := 0
		for _, p := range registry.NodeProfiles {
			if util.InSlice(p.Profiles, input.ID) {
				profilesCount++
			}
		}

		if nodesCount > 0 || profilesCount > 0 {
			return fmt.Errorf("profile '%s' is in use by %v nodes and %v profiles", input.ID, nodesCount, profilesCount)
		}

		delErr := registry.DelProfile(input.ID)
		if delErr != nil {
			return delErr
		}

		if err := registry.Persist(); err != nil {
			return err
		}

		*output = *profile

		_ = daemon.DaemonReload()
		return nil
	})
	u.SetTitle("Delete a profile")
	u.SetDescription("Delete an existing node profile.")
	u.SetTags("Profile")

	return u
}
