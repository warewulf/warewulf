package main

import (
    "fmt"
    "github.com/hpcng/warewulf/internal/pkg/assets"
    "os"
    "os/exec"
    "path"
    "strings"
    "sync"
)

const LocalStateDir = "/var/warewulf"

func vnfsBuild(vnfsPath string, wg *sync.WaitGroup) {
	defer wg.Done()
	if _, err := os.Stat(vnfsPath); err == nil {
		// TODO: Build VNFS to temporary file and move to real location when complete atomically
		// TODO: Check time stamps of sourcedir and build file to see if we need to rebuild or skip
		vnfsDestination := fmt.Sprintf("%s/provision/vnfs/%s.img.gz", LocalStateDir, path.Base(vnfsPath))

		err := os.MkdirAll(path.Dir(vnfsDestination), 0755)
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			return
		}

		fmt.Printf("BUILDING VNFS:  %s\n", vnfsPath)

		cmd := fmt.Sprintf("cd %s; find . | cpio --quiet -o -H newc | gzip -c > \"%s\"", vnfsPath, vnfsDestination)
		err = exec.Command("/bin/sh", "-c", cmd).Run()
		if err != nil {
			fmt.Printf("ERROR: %s\n", err)
			os.Exit(1)
		}
		fmt.Printf("BUILD DONE:     %s\n", vnfsPath)

	} else {
		fmt.Printf("SKIPPING VNFS:  (bad path) %s\n", vnfsPath)
	}
}

func main() {

	if len(os.Args) < 2 {
		fmt.Printf("USAGE: %s [vnfs/kernel/overlay/all]\n", os.Args[0])
		return
	}

	if os.Args[1] == "vnfs" {
		var nodeList []assets.NodeInfo
		set := make(map[string]bool)
		var wg sync.WaitGroup

		if len(os.Args) < 3 {
			fmt.Printf("USAGE: %s vnfs [node name pattern/ALL]\n", os.Args[0])
			return
		}

		if os.Args[2] == "ALL" {
			nodeList, _ = assets.FindAllNodes()
		} else {
			nodeList, _ = assets.SearchByName(os.Args[2])
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
			go vnfsBuild(entry, &wg)
		}
		wg.Wait()

	} else if os.Args[1] == "kernel" {
		var nodeList []assets.NodeInfo
		set := make(map[string]bool)

		if len(os.Args) < 3 {
			fmt.Printf("USAGE: %s vnfs [node name pattern/ALL]\n", os.Args[0])
			return
		}

		if os.Args[2] == "ALL" {
			nodeList, _ = assets.FindAllNodes()
		} else {
			nodeList, _ = assets.SearchByName(os.Args[2])
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
				kernelDestination := fmt.Sprintf("%s/provision/kernel/vmlinuz-%s", LocalStateDir, kernelVers)
				kmodsDestination := fmt.Sprintf("%s/provision/kernel/kmods-%s.img", LocalStateDir, kernelVers)

				err := os.MkdirAll(path.Dir(kernelDestination), 0755)
				if err != nil {
					fmt.Printf("ERROR: %s\n", err)
					return
				}

				fmt.Printf("SETUP KERNEL:   %s (%s)\n", kernelSource, kernelDestination)
				err = exec.Command("cp", kernelSource, kernelDestination).Run()
				if err != nil {
					fmt.Printf("%s", err)
				}

				kernelMods := fmt.Sprintf("./lib/modules/%s", kernelVers)
				if _, err := os.Stat(kernelMods); err == nil {
					fmt.Printf("BUILDING MODS:  %s\n", kernelMods)
					cmd := fmt.Sprintf("cd /; find %s | cpio --quiet -o -H newc -F \"%s\"", kernelMods, kmodsDestination)
					err := exec.Command("/bin/sh", "-c", cmd).Run()
					if err != nil {
						fmt.Printf("OUTPUT: %s", err)
					}

				}
			}
		}
	} else if os.Args[1] == "overlay" {
		var nodeList []assets.NodeInfo
		var wg sync.WaitGroup

		if len(os.Args) < 3 {
			fmt.Printf("USAGE: %s vnfs [node name pattern/ALL]\n", os.Args[0])
			return
		}

		if os.Args[2] == "ALL" {
			nodeList, _ = assets.FindAllNodes()
		} else {
			nodeList, _ = assets.SearchByName(os.Args[2])
		}

		if len(nodeList) == 0 {
			fmt.Printf("ERROR: No nodes found\n")
			return
		}

		for _, node := range nodeList {

			replace := make(map[string]string)
			replace["HOSTNAME"] = node.HostName
			replace["FQDN"] = node.Fqdn
			replace["VNFS"] = node.Vnfs
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
