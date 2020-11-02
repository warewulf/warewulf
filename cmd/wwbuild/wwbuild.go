
package main

import (
    "fmt"
    "github.com/hpcng/warewulf/internal/pkg/assets"
    "github.com/hpcng/warewulf/internal/pkg/util"
    "os"
    "os/exec"
    "path"
    "strings"
)


const LocalStateDir = "/var/warewulf"

func vnfsBuild(vnfsPath string) {
    fmt.Printf("BUILDING VNFS:  %s\n", vnfsPath)
    if _, err := os.Stat(vnfsPath); err == nil {
        // TODO: Build VNFS to temporary file and move to real location when complete atomically
        // TODO: Check time stamps of sourcedir and build file to see if we need to rebuild or skip
        cmd := fmt.Sprintf("cd %s; find . | cpio --quiet -o -H newc | gzip -c > \"%s/provision/bases/%s.img.gz\"", vnfsPath, LocalStateDir, path.Base(vnfsPath))
        err := exec.Command("/bin/sh", "-c", cmd).Run()
        if err != nil {
            fmt.Printf("%s", err)
        }
    } else {
        fmt.Printf("SKIPPING VNFS:  (bad path) %s\n", vnfsPath)
    }
}



func main(){

    if len(os.Args) < 2 {
        fmt.Printf("USAGE: %s [vnfs/kernel/overlays/all]\n", os.Args[0])
        return
    }


    if os.Args[1] == "vnfs" {
        if len(os.Args) >= 3 {
            vnfsBuild(os.Args[3])
        } else {
            for _, vnfs := range assets.FindAllVnfs() {
                vnfsBuild(vnfs)
            }
        }
    } else if os.Args[1] == "kernel" {
        for _, kernelVers := range assets.FindAllKernels() {
            kernelSource := fmt.Sprintf("/boot/vmlinuz-%s", kernelVers)
            // TODO: Check time stamps of source and dests to see if we need to rebuild or skip
            if _, err := os.Stat(kernelSource); err == nil {
                kernelDestination := fmt.Sprintf("%s/provision/kernels/vmlinuz-%s", LocalStateDir, kernelVers)
                fmt.Printf("SETUP KERNEL:   %s (%s)\n", kernelSource, kernelDestination)
                err := exec.Command("cp", kernelSource, kernelDestination).Run()
                if err != nil {
                    fmt.Printf("%s", err)
                }

                kernelMods := fmt.Sprintf("/lib/modules/%s", kernelVers)
                if _, err := os.Stat(kernelMods); err == nil {
                    fmt.Printf("BUILDING MODS:  %s\n", kernelMods)
                    cmd := fmt.Sprintf("find %s | cpio --quiet -o -H newc -F \"%s/provision/kernels/kmods-%s.img\"", kernelMods, LocalStateDir, kernelVers)
                    err := exec.Command("/bin/sh", "-c", cmd).Run()
                    if err != nil {
                        fmt.Printf("OUTPUT: %s", err)
                    }

                }
            }
        }
    } else if os.Args[1] == "overlay" {
        fmt.Printf("note: This needs to create an overlay for each node with macro expansions\n")

        for _, node := range assets.FindAllNodes() {

            overlayDir := fmt.Sprintf("/etc/warewulf/overlays/%s", node.Overlay)
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

            destFile := fmt.Sprintf("%s/provision/overlays/%s.img", LocalStateDir, node.Fqdn)

            destMod, _ := os.Stat(destFile)
            destModTime := destMod.ModTime()

            configMod, _ := os.Stat("/etc/warewulf/nodes.yaml")
            configModTime := configMod.ModTime()

            sourceModTime := util.DirModTime(overlayDir)

            if sourceModTime.After(destModTime) || configModTime.After(destModTime) {
                fmt.Printf("BUILDING OVERLAY:  %s\n", node.Fqdn)

                overlayDest := "/tmp/.overlay-" + util.RandomString(16)
                BuildOverlayDir(overlayDir, overlayDest, replace)

                cmd := fmt.Sprintf("cd %s; find . | cpio --quiet -o -H newc -F \"%s\"", overlayDest, destFile)
                err := exec.Command("/bin/sh", "-c", cmd).Run()
                if err != nil {
                    fmt.Printf("%s", err)
                }

                os.RemoveAll(overlayDest)
            } else {
                fmt.Printf("Skipping overlay (nothing changed): %s\n", node.Fqdn)
            }
        }
    }
}
