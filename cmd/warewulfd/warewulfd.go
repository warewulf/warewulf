
package main

import (
    "fmt"
    "io"
    "os"
    "path"
    "strconv"
    "strings"

    "github.com/hpcng/warewulf/internal/pkg/assets"
    "net/http"
)
// TODO: https://github.com/danderson/netboot/blob/master/pixiecore/dhcp.go
// TODO: https://github.com/pin/tftp


const LocalStateDir = "/var/warewulf"

func ipxe(w http.ResponseWriter, req *http.Request) {
    url := strings.Split(req.URL.Path, "/")

    if url[2] == "" {
        fmt.Printf("ERROR: Bad iPXE request from %s\n", req.RemoteAddr)
        return
    }

    node := assets.FindByHwaddr(url[2])

    if node.HostName != "" {
        fmt.Printf("IPXE:  %15s: hwaddr=%s\n", node.Fqdn, url[2])

        fmt.Fprintf(w, "#!ipxe\n")

        fmt.Fprintf(w, "echo Now booting Warewulf - v4 Proof of Concept\n")
        fmt.Fprintf(w, "set base http://192.168.1.1:9873/\n")
        fmt.Fprintf(w, "kernel ${base}/files/kernel/%s crashkernel=no quiet\n", url[2])
        fmt.Fprintf(w, "initrd ${base}/files/vnfs/%s\n", url[2])
        fmt.Fprintf(w, "initrd ${base}/files/kmods/%s\n", url[2])
        fmt.Fprintf(w, "initrd ${base}/files/overlay/%s\n", url[2])
        fmt.Fprintf(w, "boot\n")
    } else {
        fmt.Printf("ERROR: iPXE request from unknown Node (hwaddr=%s)\n", url[2])
    }
    return
}


func files(w http.ResponseWriter, req *http.Request) {
    url := strings.Split(req.URL.Path, "/")

    node := assets.FindByHwaddr(url[3])

    fmt.Printf("GET:   %15s: %s\n", node.Fqdn, req.URL.Path)

    if url[2] == "kernel" {
        if node.KernelVersion != "" {
            kernelFile := fmt.Sprintf("%s/provision/kernels/vmlinuz-%s", LocalStateDir, node.KernelVersion)

            sendFile(w, kernelFile, node.Fqdn)
        }
    } else if url[2] == "kmods" {
        if node.KernelVersion != "" {
            kmodsFile := fmt.Sprintf("%s/provision/kernels/kmods-%s.img", LocalStateDir, node.KernelVersion)

            sendFile(w, kmodsFile, node.Fqdn)
        }
    } else if url[2] == "vnfs" {
        if node.Vnfs != "" {
            vnfsFile := fmt.Sprintf("%s/provision/bases/%s.img.gz", LocalStateDir, path.Base(node.Vnfs))

            sendFile(w, vnfsFile, node.Fqdn)
        }
    } else if url[2] == "overlay" {
        if node.Overlay!= "" {
            overlayFile := fmt.Sprintf("%s/provision/overlays/%s.img", LocalStateDir, node.Fqdn)

            sendFile(w, overlayFile, node.Fqdn)
        }
    }
    return
}


func sendFile(w http.ResponseWriter, filename string, sendto string) {

    fmt.Printf("SEND:  %15s: %s\n", sendto, filename)

    fd, err := os.Open(filename)
    if err != nil {
        fmt.Println("ERROR:   %s\n", err)
        return
    }

    FileHeader := make([]byte, 512)
    fd.Read(FileHeader)
    FileContentType := http.DetectContentType(FileHeader)
    FileStat, _ := fd.Stat()
    FileSize := strconv.FormatInt(FileStat.Size(), 10)

    w.Header().Set("Content-Disposition", "attachment; filename=kernel")
    w.Header().Set("Content-Type", FileContentType)
    w.Header().Set("Content-Length", FileSize)

    fd.Seek(0, 0)
    io.Copy(w, fd)
}



func main() {

    http.HandleFunc("/ipxe/", ipxe)
    http.HandleFunc("/files/", files)

    http.ListenAndServe(":9873", nil)
}
