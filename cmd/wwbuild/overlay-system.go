package main

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/assets"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"os"
	"os/exec"
	"path"
	"sync"
	"time"
)

func overlaySystem(node assets.NodeInfo, replace map[string]string, wg *sync.WaitGroup) {
	defer wg.Done()

	OverlayDir := fmt.Sprintf("%s/overlays/system/%s", LocalStateDir, node.SystemOverlay)
	OverlayFile := fmt.Sprintf("%s/provision/overlays/system/%s.img", LocalStateDir, node.Fqdn)

	destModTime := time.Time{}
	destMod, err := os.Stat(OverlayFile)
	if err == nil {
		destModTime = destMod.ModTime()
	}
	configMod, err := os.Stat("/etc/warewulf/nodes.yaml")
	if err != nil {
		fmt.Printf("ERROR: could not find node file: /etc/warewulf/nodes.yaml")
		os.Exit(1)
	}
	configModTime := configMod.ModTime()
	sourceModTime, _ := util.DirModTime(OverlayDir)

	err = os.MkdirAll(path.Dir(OverlayFile), 0755)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}
	err = os.MkdirAll(OverlayDir, 0755)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}

	if sourceModTime.After(destModTime) || configModTime.After(destModTime) {
		fmt.Printf("SYSTEM:  %s\n", node.Fqdn)

		overlayDest := "/tmp/.overlay-" + util.RandomString(16)
		BuildOverlayDir(OverlayDir, overlayDest, replace)

		cmd := fmt.Sprintf("cd %s && find . | cpio --quiet -o -H newc -F \"%s\"", overlayDest, OverlayFile)
		err = exec.Command("/bin/sh", "-c", cmd).Run()
		if err != nil {
			fmt.Printf("%s", err)
		}

		os.RemoveAll(overlayDest)
	}
}
