package api

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"sort"
	"time"

	"dario.cat/mergo"
	"github.com/swaggest/usecase"
	"github.com/swaggest/usecase/status"
	"github.com/warewulf/warewulf/internal/pkg/image"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func getNodes() usecase.Interactor {
	u := usecase.NewInteractor(func(ctx context.Context, _ struct{}, output *map[string]*node.Node) error {
		wwlog.Debug("api.getNodes()")
		if registry, err := node.New(); err != nil {
			return err
		} else {
			nodeMap := make(map[string]*node.Node)
			if nodeList, err := registry.FindAllNodes(); err != nil {
				return err
			} else {
				for i := range nodeList {
					nodeMap[nodeList[i].Id()] = &nodeList[i]
				}
				*output = nodeMap
				return nil
			}
		}
	})
	u.SetTitle("Get nodes")
	u.SetDescription("Get all nodes, including field values from associated profiles.")
	u.SetTags("Node")
	return u
}

func getNodeOverlayInfo() usecase.Interactor {
	type getOverlaysInput struct {
		ID string `path:"id" description:"ID of node to retrieve overlays for"`
	}
	type overlayInfo struct {
		Overlays []string   `json:"overlays,omitempty" yaml:"overlays,omitempty"`
		MTime    *time.Time `json:"mtime,omitempty" yaml:"mtime,omitempty"`
	}
	type getOverlaysOutput struct {
		SystemOverlay  *overlayInfo `json:"system overlay,omitempty" yaml:"system overlay,omitempty"`
		RuntimeOverlay *overlayInfo `json:"runtime overlay,omitempty" yaml:"runtime overlay,omitempty"`
	}
	u := usecase.NewInteractor(func(ctx context.Context, input *getOverlaysInput, output *getOverlaysOutput) error {
		wwlog.Debug("api.getNodeOverlayInfo()")
		if registry, err := node.New(); err != nil {
			return err
		} else {
			if node_, err := registry.GetNode(input.ID); err != nil {
				return status.Wrap(err, status.NotFound)
			} else {
				out := getOverlaysOutput{
					SystemOverlay: &overlayInfo{
						Overlays: node_.SystemOverlay,
					},
					RuntimeOverlay: &overlayInfo{
						Overlays: node_.RuntimeOverlay,
					},
				}
				sysImagePath := overlay.OverlayImage(input.ID, "system", node_.SystemOverlay)
				if sysImageStat, err := os.Stat(sysImagePath); err == nil {
					mtime := sysImageStat.ModTime()
					out.SystemOverlay.MTime = &mtime
				}

				runtimeImagePath := overlay.OverlayImage(input.ID, "runtime", node_.RuntimeOverlay)
				if runtimeImageStat, err := os.Stat(runtimeImagePath); err == nil {
					mtime := runtimeImageStat.ModTime()
					out.RuntimeOverlay.MTime = &mtime
				}
				*output = out
				return nil
			}
		}
	})
	u.SetTitle("Get overlay info for a node")
	u.SetDescription("Get system and runtime overlay info for a node.")
	u.SetTags("Node")

	return u
}

func getNodeByID() usecase.Interactor {
	type getNodeByIDInput struct {
		ID string `path:"id" required:"true" description:"ID of node to get"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input getNodeByIDInput, output *node.Node) error {
		wwlog.Debug("api.getNodeByID(ID:%v)", input.ID)
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
	u.SetDescription("Get a node by its ID, including field values from associated profiles.")
	u.SetTags("Node")
	u.SetExpectedErrors(status.NotFound)
	return u
}

func getRawNodeByID() usecase.Interactor {
	type getNodeByIDInput struct {
		ID string `path:"id" required:"true" description:"ID of node to get"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input getNodeByIDInput, output *node.Node) error {
		wwlog.Debug("api.getRawNodeByID(ID:%v)", input.ID)
		if registry, err := node.New(); err != nil {
			return err
		} else {
			if node_, ok := registry.Nodes[input.ID]; !ok {
				return status.Wrap(fmt.Errorf("node not found: %v", input.ID), status.NotFound)
			} else {
				*output = *node_
				return nil
			}
		}
	})
	u.SetTitle("Get a raw node")
	u.SetDescription("Get a node by its ID, without field values from associated profiles.")
	u.SetTags("Node")
	u.SetExpectedErrors(status.NotFound)
	return u
}

func getNodeFields() usecase.Interactor {
	type getNodeByIDInput struct {
		ID string `path:"id" required:"true" description:"ID of node from which to retrieve fields"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input getNodeByIDInput, output *[]node.Field) error {
		wwlog.Debug("api.getNodeFields(ID:%v)", input.ID)
		if registry, err := node.New(); err != nil {
			return err
		} else {
			if n, fields, err := registry.MergeNode(input.ID); err != nil {
				return status.Wrap(fmt.Errorf("node not found: %v (%v)", input.ID, err), status.NotFound)
			} else {
				*output = fields.List(n)
				return nil
			}
		}
	})
	u.SetTitle("Get node fields")
	u.SetDescription("Get the fields and values of a node, indicating which profiles each field originates from.")
	u.SetTags("Node")
	u.SetExpectedErrors(status.NotFound)
	return u
}

func addNode() usecase.Interactor {
	type addNodeInput struct {
		ID          string    `path:"id" required:"true" description:"ID of node to be added"`
		Node        node.Node `json:"node" required:"true" description:"Field values in JSON format for added node"`
		IfNoneMatch string    `header:"If-None-Match" description:"Set to '*' to indicate that the node should only be created if it does not already exist"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input addNodeInput, output *node.Node) error {
		wwlog.Debug("api.addNode(ID:%v, Node:%+v)", input.ID, input.Node)
		if registry, err := node.New(); err != nil {
			return err
		} else {
			if input.IfNoneMatch == "*" {
				if _, ok := registry.Nodes[input.ID]; ok {
					return status.Wrap(fmt.Errorf("node '%s' already exists", input.ID), status.InvalidArgument)
				}
			}
			for _, profile := range input.Node.Profiles {
				if _, ok := registry.NodeProfiles[profile]; !ok {
					return status.Wrap(fmt.Errorf("profile '%s' does not exist", profile), status.InvalidArgument)
				}
			}
			if input.Node.ImageName != "" && !image.ValidSource(input.Node.ImageName) {
				return status.Wrap(fmt.Errorf("image '%s' does not exist", input.Node.ImageName), status.InvalidArgument)
			}
			for _, overlay_ := range input.Node.SystemOverlay {
				if !overlay.GetOverlay(overlay_).Exists() {
					return status.Wrap(fmt.Errorf("overlay '%s' does not exist", overlay_), status.InvalidArgument)
				}
			}
			for _, overlay_ := range input.Node.RuntimeOverlay {
				if !overlay.GetOverlay(overlay_).Exists() {
					return status.Wrap(fmt.Errorf("overlay '%s' does not exist", overlay_), status.InvalidArgument)
				}
			}
			registry.Nodes[input.ID] = &input.Node
			if err := registry.Persist(); err != nil {
				return err
			}
			warewulfd.Reload()
			*output = *(registry.Nodes[input.ID])
			return nil
		}
	})
	u.SetTitle("Add a node")
	u.SetDescription("Add a new node.")
	u.SetTags("Node")

	return u
}

func deleteNode() usecase.Interactor {
	type deleteNodeInput struct {
		ID string `path:"id" required:"true" description:"ID of node to delete"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input deleteNodeInput, output *node.Node) error {
		wwlog.Debug("api.deleteNode(ID:%v)", input.ID)
		if registry, err := node.New(); err != nil {
			return err
		} else {
			if node, ok := registry.Nodes[input.ID]; ok {
				*output = *node
			}
			if err := registry.DelNode(input.ID); err != nil {
				return err
			}
			if err := registry.Persist(); err != nil {
				return err
			}
			warewulfd.Reload()
			return nil
		}
	})
	u.SetTitle("Delete a node")
	u.SetDescription("Delete an existing node.")
	u.SetTags("Node")

	return u
}

func updateNode() usecase.Interactor {
	type updateNodeInput struct {
		ID   string    `path:"id" description:"ID of node to update"`
		Node node.Node `json:"node" required:"true" description:"Field values in JSON format to update on node"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input updateNodeInput, output *node.Node) error {
		wwlog.Debug("api.updateNode(ID:%v, Node:%+v)", input.ID, input.Node)
		if registry, err := node.New(); err != nil {
			return err
		} else {
			for _, profile := range input.Node.Profiles {
				if _, ok := registry.NodeProfiles[profile]; !ok {
					return status.Wrap(fmt.Errorf("profile '%s' does not exist", profile), status.InvalidArgument)
				}
			}
			if input.Node.ImageName != "" && !image.ValidSource(input.Node.ImageName) {
				return status.Wrap(fmt.Errorf("image '%s' does not exist", input.Node.ImageName), status.InvalidArgument)
			}
			for _, overlay_ := range input.Node.SystemOverlay {
				if !overlay.GetOverlay(overlay_).Exists() {
					return status.Wrap(fmt.Errorf("overlay '%s' does not exist", overlay_), status.InvalidArgument)
				}
			}
			for _, overlay_ := range input.Node.RuntimeOverlay {
				if !overlay.GetOverlay(overlay_).Exists() {
					return status.Wrap(fmt.Errorf("overlay '%s' does not exist", overlay_), status.InvalidArgument)
				}
			}
			if nodePtr, err := registry.GetNodeOnlyPtr(input.ID); err != nil {
				return status.Wrap(err, status.NotFound)
			} else {
				if err := mergo.MergeWithOverwrite(nodePtr, &input.Node); err != nil {
					return err
				}
				if err := registry.Persist(); err != nil {
					return err
				}
				warewulfd.Reload()
				*output = *nodePtr
				return nil
			}
		}
	})
	u.SetTitle("Update a node")
	u.SetDescription("Update an existing node.")
	u.SetTags("Node")

	return u
}

func buildAllOverlays() usecase.Interactor {
	u := usecase.NewInteractor(func(ctx context.Context, _ struct{}, output *[]string) error {
		wwlog.Debug("api.buildAllOverlays()")
		if registry, err := node.New(); err != nil {
			return err
		} else {
			if nodes, err := registry.FindAllNodes(); err != nil {
				return fmt.Errorf("could not get node list: %w", err)
			} else {
				ret := make([]string, len(nodes))
				for i := range nodes {
					ret[i] = nodes[i].Id()
				}
				sort.Strings(ret)
				if err := overlay.BuildAllOverlays(nodes, nodes, runtime.NumCPU()); err != nil {
					return err
				}
				*output = ret
				return nil
			}
		}
	})
	u.SetTitle("Build all overlay images")
	u.SetDescription("Build system and runtime overlay images for all nodes.")
	u.SetTags("Node")

	return u
}

func buildOverlays() usecase.Interactor {
	type buildOverlayInput struct {
		ID string `path:"id" description:"ID of node to build overlay images for"`
	}
	u := usecase.NewInteractor(func(ctx context.Context, input *buildOverlayInput, output *string) error {
		wwlog.Debug("api.buildOverlays()")
		if registry, err := node.New(); err != nil {
			return err
		} else {
			nodes, err := registry.FindAllNodes()
			if err != nil {
				return err
			}

			if node_, err := registry.GetNode(input.ID); err != nil {
				return status.Wrap(err, status.NotFound)
			} else {
				if err := overlay.BuildAllOverlays([]node.Node{node_}, nodes, runtime.NumCPU()); err != nil {
					return err
				}
				*output = input.ID
				return nil
			}
		}
	})
	u.SetTitle("Build overlay images for a node")
	u.SetDescription("Build system and runtime overlay images for a node.")
	u.SetTags("Node")

	return u
}
