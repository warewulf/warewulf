package main

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/swaggest/rest/web"
	"github.com/swaggest/usecase"
	"github.com/swaggest/usecase/status"

	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/version"
)

func main() {
	node.ConfigFile = "/etc/warewulf/nodes.conf"
	service := web.DefaultService()

	service.OpenAPI.Info.Title = "Warewulf v4 API"
	service.OpenAPI.Info.WithDescription("This service provides an API to a Warewulf v4 server.")
	service.OpenAPI.Info.Version = version.GetVersion()

	service.Get("/nodes", getNodes())
	service.Get("/nodes/{id}", getNodeByID())
	service.Get("/profiles", getProfiles())
	service.Get("/profiles/{id}", getProfileByID())

	log.Println("Starting service")
	if err := http.ListenAndServe("localhost:8080", service); err != nil {
		log.Fatal(err)
	}
}

func getNodes() usecase.Interactor {
	u := usecase.NewInteractor(func(ctx context.Context, _ struct{}, output *map[string]*node.Node) error {
		if registry, err := node.New(); err != nil {
			return err
		} else {
			*output = registry.Nodes
			return nil
		}
	})
	u.SetTags("Node")
	return u
}

func getNodeByID() usecase.Interactor {
	type getNodeByIDInput struct {
		ID string `path:"id"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input getNodeByIDInput, output *node.Node) error {
		if registry, err := node.New(); err != nil {
			return err
		} else {
			if node_, err := registry.GetNodeOnly(input.ID); err != nil {
				return status.Wrap(errors.New(fmt.Sprintf("node not found: %v (%v)", input.ID, err)), status.NotFound)
			} else {
				*output = node_
				return nil
			}
		}
	})
	u.SetTags("Node")
	u.SetExpectedErrors(status.NotFound)
	return u
}

func getProfileByID() usecase.Interactor {
	type getProfileByIDInput struct {
		ID string `path:"id"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input getProfileByIDInput, output *node.Profile) error {
		if registry, err := node.New(); err != nil {
			return err
		} else {
			if profile, err := registry.GetProfile(input.ID); err != nil {
				return status.Wrap(errors.New(fmt.Sprintf("profile not found: %v (%v)", input.ID, err)), status.NotFound)
			} else {
				*output = profile
				return nil
			}
		}
	})
	u.SetTags("Profile")
	u.SetExpectedErrors(status.NotFound)
	return u
}

func getProfiles() usecase.Interactor {
	u := usecase.NewInteractor(func(ctx context.Context, _ struct{}, output *map[string]*node.Profile) error {
		if registry, err := node.New(); err != nil {
			return err
		} else {
			*output = registry.NodeProfiles
			return nil
		}
	})
	u.SetTags("Profile")
	return u
}
