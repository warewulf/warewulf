package api

import (
	"context"
	"fmt"
	"runtime"
	"sort"

	"dario.cat/mergo"
	"github.com/swaggest/usecase"
	"github.com/swaggest/usecase/status"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/warewulfd/daemon"
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
		ID   string    `path:"id" required:"true" description:"ID of node to be added"`
		Node node.Node `json:"node" required:"true" description:"Field values in JSON format for added node"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input addNodeInput, output *node.Node) error {
		wwlog.Debug("api.addNode(ID:%v, Node:%+v)", input.ID, input.Node)
		registry, regErr := node.New()
		if regErr != nil {
			return regErr
		}

		registry.Nodes[input.ID] = &input.Node
		persistErr := registry.Persist()
		if persistErr != nil {
			return persistErr
		}

		*output = *(registry.Nodes[input.ID])

		_ = daemon.DaemonReload()
		return nil
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
		registry, regErr := node.New()
		if regErr != nil {
			return regErr
		}

		node, ok := registry.Nodes[input.ID]
		if !ok {
			return fmt.Errorf("node '%s' does not exist", input.ID)
		}

		delErr := registry.DelNode(input.ID)
		if delErr != nil {
			return delErr
		}

		if err := registry.Persist(); err != nil {
			return err
		}

		*output = *node

		_ = daemon.DaemonReload()
		return nil
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
		nodeDB, err := node.New()
		if err != nil {
			return fmt.Errorf("failed to initialize nodeDB, err: %w", err)
		}
		nodePtr, err := nodeDB.GetNodeOnlyPtr(input.ID)
		if err != nil {
			return fmt.Errorf("failed to retrieve node by its id, err: %w", err)
		}
		err = mergo.MergeWithOverwrite(nodePtr, &input.Node)
		if err != nil {
			return err
		}

		err = nodeDB.Persist()
		if err != nil {
			return err
		}

		*output = *nodePtr

		_ = daemon.DaemonReload()
		return nil
	})
	u.SetTitle("Update a node")
	u.SetDescription("Update an existing node.")
	u.SetTags("Node")

	return u
}

func buildAllOverlays() usecase.Interactor {
	u := usecase.NewInteractor(func(ctx context.Context, _ struct{}, output *[]string) error {
		wwlog.Debug("api.buildAllOverlays()")
		nodeDB, err := node.New()
		if err != nil {
			return fmt.Errorf("could not open node configuration: %w", err)
		}

		allNodes, err := nodeDB.FindAllNodes()
		if err != nil {
			return fmt.Errorf("could not get node list: %w", err)
		}

		ret := make([]string, len(allNodes))
		for i := range allNodes {
			ret[i] = allNodes[i].Id()
		}
		sort.Strings(ret)

		err = overlay.BuildAllOverlays(allNodes, allNodes, runtime.NumCPU())
		if err != nil {
			return err
		}

		*output = ret
		return nil
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
		nodeDB, err := node.New()
		if err != nil {
			return fmt.Errorf("could not open node configuration: %w", err)
		}

		allNodes, err := nodeDB.FindAllNodes()
		if err != nil {
			return fmt.Errorf("could not get node list: %w", err)
		}

		targetNode, err := nodeDB.GetNode(input.ID)
		if err != nil {
			return fmt.Errorf("failed to get node with id: %s", input.ID)
		}

		err = overlay.BuildAllOverlays([]node.Node{targetNode}, allNodes, runtime.NumCPU())
		if err != nil {
			return err
		}

		*output = input.ID
		return nil
	})
	u.SetTitle("Build overlay images for a node")
	u.SetDescription("Build system and runtime overlay images for a node.")
	u.SetTags("Node")

	return u
}
