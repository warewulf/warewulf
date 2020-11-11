package main

import (
	"fmt"
	"github.com/hpcng/warewulf/internal/pkg/assets"
	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"os"
	"os/exec"
	"path"
	"sync"
)

func overlayRuntime(node assets.NodeInfo, replace map[string]string, wg *sync.WaitGroup) {
	defer wg.Done()

	OverlayDir := fmt.Sprintf("%s/overlays/runtime/%s", config.LocalStateDir, node.RuntimeOverlay)
	OverlayFile := fmt.Sprintf("%s/provision/overlays/runtime/%s.img", config.LocalStateDir, node.Fqdn)
	/*
		destModTime := time.Time{}
		destMod, err := os.Stat(OverlayFile)
		if err == n

	il {
			destModTime = destMod.ModTime()
		}
		configMod, err := os.Stat("/etc/warewulf/nodes.conf")
		if err != nil {
			fmt.Printf("ERROR: could not find node file: /etc/warewulf/nodes.conf")
			os.Exit(1)
		}
		configModTime := configMod.ModTime()
		sourceModTime, _ := util.DirModTime(OverlayDir)
	*/
	err := os.MkdirAll(path.Dir(OverlayFile), 0755)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}
	err = os.MkdirAll(OverlayDir, 0755)
	if err != nil {
		fmt.Printf("ERROR: %s\n", err)
		return
	}
	//	if sourceModTime.After(destModTime) || configModTime.After(destModTime) {
	fmt.Printf("RUNTIME: %s\n", node.Fqdn)

	overlayDest := "/tmp/.system-overlay-" + util.RandomString(16)
	BuildOverlayDir(OverlayDir, overlayDest, replace)

	cmd := fmt.Sprintf("cd %s && find . | cpio --quiet -o -H newc -F \"%s\"", overlayDest, OverlayFile)
	err = exec.Command("/bin/sh", "-c", cmd).Run()
	if err != nil {
		fmt.Printf("%s", err)
	}

	os.RemoveAll(overlayDest)
	//	} else {
	//		fmt.Printf("RUNTIME: %s (skipped no changes)\n", node.Fqdn)
	//	}
}
