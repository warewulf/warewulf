
package main

import (
    "fmt"
    "io"
    "os"
    "strconv"
    "strings"

    "github.com/hpcng/warewulf/internal/pkg/errors"
    "github.com/hpcng/warewulf/internal/pkg/assets"
    "net/http"
)
// TODO: https://github.com/danderson/netboot/blob/master/pixiecore/dhcp.go
// TODO: https://github.com/pin/tftp


const LocalStateDir = "/var/warewulf"

func getSanity(req *http.Request) (assets.NodeInfo, error) {
    url := strings.Split(req.URL.Path, "/")

    hwaddr := strings.ReplaceAll(url[2], "-", ":")
    node, err := assets.FindByHwaddr(hwaddr)
    if err != nil {
        return node, errors.New("Could not find HW address")
    }

    if node.Fqdn == "" {
        fmt.Printf("UNKNOWN: %15s: %s\n", node.Fqdn, req.URL.Path)
        return node, errors.New("Unknown Node: "+ hwaddr)
    }
    fmt.Printf("GET:   %15s: %s\n", node.Fqdn, req.URL.Path)

    return node, nil
}

/*
func files(w http.ResponseWriter, req *http.Request) {
    url := strings.Split(req.URL.Path, "/")

    node := assets.FindByHwaddr(strings.ReplaceAll(url[3], "-", ":"))

    if node.Fqdn == "" {
        fmt.Printf("UNKNOWN: %15s: %s\n", node.Fqdn, req.URL.Path)
    }
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
    } else if url[2] == "runtime" {
        fmt.Printf("FROM: %s\n", req.RemoteAddr)

        remote := strings.Split(req.RemoteAddr, ":")
        port, _ := strconv.Atoi(remote[1])

        if port >= 1024 {
            fmt.Printf("DENIED: Connection coming from non-privledged port: %s\n", req.RemoteAddr)
            return
        }

        if node.Overlay!= "" {
            overlayFile := fmt.Sprintf("%s/provision/runtime/%s.img", LocalStateDir, node.Fqdn)

            sendFile(w, overlayFile, node.Fqdn)
        }
    }

    return
}
 */

func sendFile(w http.ResponseWriter, filename string, sendto string) error {

    fmt.Printf("SEND:  %15s: %s\n", sendto, filename)

    fd, err := os.Open(filename)
    if err != nil {
        fmt.Printf("ERROR:   %s\n", err)
        return err
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

    fd.Close()
    return nil
}


func main() {

    http.HandleFunc("/ipxe/", ipxe)
    http.HandleFunc("/kernel/", kernel)
    http.HandleFunc("/kmods/", kmods)
    http.HandleFunc("/vnfs/", vnfs)
    http.HandleFunc("/overlay/", overlay)
    http.HandleFunc("/runtime/", runtime)

    http.ListenAndServe(":9873", nil)
}
