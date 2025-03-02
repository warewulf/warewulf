package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/swaggest/usecase"
	"github.com/swaggest/usecase/status"
	image_api "github.com/warewulf/warewulf/internal/pkg/api/image"
	"github.com/warewulf/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/warewulf/warewulf/internal/pkg/image"
	"github.com/warewulf/warewulf/internal/pkg/kernel"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

type Image struct {
	Kernels   []string `json:"kernels"`
	Size      int      `json:"size"`
	BuildTime int64    `json:"buildtime"`
	Writable  bool     `json:"writable"`
}

func NewImage(name string) *Image {
	c := new(Image)
	c.Kernels = []string{}
	for _, k := range kernel.FindKernels(name) {
		c.Kernels = append(c.Kernels, k.Path)
	}
	c.Size = image.ImageSize(name)
	modTime := image.ImageModTime(name)
	if modTime.IsZero() {
		c.BuildTime = 0
	} else {
		c.BuildTime = modTime.Unix()
	}
	c.Writable = image.IsWriteAble(name)
	return c
}

func getImages() usecase.Interactor {
	u := usecase.NewInteractor(func(ctx context.Context, _ struct{}, output *map[string]*Image) error {
		wwlog.Debug("api.getImages()")
		m := make(map[string]*Image)
		if names, err := image.ListSources(); err != nil {
			return err
		} else {
			for _, name := range names {
				m[name] = NewImage(name)
			}
			*output = m
			return nil
		}
	})
	u.SetTitle("Get images")
	u.SetDescription("Get all node images")
	u.SetTags("Image")
	return u
}

func getImageByName() usecase.Interactor {
	type getImageByNameInput struct {
		Name string `path:"name" required:"true" description:"Name of image to add"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input getImageByNameInput, output *Image) error {
		wwlog.Debug("api.getImageByName(Name:%v)", input.Name)
		if !image.ValidSource(input.Name) {
			return status.Wrap(fmt.Errorf("image not found: %v", input.Name), status.NotFound)
		} else {
			*output = *NewImage(input.Name)
			return nil
		}
	})
	u.SetTitle("Get an image")
	u.SetDescription("Get a node image by its name")
	u.SetTags("Image")
	return u
}

func importImage() usecase.Interactor {
	type importImageInput struct {
		Name     string `path:"name" required:"true" description:"Name of image to import"`
		URI      string `json:"uri" required:"true" description:"OCI registry URI to import image definition from"`
		NoHttps  bool   `json:"nohttps" default:"false" description:"Use http, rather than https, to communicate with the registry, default:'false'"`
		User     string `json:"user" description:"Username for the registry, if needed"`
		Password string `json:"password" description:"Password for the registry, if needed"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input importImageInput, output *Image) error {
		wwlog.Debug("api.importImage(Name:%v, URI:%v, NoHttps:%v, User:%v, Password:[redacted])",
			input.Name, input.URI, input.NoHttps, input.User)
		if !strings.HasPrefix(input.URI, "docker://") {
			return status.Wrap(fmt.Errorf("missing docker:// prefix: %s", input.URI), status.InvalidArgument)
		}

		if !image.ValidName(input.Name) {
			return status.Wrap(fmt.Errorf("name contains illegal characters: %s", input.Name), status.InvalidArgument)
		}

		if sctx, err := image_api.GetSystemContext(input.NoHttps, input.User, input.Password, ""); err != nil {
			return err
		} else {
			if err := image.ImportDocker(input.URI, input.Name, sctx); err != nil {
				return err
			}
			*output = *NewImage(input.Name)
			return nil
		}
	})
	u.SetTitle("Import an image")
	u.SetDescription("Import a node image from an OCI registry")
	u.SetTags("Image")

	return u
}

func deleteImage() usecase.Interactor {
	type deleteImageInput struct {
		Name string `path:"name" required:"true" description:"Name of image to delete"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input deleteImageInput, output *Image) error {
		wwlog.Debug("api.deleteImage(Name:%v)", input.Name)
		if image.ValidSource(input.Name) {
			*output = *NewImage(input.Name)
		}

		if registry, err := node.New(); err != nil {
			return err
		} else {
			nodesCount := len(registry.ListNodesUsingImage(input.Name))
			profilesCount := len(registry.ListProfilesUsingImage(input.Name))
			if nodesCount > 0 || profilesCount > 0 {
				return status.Wrap(fmt.Errorf(
					"image '%s' is in use by %v nodes and %v profiles", input.Name, nodesCount, profilesCount),
					status.InvalidArgument)
			}
		}

		cdp := &wwapiv1.ImageDeleteParameter{
			ImageNames: []string{input.Name},
		}

		return image_api.ImageDelete(cdp)
	})
	u.SetTitle("Delete an image")
	u.SetDescription("Delete an existing node image")
	u.SetTags("Image")

	return u
}

func updateImage() usecase.Interactor {
	type renameImageInput struct {
		Name    string `path:"name" required:"true" description:"Name of image to update"`
		NewName string `json:"name" description:"New name to rename the image to"`
		Build   bool   `query:"build" default:"true" description:"Build the image image after renaming, default:'true'"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input renameImageInput, output *Image) error {
		wwlog.Debug("api.updateImage(Name:%v, NewName:%v, Build:%v)", input.Name, input.NewName, input.Build)
		name := input.Name
		if input.NewName != "" {
			crp := &wwapiv1.ImageRenameParameter{
				ImageName:  input.Name,
				TargetName: input.NewName,
				Build:      input.Build,
			}

			if err := image_api.ImageRename(crp); err != nil {
				return err
			}
			name = input.NewName
		}

		*output = *NewImage(name)
		return nil
	})
	u.SetTitle("Update or rename an image")
	u.SetDescription("Update or rename an existing node image")
	u.SetTags("Image")

	return u
}

func buildImage() usecase.Interactor {
	type buildImageInput struct {
		Name  string `path:"name" required:"true" description:"Name of image to build"`
		Force bool   `query:"force" default:"false" description:"Build the image image even if it appears unnecessary, default:'false'"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input buildImageInput, output *Image) error {
		wwlog.Debug("api.buildImage(Name:%v, Force:%v)", input.Name, input.Force)
		cbp := &wwapiv1.ImageBuildParameter{
			ImageNames: []string{input.Name},
			Force:      input.Force,
		}

		if err := image_api.ImageBuild(cbp); err != nil {
			return err
		}

		*output = *NewImage(input.Name)
		return nil
	})
	u.SetTitle("Build an image")
	u.SetDescription("Build a node image")
	u.SetTags("Image")

	return u
}
