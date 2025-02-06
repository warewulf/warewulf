package completions

import (
	"github.com/spf13/cobra"

	"github.com/warewulf/warewulf/internal/pkg/hostlist"
	"github.com/warewulf/warewulf/internal/pkg/image"
	"github.com/warewulf/warewulf/internal/pkg/kernel"
	"github.com/warewulf/warewulf/internal/pkg/node"
)

func NodeKernelVersion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var kernelVersions []string
	registry, err := node.New()
	if err != nil {
		return kernelVersions, cobra.ShellCompDirectiveNoFileComp
	}
	nodes := hostlist.Expand(args)
	for _, id := range nodes {
		if node_, err := registry.GetNode(id); err != nil {
			continue
		} else if node_.ImageName != "" {
			kernels := kernel.FindKernels(node_.ImageName)
			for _, kernel_ := range kernels {
				kernelVersions = append(kernelVersions, kernel_.Version(), kernel_.Path)
			}
		}
	}
	return kernelVersions, cobra.ShellCompDirectiveNoFileComp
}

func ProfileKernelVersion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	var kernelVersions []string
	registry, err := node.New()
	if err != nil {
		return kernelVersions, cobra.ShellCompDirectiveNoFileComp
	}
	for _, id := range args {
		if profile, err := registry.GetProfile(id); err != nil {
			continue
		} else if profile.ImageName != "" {
			kernels := kernel.FindKernels(profile.ImageName)
			for _, kernel_ := range kernels {
				kernelVersions = append(kernelVersions, kernel_.Version(), kernel_.Path)
			}
		}
	}
	return kernelVersions, cobra.ShellCompDirectiveNoFileComp
}

func Images(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	sources, _ := image.ListSources()
	return sources, cobra.ShellCompDirectiveNoFileComp
}

func Profiles(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) != 0 {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	registry, _ := node.New()
	return registry.ListAllProfiles(), cobra.ShellCompDirectiveNoFileComp
}
