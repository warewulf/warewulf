package api

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path"
	"slices"

	"github.com/swaggest/usecase"
	"github.com/swaggest/usecase/status"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

type OverlayResponse struct {
	Files []string `json:"files"`
	Site  bool     `json:"site"`
}

func NewOverlayResponse(name string) *OverlayResponse {
	o := new(OverlayResponse)
	o.Files = []string{}
	if files, err := overlay.OverlayGetFiles(name); err == nil {
		o.Files = files
	}
	o.Site = overlay.GetOverlay(name).IsSiteOverlay()
	return o
}

func getOverlays() usecase.Interactor {
	u := usecase.NewInteractor(func(ctx context.Context, _ struct{}, output *map[string]*OverlayResponse) error {
		wwlog.Debug("api.getOverlays()")
		m := make(map[string]*OverlayResponse)
		names := overlay.FindOverlays()
		for _, name := range names {
			m[name] = NewOverlayResponse(name)
		}
		*output = m
		return nil
	})
	u.SetTitle("Get overlays")
	u.SetDescription("Get all overlays.")
	u.SetTags("Overlay")
	return u
}

func getOverlayByName() usecase.Interactor {
	type getOverlayByNameInput struct {
		Name string `path:"name" required:"true" description:"Name of overlay to get"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input getOverlayByNameInput, output *OverlayResponse) error {
		wwlog.Debug("api.getOverlayByName(Name:%v)", input.Name)
		if !overlay.GetOverlay(input.Name).Exists() {
			return status.Wrap(fmt.Errorf("overlay not found: %v", input.Name), status.NotFound)
		} else {
			*output = *NewOverlayResponse(input.Name)
			return nil
		}
	})
	u.SetTitle("Get an overlay")
	u.SetDescription("Get an overlay by its name.")
	u.SetTags("Overlay")
	return u
}

type OverlayFile struct {
	Overlay  string `json:"overlay"`
	Path     string `json:"path"`
	Contents string `json:"contents"`
	rendered bool
}

func (of *OverlayFile) FullPath() string {
	return path.Join(overlay.GetOverlay(of.Overlay).Rootfs(), of.Path)
}

func (of *OverlayFile) Exists() bool {
	return overlay.GetOverlay(of.Overlay).Exists() && util.IsFile(of.FullPath())
}

func (of *OverlayFile) readContents() (string, error) {
	f, err := os.ReadFile(of.FullPath())
	return string(f), err
}

func (of *OverlayFile) renderContents(nodeName string) (string, error) {
	if !(path.Ext(of.Path) == ".ww") {
		return "", fmt.Errorf("'%s' does not end with '.ww'", of.Path)
	}

	if of.rendered {
		return "", fmt.Errorf("already rendered")
	}

	registry, regErr := node.New()
	if regErr != nil {
		return "", regErr
	}

	renderNode, nodeErr := registry.GetNode(nodeName)
	if nodeErr != nil {
		return "", nodeErr
	}

	allNodes, allNodesErr := registry.FindAllNodes()
	if allNodesErr != nil {
		return "", allNodesErr
	}

	tstruct, structErr := overlay.InitStruct(of.Overlay, renderNode, allNodes)
	if structErr != nil {
		return "", structErr
	}
	tstruct.BuildSource = of.Path

	buffer, _, _, renderErr := overlay.RenderTemplateFile(of.FullPath(), tstruct)
	if renderErr != nil {
		return "", renderErr
	}

	return buffer.String(), nil
}

func NewOverlayFile(name string, path string, renderNodeName string) (*OverlayFile, error) {
	of := new(OverlayFile)
	of.Overlay = name
	of.Path = path
	if renderNodeName == "" {
		if contents, err := of.readContents(); err != nil {
			return of, err
		} else {
			of.Contents = contents
		}
	} else {
		if contents, err := of.renderContents(renderNodeName); err != nil {
			return of, err
		} else {
			of.Contents = contents
		}
	}
	return of, nil
}

func getOverlayFile() usecase.Interactor {
	type getOverlayFileInput struct {
		Name string `path:"name" required:"true" description:"Name of overlay to get a file from"`
		Path string `query:"path" required:"true" description:"Path to file to get from an overlay"`
		Node string `query:"render" description:"ID of node to render a template for"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input getOverlayFileInput, output *OverlayFile) error {
		wwlog.Debug("api.getOverlayFile(Name:%v, Path:%v, Node:%v)", input.Name, input.Path, input.Node)
		if input.Path == "" {
			return status.Wrap(fmt.Errorf("must specify a path"), status.InvalidArgument)
		}

		if relPath, err := url.QueryUnescape(input.Path); err != nil {
			return fmt.Errorf("failed to decode path: %v: %w", input.Path, err)
		} else {
			if overlayFile, err := NewOverlayFile(input.Name, relPath, input.Node); err != nil {
				return fmt.Errorf("unable to read overlay file %v: %v: %w", input.Name, relPath, err)
			} else {
				*output = *overlayFile
				return nil
			}
		}
	})
	u.SetTitle("Get a file from an overlay")
	u.SetDescription("Get a file from an overlay from the overlay name and file path, optionally rendered for a given node.")
	u.SetTags("Overlay")
	return u
}

func createOverlay() usecase.Interactor {
	type createOverlayInput struct {
		Name string `path:"name" required:"true" description:"Name of overlay to create"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input createOverlayInput, output *OverlayResponse) error {
		wwlog.Debug("api.createOverlay(Name:%v)", input.Name)
		newOverlay := overlay.GetSiteOverlay(input.Name)
		if err := newOverlay.Create(); err != nil {
			return err
		}
		*output = *NewOverlayResponse(newOverlay.Name())
		return nil
	})
	u.SetTitle("Create an overlay")
	u.SetDescription("Create an overlay.")
	u.SetTags("Overlay")
	return u
}

func deleteOverlay() usecase.Interactor {
	type deleteOverlayInput struct {
		Name  string `path:"name" required:"true" description:"Name of overlay to delete"`
		Force bool   `query:"force" default:"false" description:"Whether to delete a non-empty overlay, default:'false'"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input deleteOverlayInput, output *OverlayResponse) error {
		wwlog.Debug("api.deleteOverlay(Name:%v, Force:%v)", input.Name, input.Force)
		if registry, err := node.New(); err != nil {
			return err
		} else {
			nodesCount := len(registry.ListNodesUsingOverlay(input.Name))
			profilesCount := len(registry.ListProfilesUsingOverlay(input.Name))
			if nodesCount > 0 || profilesCount > 0 {
				return status.Wrap(fmt.Errorf(
					"overlay '%s' is in use by %v nodes and %v profiles", input.Name, nodesCount, profilesCount),
					status.InvalidArgument)
			}
		}
		*output = *NewOverlayResponse(input.Name)
		overlay_ := overlay.GetSiteOverlay(input.Name)
		if input.Force {
			if err := os.RemoveAll(overlay_.Path()); err != nil {
				return err
			}
		} else {
			if err := os.Remove(overlay_.Path()); err != nil {
				return err
			}
		}
		return nil
	})
	u.SetTitle("Delete an overlay")
	u.SetDescription("Delete an overlay.")
	u.SetTags("Overlay")
	return u
}

func buildOverlay() usecase.Interactor {
	type buildOverlayInput struct {
		Name string `path:"name" required:"true" description:"Name of overlay to create"`
	}
	u := usecase.NewInteractor(func(ctx context.Context, input *buildOverlayInput, output *OverlayResponse) error {
		wwlog.Debug("api.buildSpecificOverlay()")
		if registry, err := node.New(); err != nil {
			return err
		} else {
			nodes, err := registry.FindAllNodes()
			if err != nil {
				return err
			}

			for _, node := range nodes {
				if slices.Contains(node.RuntimeOverlay, input.Name) || slices.Contains(node.SystemOverlay, input.Name) {
					if err := overlay.BuildOverlay(node, nodes, "", []string{input.Name}); err != nil {
						return err
					}
				}
			}

			*output = *NewOverlayResponse(input.Name)
			return nil
		}
	})
	u.SetTitle("Build specific overlay")
	u.SetDescription("Build specific overlay.")
	u.SetTags("Overlay")
	return u
}
