
package main


import(
    "os"
    "os/exec"
    "fmt"
    "path"

    "github.com/hpcng/warewulf/internal/pkg/assets"
)


const LocalStateDir = "/var/warewulf"



func main(){

    if len(os.Args) < 2 {
        fmt.Printf("USAGE: %s [vnfs/kernel/overlays/all]\n", os.Args[0]);
        return
    }


    if os.Args[1] == "vnfs" {
        for _, vnfs := range assets.FindAllVnfs() {
            if _, err := os.Stat(vnfs); err == nil {
                cmd := fmt.Sprintf("cd %s; find . | cpio --quiet -o -H newc | gzip -c > \"%s/provision/bases/%s.img.gz\"", vnfs, LocalStateDir, path.Base(vnfs));
                fmt.Printf("BUILDING VNFS:  %s\n", vnfs);
                out, err := exec.Command("/bin/sh", "-c", cmd).Output();
                if err != nil {
                    fmt.Printf("%s", err)
                }

                output := string(out[:])
                fmt.Println(output)
            } else {
                fmt.Printf("SKIPPING VNFS:  (bad path) %s\n", vnfs);
            }
        }
    } else if os.Args[1] == "kernel" {
        for _, kernelVers := range assets.FindAllKernels() {
            kernelSource := fmt.Sprintf("/boot/vmlinuz-%s", kernelVers)
            if _, err := os.Stat(kernelSource); err == nil {
                kernelDestination := fmt.Sprintf("%s/provision/kernels/vmlinuz-%s", LocalStateDir, kernelVers);
                fmt.Printf("SETUP KERNEL:   %s (%s)\n", kernelSource, kernelDestination);
                err := exec.Command("cp", kernelSource, kernelDestination).Run()
                if err != nil {
                    fmt.Printf("%s", err)
                }

                kernelMods := fmt.Sprintf("/lib/modules/%s", kernelVers)
                if _, err := os.Stat(kernelMods); err == nil {
                    fmt.Printf("BUILDING MODS:  %s\n", kernelMods);
                    cmd := fmt.Sprintf("find %s | cpio --quiet -o -H newc -F \"%s/provision/kernels/kmods-%s.img\"", kernelMods, LocalStateDir, kernelVers);
                    err := exec.Command("/bin/sh", "-c", cmd).Run();
                    if err != nil {
                        fmt.Printf("OUTPUT: %s", err)
                    }

                }
            }
        }
    } else if os.Args[1] == "overlay" {
        fmt.Printf("note: This needs to create an overlay for each node with macro expansions\n");


        for _, node := range assets.FindAllNodes() {
            overlayDir := fmt.Sprintf("/etc/warewulf/overlays/%s", node.Overlay);
            if _, err := os.Stat(overlayDir); err == nil {
                cmd := fmt.Sprintf("cd %s; find . | cpio --quiet -o -H newc | gzip -c > \"%s/provision/overlays/%s.img\"", overlayDir, LocalStateDir, node.Fqdn);
                fmt.Printf("BUILDING OVERLAY:  %s\n", node.Fqdn);
                err := exec.Command("/bin/sh", "-c", cmd).Run();
                if err != nil {
                    fmt.Printf("%s", err)
                }
            } else {
                fmt.Printf("SKIPPING OVERLAY:  (bad path) %s\n", overlayDir);
            }
        }
        

    }
}
