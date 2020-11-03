
package main

import (
    "fmt"
    "github.com/hpcng/warewulf/internal/pkg/assets"
    "github.com/hpcng/warewulf/internal/pkg/util"
    "log"
    "os"
    "os/exec"
    "path"
    "strings"
    "time"
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
            nodeList, err := assets.FindAllVnfs()
            if err != nil {
                log.Panicf("Could not locate VNFS images: %s\n", err)
                os.Exit(1)
            }

            for _, vnfs := range nodeList {
                vnfsBuild(vnfs)
            }
        }
    } else if os.Args[1] == "kernel" {
        nodeList, err := assets.FindAllKernels()
        if err != nil {
            log.Panicf("Could not locate Kernel Versions: %s\n", err)
            os.Exit(1)
        }

        for _, kernelVers := range nodeList {
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
        //TODO: Move this all to warewulfd and generate on demand when needed
        nodeList, err := assets.FindAllNodes()
        if err != nil {
            log.Panicf("Could not identify nodes: %s\n", err)
            os.Exit(1)
        }

        for _, node := range nodeList {

            overlayDir := fmt.Sprintf("/etc/warewulf/overlays/%s", node.Overlay)

            //TODO: Move this all to the Asset package
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

            destModTime := time.Time{}
            destMod, err := os.Stat(destFile)
            if err == nil {
                destModTime = destMod.ModTime()
            }

            configMod, err := os.Stat("/etc/warewulf/nodes.yaml")
            if err != nil {
                fmt.Printf("ERROR: could not find node file: /etc/warewulf/nodes.yaml")
                os.Exit(1)
            }
            configModTime := configMod.ModTime()

            sourceModTime, _ := util.DirModTime(overlayDir)

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
