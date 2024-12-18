package completions

import (
	"github.com/spf13/cobra"

	"github.com/warewulf/warewulf/internal/pkg/container"
	"github.com/warewulf/warewulf/internal/pkg/hostlist"
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
		} else if node_.ContainerName != "" {
			kernels := kernel.FindKernels(node_.ContainerName)
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
		} else if profile.ContainerName != "" {
			kernels := kernel.FindKernels(profile.ContainerName)
			for _, kernel_ := range kernels {
				kernelVersions = append(kernelVersions, kernel_.Version(), kernel_.Path)
			}
		}
	}
	return kernelVersions, cobra.ShellCompDirectiveNoFileComp
}

func Containers(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	sources, _ := container.ListSources()
	return sources, cobra.ShellCompDirectiveNoFileComp
}