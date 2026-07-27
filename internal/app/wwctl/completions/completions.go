package completions

import (
	"strings"

	"github.com/spf13/cobra"

	"github.com/warewulf/warewulf/internal/pkg/hostlist"
	"github.com/warewulf/warewulf/internal/pkg/image"
	"github.com/warewulf/warewulf/internal/pkg/kernel"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/overlay"
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
	if sources, err := image.ListSources(); err == nil {
		return sources, cobra.ShellCompDirectiveNoFileComp
	}
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func Nodes(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	registry, err := node.New()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	candidates := registry.ListAllNodes()
	hasAll := false
	for _, name := range registry.ListAllGroups() {
		if name == "all" {
			hasAll = true
		}
		candidates = append(candidates, "@"+name)
	}
	if !hasAll {
		candidates = append(candidates, "@all")
	}
	return candidates, cobra.ShellCompDirectiveNoFileComp
}

// Groups completes group names without an "@" prefix. Use this for commands
// that take literal group names (e.g., `wwctl group list`). The reserved
// "all" group is included so it shows up even when nothing is declared yet.
func Groups(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	registry, err := node.New()
	if err != nil {
		return nil, cobra.ShellCompDirectiveNoFileComp
	}
	candidates := registry.ListAllGroups()
	hasAll := false
	for _, n := range candidates {
		if n == "all" {
			hasAll = true
			break
		}
	}
	if !hasAll {
		candidates = append(candidates, "all")
	}
	return candidates, cobra.ShellCompDirectiveNoFileComp
}

func Profiles(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if registry, err := node.New(); err == nil {
		return registry.ListAllProfiles(), cobra.ShellCompDirectiveNoFileComp
	}
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func Overlays(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	list := overlay.FindOverlays()
	return list, cobra.ShellCompDirectiveNoFileComp
}

func OverlayList(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	list, directive := Overlays(cmd, args, toComplete)
	lastCommaIndex := strings.LastIndex(toComplete, ",")
	if lastCommaIndex >= 0 {
		for i := range list {
			list[i] = toComplete[:lastCommaIndex+1] + list[i]
		}
	}
	return list, directive
}

func OverlayFiles(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	myOverlay, _ := overlay.Get(args[0])
	ret, _ := myOverlay.GetFiles()
	return ret, cobra.ShellCompDirectiveNoFileComp
}

func OverlayAndFiles(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	if len(args) == 0 {
		return Overlays(cmd, args, toComplete)
	} else {
		return OverlayFiles(cmd, args, toComplete)
	}
}

func None(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveNoFileComp
}

func LocalFiles(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	return nil, cobra.ShellCompDirectiveDefault
}
