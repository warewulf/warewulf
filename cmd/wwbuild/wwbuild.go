package main

import (
	"fmt"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/hpcng/warewulf/internal/pkg/assets"
	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/vnfs"
)

func main() {

	if len(os.Args) < 2 {
		fmt.Printf("USAGE: %s [vnfs/kernel/system-overlay] (node regex)\n", os.Args[0])
		return
	}

	if os.Args[1] == "vnfs" {
		var nodeList []assets.NodeInfo
		set := make(map[string]bool)
		var wg sync.WaitGroup

		if len(os.Args) >= 3 {
			nodeList, _ = assets.SearchByName(os.Args[2])
		} else {
			nodeList, _ = assets.FindAllNodes()
		}

		if len(nodeList) == 0 {
			fmt.Printf("ERROR: No nodes found\n")
			return
		}

		for _, node := range nodeList {
			if node.Vnfs != "" {
				set[node.Vnfs] = true
			}
		}

		for entry := range set {
			wg.Add(1)
			if strings.HasPrefix(entry, "/") {
				vnfsLocalBuild(entry, &wg)
			} else {
				vnfsOciBuild(entry, &wg)
			}
		}

		time.Sleep(1000 * time.Millisecond)
		fmt.Printf("Waiting for build(s) to complete...\n")
		wg.Wait()

	} else if os.Args[1] == "kernel" {
		var nodeList []assets.NodeInfo
		set := make(map[string]bool)

		if len(os.Args) >= 3 {
			nodeList, _ = assets.SearchByName(os.Args[2])
		} else {
			nodeList, _ = assets.FindAllNodes()
		}

		if len(nodeList) == 0 {
			fmt.Printf("ERROR: No nodes found\n")
			return
		}

		for _, node := range nodeList {
			if node.KernelVersion != "" {
				set[node.KernelVersion] = true
			}
		}

		for kernelVers := range set {
			kernelSource := fmt.Sprintf("/boot/vmlinuz-%s", kernelVers)
			// TODO: Check time stamps of source and dests to see if we need to rebuild or skip
			if _, err := os.Stat(kernelSource); err == nil {
				kernelDestination := fmt.Sprintf("%s/provision/kernel/vmlinuz-%s", config.LocalStateDir, kernelVers)
				kmodsDestination := fmt.Sprintf("%s/provision/kernel/kmods-%s.img", config.LocalStateDir, kernelVers)

				err := os.MkdirAll(path.Dir(kernelDestination), 0755)
				if err != nil {
					fmt.Printf("ERROR: %s\n", err)
					return
				}

				fmt.Printf("SETUP KERNEL:   %s\n", kernelSource)
				err = exec.Command("cp", kernelSource, kernelDestination).Run()
				if err != nil {
					fmt.Printf("%s", err)
				}

				kernelMods := fmt.Sprintf("/lib/modules/%s", kernelVers)
				if _, err := os.Stat(kernelMods); err == nil {
					fmt.Printf("BUILDING MODS:  %s\n", kernelMods)
					cmd := fmt.Sprintf("cd /; find .%s | cpio --quiet -o -H newc -F \"%s\"", kernelMods, kmodsDestination)
					err := exec.Command("/bin/sh", "-c", cmd).Run()
					if err != nil {
						fmt.Printf("OUTPUT: %s", err)
					}

				}
			}
		}
	} else if os.Args[1] == "system-overlay" {
		var nodeList []assets.NodeInfo
		var wg sync.WaitGroup

		if len(os.Args) >= 3 {
			nodeList, _ = assets.SearchByName(os.Args[2])
		} else {
			nodeList, _ = assets.FindAllNodes()
		}

		if len(nodeList) == 0 {
			fmt.Printf("ERROR: No nodes found\n")
			return
		}

		for _, node := range nodeList {
			v := vnfs.New(node.Vnfs)
			replace := make(map[string]string)
			replace["HOSTNAME"] = node.HostName
			replace["FQDN"] = node.Fqdn
			replace["VNFS"] = node.Vnfs
			replace["VNFSDIR"] = v.Root()
			replace["KERNELVERSION"] = node.KernelVersion
			replace["GROUPNAME"] = node.GroupName
			replace["DOMAIN"] = node.DomainName
			for key, dev := range node.NetDevs {
				replace[fmt.Sprintf("%s:NAME", key)] = key
				replace[fmt.Sprintf("%s:HWADDR", key)] = strings.ReplaceAll(dev.Hwaddr, "-", ":")
				replace[fmt.Sprintf("%s:IPADDR", key)] = dev.Ipaddr
				replace[fmt.Sprintf("%s:NETMASK", key)] = dev.Netmask
				replace[fmt.Sprintf("%s:GATEWAY", key)] = dev.Gateway
			}

			wg.Add(2)
			overlayRuntime(node, replace, &wg)
			overlaySystem(node, replace, &wg)
		}
		wg.Wait()

	}
}
