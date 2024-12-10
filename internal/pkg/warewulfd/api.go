package warewulfd

import (
	"context"
	"fmt"

	"github.com/swaggest/openapi-go/openapi3"
	"github.com/swaggest/rest/web"
	swgui "github.com/swaggest/swgui/v5emb"
	"github.com/swaggest/usecase"
	"github.com/swaggest/usecase/status"

	"github.com/warewulf/warewulf/internal/pkg/container"
	"github.com/warewulf/warewulf/internal/pkg/kernel"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/version"
)

func apiHandler() *web.Service {
	api := web.NewService(openapi3.NewReflector())

	api.OpenAPISchema().SetTitle("Warewulf v4 API")
	api.OpenAPISchema().SetDescription("This service provides an API to a Warewulf v4 server.")
	api.OpenAPISchema().SetVersion(version.GetVersion())

	api.Get("/api/raw-nodes", apiGetRawNodes())
	api.Get("/api/raw-nodes/{id}", apiGetRawNodeByID())

	api.Get("/api/nodes", apiGetNodes())
	api.Get("/api/nodes/{id}", apiGetNodeByID())

	api.Get("/api/profiles", apiGetProfiles())
	api.Get("/api/profiles/{id}", apiGetProfileByID())

	api.Get("/api/containers", apiGetContainers())

	api.Docs("/api/docs", swgui.New)

	return api
}

func apiGetRawNodes() usecase.Interactor {
	u := usecase.NewInteractor(func(ctx context.Context, _ struct{}, output *map[string]*node.Node) error {
		if registry, err := node.New(); err != nil {
			return err
		} else {
			*output = registry.Nodes
			return nil
		}
	})
	u.SetTitle("Get raw nodes")
	u.SetDescription("Get all nodes, without merging in values from associated profiles.")
	u.SetTags("Node")
	return u
}

func apiGetRawNodeByID() usecase.Interactor {
	type getNodeByIDInput struct {
		ID string `path:"id"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input getNodeByIDInput, output *node.Node) error {
		if registry, err := node.New(); err != nil {
			return err
		} else {
			if node_, err := registry.GetNodeOnly(input.ID); err != nil {
				return status.Wrap(fmt.Errorf("node not found: %v (%v)", input.ID, err), status.NotFound)
			} else {
				*output = node_
				return nil
			}
		}
	})
	u.SetTitle("Get a raw node")
	u.SetDescription("Get a node by its ID, without merging in values from associated profiles.")
	u.SetTags("Node")
	u.SetExpectedErrors(status.NotFound)
	return u
}

func apiGetNodes() usecase.Interactor {
	u := usecase.NewInteractor(func(ctx context.Context, _ struct{}, output *map[string]*node.Node) error {
		if registry, err := node.New(); err != nil {
			return err
		} else {
			nodeMap := make(map[string]*node.Node)
			if nodeList, err := registry.FindAllNodes(); err != nil {
				return err
			} else {
				for _, n := range nodeList {
					nodeMap[n.Id()] = &n
				}
				*output = nodeMap
				return nil
			}
		}
	})
	u.SetTitle("Get nodes")
	u.SetDescription("Get all nodes, including values from associated profiles.")
	u.SetTags("Node")
	return u
}

func apiGetNodeByID() usecase.Interactor {
	type getNodeByIDInput struct {
		ID string `path:"id"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input getNodeByIDInput, output *node.Node) error {
		if registry, err := node.New(); err != nil {
			return err
		} else {
			if node_, err := registry.GetNode(input.ID); err != nil {
				return status.Wrap(fmt.Errorf("node not found: %v (%v)", input.ID, err), status.NotFound)
			} else {
				*output = node_
				return nil
			}
		}
	})
	u.SetTitle("Get a node")
	u.SetDescription("Get a node by its ID, including values from associated profiles.")
	u.SetTags("Node")
	u.SetExpectedErrors(status.NotFound)
	return u
}

func apiGetProfiles() usecase.Interactor {
	u := usecase.NewInteractor(func(ctx context.Context, _ struct{}, output *map[string]*node.Profile) error {
		if registry, err := node.New(); err != nil {
			return err
		} else {
			*output = registry.NodeProfiles
			return nil
		}
	})
	u.SetTitle("Get node profiles")
	u.SetDescription("Get all node profiles.")
	u.SetTags("Profile")
	return u
}

func apiGetProfileByID() usecase.Interactor {
	type getProfileByIDInput struct {
		ID string `path:"id"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input getProfileByIDInput, output *node.Profile) error {
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
	u.SetTitle("Get a node profile")
	u.SetDescription("Get a node profile by its ID.")
	u.SetTags("Profile")
	u.SetExpectedErrors(status.NotFound)
	return u
}

type Container struct {
	Nodes    []string `json:"nodes"`
	Kernels  []string `json:"kernel,omitempty"`
	Size     int      `json:"size"`
	Writable bool     `json:"writable"`
}

func apiGetContainers() usecase.Interactor {
	u := usecase.NewInteractor(func(ctx context.Context, _ struct{}, output *map[string]*Container) error {
		m := make(map[string]*Container)
		if names, err := container.ListSources(); err != nil {
			return err
		} else {
			for _, name := range names {
				m[name] = new(Container)
				m[name].Size = container.ImageSize(name)
				m[name].Writable = container.IsWriteAble(name)
				for _, k := range kernel.FindKernels(name) {
					m[name].Kernels = append(m[name].Kernels, k.Path)
				}
			}
			*output = m
			return nil
		}
	})
	u.SetTitle("Get container images")
	u.SetDescription("Get all container images.")
	u.SetTags("Container")
	return u
}
