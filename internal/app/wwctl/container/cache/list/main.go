package cachelist

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/containers/image/v5/oci/layout"
	"github.com/containers/image/v5/types"
	imgSpecs "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/opencontainers/umoci"
	"github.com/spf13/cobra"
	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(vars *variables) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		eng, err := umoci.OpenLayout(warewulfconf.Get().Warewulf.DataStore + "/oci")
		ctx := context.Background()
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()
		if !vars.allblobs {
			refs, err := eng.ListReferences(ctx)
			if err != nil {
				return err
			}
			for _, ref := range refs {
				wwlog.Info("reference: %v\n", ref)
				refImg, err := layout.ParseReference(warewulfconf.Get().Warewulf.DataStore + "/oci:" + ref)
				if err != nil {
					return err
				}
				if vars.showManifest {
					img, err := refImg.NewImageSource(ctx, &types.SystemContext{})
					if err != nil {
						return err
					}
					manifestTmp, _, err := img.GetManifest(ctx, nil)
					if err != nil {
						return err
					}
					var manifest imgSpecs.Manifest
					if err := json.Unmarshal(manifestTmp, &manifest); err != nil {
						return fmt.Errorf("unable to unmarshall manifest json: %v", err)
					}
					tmpOut, _ := json.MarshalIndent(manifest, "", "  ")
					wwlog.Info("manifest: %s", tmpOut)
				}
				/*
					if vars.showDigest {
						img, err := refImg.NewImageSource(ctx, &types.SystemContext{})
						if err != nil {
							return err
						}

						ref.
					}
				*/
			}
		} else {
			blobs, err := eng.ListBlobs(ctx)
			if err != nil {
				return err
			}
			for _, blob := range blobs {
				wwlog.Info(string(blob))
			}
		}

		return
	}
}
