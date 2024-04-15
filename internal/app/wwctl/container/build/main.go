package build

import (
	"fmt"

	"github.com/spf13/cobra"
	cexec "github.com/warewulf/warewulf/internal/app/wwctl/container/exec"
	"github.com/warewulf/warewulf/internal/pkg/api/container"
	"github.com/warewulf/warewulf/internal/pkg/api/routes/wwapiv1"
	"github.com/warewulf/warewulf/internal/pkg/config"
	pkgcontianer "github.com/warewulf/warewulf/internal/pkg/container"
	"github.com/warewulf/warewulf/internal/pkg/kernel"
	"github.com/warewulf/warewulf/internal/pkg/util"
)

func CobraRunE(cmd *cobra.Command, args []string) error {
	cbp := &wwapiv1.ContainerBuildParameter{
		ContainerNames: args,
		Force:          BuildForce,
		All:            BuildAll,
		Default:        SetDefault,
		Initramfs:      Initramfs,
	}
	if Initramfs {
		return runInitramfsBuild(cmd, cbp)
	}
	return container.ContainerBuild(cbp)
}

func runInitramfsBuild(cmd *cobra.Command, cbp *wwapiv1.ContainerBuildParameter) (err error) {
	// binding the installed dracut modules
	dracutModules := fmt.Sprintf("%s/warewulf/dracut/modules.d/90wwinit", config.Get().Paths.Sysconfdir)
	if util.IsDir(dracutModules) {
		cexec.SetBinds([]string{fmt.Sprintf("%s:/usr/lib/dracut/modules.d/90wwinit", dracutModules)})
	}

	if cbp == nil {
		return fmt.Errorf("ContainerBuildParameter is nill")
	}

	var containers []string
	if cbp.All {
		containers, err = pkgcontianer.ListSources()
	} else {
		containers = cbp.ContainerNames
	}

	if len(containers) == 0 {
		return
	}

	for _, c := range containers {
		// kernel version, we need to set container kernel version as by default, it'll build against
		// host kernel version, which usually does not exist inside container
		var kver string
		rootfsDir := pkgcontianer.RootFsDir(c)
		kver, err = kernel.FindKernelVersion(rootfsDir)
		if err != nil {
			return fmt.Errorf("failed to locate container kernel version: %s", err)
		}

		err = cexec.CobraRunE(cmd, []string{c, "/usr/bin/dracut --no-hostonly --force --verbose --kver " + kver + " /boot/initramfs-" + kver + ".img"})
		if err != nil {
			return
		}

		// make sure the built initramfs exists
		if !util.IsFile(pkgcontianer.InitramfsBootPath(c, kver)) {
			return fmt.Errorf("file %s does not exist, probably the initramfs build failed", pkgcontianer.InitramfsBootPath(c, kver))
		}
	}
	return
}
