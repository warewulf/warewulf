package api

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"path"

	"github.com/swaggest/usecase"
	"github.com/swaggest/usecase/status"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
	"github.com/warewulf/warewulf/internal/pkg/util"
)

type Overlay struct {
	Files []string `json:"files"`
}

func NewOverlay(name string) *Overlay {
	o := new(Overlay)
	o.Files = []string{}
	if files, err := overlay.OverlayGetFiles(name); err == nil {
		o.Files = files
	}
	return o
}

func getOverlays() usecase.Interactor {
	u := usecase.NewInteractor(func(ctx context.Context, _ struct{}, output *map[string]*Overlay) error {
		m := make(map[string]*Overlay)
		names := overlay.FindOverlays()
		for _, name := range names {
			m[name] = NewOverlay(name)
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
		Name string `path:"name"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input getOverlayByNameInput, output *Overlay) error {
		if !overlay.GetOverlay(input.Name).Exists() {
			return fmt.Errorf("overlay not found: %v", input.Name)
		} else {
			*output = *NewOverlay(input.Name)
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
	type getOverlayByNameInput struct {
		Name string `path:"name"`
		Path string `query:"path" required:"true"`
		Node string `query:"render"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input getOverlayByNameInput, output *OverlayFile) error {
		if input.Path == "" {
			return status.Wrap(fmt.Errorf("must specify a path"), status.InvalidArgument)
		}

		relPath, parseErr := url.QueryUnescape(input.Path)
		if parseErr != nil {
			return fmt.Errorf("failed to decode path: %v: %w", input.Path, parseErr)
		}

		overlayFile, err := NewOverlayFile(input.Name, relPath, input.Node)
		if err != nil {
			return fmt.Errorf("unable to read overlay file %v: %v: %w", input.Name, relPath, err)
		}

		*output = *overlayFile
		return nil
	})
	u.SetTitle("Get an overlay file")
	u.SetDescription("Get an overlay file by its name and path, optionally rendered for a given node.")
	u.SetTags("Overlay")
	return u
}

func createOverlay() usecase.Interactor {
	type createOverlayInput struct {
		Name string `path:"name"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input createOverlayInput, output *Overlay) error {
		newOverlay := overlay.GetSiteOverlay(input.Name)
		if err := newOverlay.Create(); err != nil {
			return err
		}
		*output = *NewOverlay(newOverlay.Name())
		return nil
	})
	u.SetTitle("Create an overlay")
	u.SetDescription("Create an overlay.")
	u.SetTags("Overlay")
	return u
}

func deleteOverlay() usecase.Interactor {
	type deleteOverlayInput struct {
		Name  string `path:"name"`
		Force bool   `query:"force" default:"false"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input deleteOverlayInput, output *Overlay) error {
		*output = *NewOverlay(input.Name)
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
