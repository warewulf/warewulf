package build

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/pkg/image"
)

func CobraRunE(cmd *cobra.Command, imageNames []string) error {
	if BuildAll {
		var err error
		imageNames, err = image.ListSources()
		if err != nil {
			return fmt.Errorf("could not list images: %w", err)
		}
	}

	if len(imageNames) == 0 {
		return fmt.Errorf("no images specified; use --all to build all images")
	}

	if SyncUser {
		for _, name := range imageNames {
			if err := image.Syncuser(name, true); err != nil {
				return fmt.Errorf("syncuser error: %w", err)
			}
		}
	}

	for _, imageName := range imageNames {
		if err := image.Build(imageName, BuildForce); err != nil {
			return fmt.Errorf("error building image %s: %s", imageName, err)
		}
	}

	return nil
}
