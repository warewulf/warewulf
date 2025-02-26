package build

import (
	"os"
	"runtime"
	"strconv"

	"github.com/spf13/cobra"
	"github.com/warewulf/warewulf/internal/app/wwctl/completions"
)

var (
	baseCmd = &cobra.Command{
		DisableFlagsInUseLine: true,
		Use:                   "build [OPTIONS] NODENAME...",
		Short:                 "(Re)build node overlays",
		Long:                  "This command builds overlays for given nodes.",
		RunE:                  CobraRunE,
		ValidArgsFunction:     completions.Nodes,
	}
	OverlayNames []string
	OverlayDir   string
	Workers      int
)

func init() {
	baseCmd.PersistentFlags().StringSliceVarP(&OverlayNames, "overlay", "O", []string{}, "Build only specific overlay(s)")

	if err := baseCmd.RegisterFlagCompletionFunc("overlay", completions.Overlays); err != nil {
		panic(err)
	}
	baseCmd.PersistentFlags().StringVarP(&OverlayDir, "output", "o", "", `Do not create an overlay image for distribution but write to
	the given directory. An overlay must also be ge given to use this option.`)
	workers := runtime.NumCPU()
	numCPU := os.Getenv("WWWORKER")
	wwWorker, err := strconv.Atoi(numCPU)
	if err == nil {
		workers = wwWorker
	}
	baseCmd.PersistentFlags().IntVar(&Workers, "workers", workers, "The number of parallel workers building overlays")
}

// GetRootCommand returns the root cobra.Command for the application.
func GetCommand() *cobra.Command {
	return baseCmd
}
