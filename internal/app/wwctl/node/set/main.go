package set

import (
	"fmt"

	"github.com/hpcng/warewulf/internal/pkg/api/node"
	"github.com/hpcng/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/hpcng/warewulf/internal/pkg/api/util"
	"github.com/spf13/cobra"
)

func CobraRunE(cmd *cobra.Command, args []string) (err error) {
	set := wwapiv1.NodeSetParameter{
		Comment:        SetComment,
		Container:      SetContainer,
		KernelOverride: SetKernelOverride,
		KernelArgs:     SetKernelArgs,
		Netname:        SetNetName,
		Netdev:         SetNetDev,
		Ipaddr:         SetIpaddr,
		Netmask:        SetNetmask,
		Gateway:        SetGateway,
		Hwaddr:         SetHwaddr,
		Type:           SetType,
		Onboot:         SetNetOnBoot,
		NetDefault:     SetNetDefault,
		NetdevDelete:   SetNetDevDel,
		Cluster:        SetClusterName,
		Ipxe:           SetIpxe,
		InitOverlay:    SetInitOverlay,
		RuntimeOverlay: SetRuntimeOverlay,
		SystemOverlay:  SetSystemOverlay,
		IpmiIpaddr:     SetIpmiIpaddr,
		IpmiNetmask:    SetIpmiNetmask,
		IpmiPort:       SetIpmiPort,
		IpmiGateway:    SetIpmiGateway,
		IpmiUsername:   SetIpmiUsername,
		IpmiPassword:   SetIpmiPassword,
		IpmiInterface:  SetIpmiInterface,
		AllNodes:       SetNodeAll,
		Profile:        SetProfile,
		ProfileAdd:     SetAddProfile,
		ProfileDelete:  SetDelProfile,
		Force:          SetForce,
		Init:           SetInit,
		Discoverable:   SetDiscoverable,
		Undiscoverable: SetUndiscoverable,
		Root:           SetRoot,
		Tags:           SetTags,
		TagsDelete:     SetDelTags,
		AssetKey:       SetAssetKey,
		NodeNames:      args,
		IpmiWrite:      SetIpmiWrite,
	}

	if !SetYes {
		var nodeCount uint
		// The checks run twice in the prompt case.
		// Avoiding putting in a blocking prompt in an API.
		_, nodeCount, err = node.NodeSetParameterCheck(&set, false)
		if err != nil {
			return
		}
		yes := util.ConfirmationPrompt(fmt.Sprintf("Are you sure you want to modify %d nodes(s)", nodeCount))
		if !yes {
			return
		}
	}
	return node.NodeSet(&set)
}
