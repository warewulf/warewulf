package api

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/swaggest/usecase"
	image_api "github.com/warewulf/warewulf/internal/pkg/api/image"
	"github.com/warewulf/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/warewulf/warewulf/internal/pkg/image"
	"github.com/warewulf/warewulf/internal/pkg/kernel"
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
	u.SetTitle("Get all images")
	u.SetDescription("Get all defined images")
	u.SetTags("Image")
	return u
}

func getImageByName() usecase.Interactor {
	type getImageByNameInput struct {
		Name string `path:"name"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input getImageByNameInput, output *Image) error {
		if !image.ValidSource(input.Name) {
			return fmt.Errorf("image not found: %v", input.Name)
		} else {
			*output = *NewImage(input.Name)
			return nil
		}
	})
	u.SetTitle("Get an image")
	u.SetDescription("Get an image by its name")
	u.SetTags("Image")
	return u
}

func importImage() usecase.Interactor {
	type importImageInput struct {
		Name     string `path:"name"`
		URI      string `json:"uri" required:"true" description:"OCI registry URI to import image definition from"`
		NoHttps  bool   `json:"nohttps" default:"false" description:"Use http, rather than https, to communicate with the registry, default:'false'"`
		User     string `json:"user" description:"Username for the registry, if needed"`
		Password string `json:"password" description:"Password for the registry, if needed"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input importImageInput, output *Image) error {
		if !strings.HasPrefix(input.URI, "docker://") {
			return errors.New("uri only supports docker:// prefix for now")
		}

		if !image.ValidName(input.Name) {
			return fmt.Errorf("VNFS name contains illegal characters: %s", input.Name)
		}

		sctx, err := image_api.GetSystemContext(input.NoHttps, input.User, input.Password, "")
		if err != nil {
			return err
		}

		err = image.ImportDocker(input.URI, input.Name, sctx)
		if err != nil {
			return err
		}

		*output = *NewImage(input.Name)
		return nil
	})
	u.SetTitle("Import an image")
	u.SetDescription("Import an image from an OCI registry")
	u.SetTags("Image")

	return u
}

func deleteImage() usecase.Interactor {
	type deleteImageInput struct {
		Name string `path:"name"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input deleteImageInput, output *Image) error {
		if !image.ValidSource(input.Name) {
			return fmt.Errorf("image not found: %v", input.Name)
		}

		*output = *NewImage(input.Name)
		cdp := &wwapiv1.ImageDeleteParameter{
			ImageNames: []string{input.Name},
		}

		err := image_api.ImageDelete(cdp)
		return err
	})
	u.SetTitle("Delete an image")
	u.SetDescription("Delete an existing image")
	u.SetTags("Image")

	return u
}

func updateImage() usecase.Interactor {
	type renameImageInput struct {
		Name    string `path:"name"`
		NewName string `json:"name"`
		Build   bool   `query:"build" default:"true" description:"Build the image image after renaming, default:'true'"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input renameImageInput, output *Image) error {
		name := input.Name
		if input.NewName != "" {
			crp := &wwapiv1.ImageRenameParameter{
				ImageName:  input.Name,
				TargetName: input.NewName,
				Build:      input.Build,
			}

			err := image_api.ImageRename(crp)
			if err != nil {
				return err
			}
			name = input.NewName
		}

		*output = *NewImage(name)
		return nil
	})
	u.SetTitle("Update or rename an image")
	u.SetDescription("Update or rename an existing image")
	u.SetTags("Image")

	return u
}

func buildImage() usecase.Interactor {
	type buildImageInput struct {
		Name  string `path:"name"`
		Force bool   `query:"force" default:"false" description:"Build the image image even if it appears unnecessary, default:'false'"`
	}

	u := usecase.NewInteractor(func(ctx context.Context, input buildImageInput, output *Image) error {
		cbp := &wwapiv1.ImageBuildParameter{
			ImageNames: []string{input.Name},
			Force:      input.Force,
		}

		err := image_api.ImageBuild(cbp)
		if err != nil {
			return err
		}

		*output = *NewImage(input.Name)
		return nil
	})
	u.SetTitle("Build an image")
	u.SetDescription("Build an image image")
	u.SetTags("Image")

	return u
}
