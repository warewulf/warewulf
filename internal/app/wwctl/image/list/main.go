package list

import (
	"os"
	"strconv"
	"time"

	"github.com/spf13/cobra"

	"github.com/warewulf/warewulf/internal/app/wwctl/table"
	"github.com/warewulf/warewulf/internal/pkg/image"
	"github.com/warewulf/warewulf/internal/pkg/kernel"
	"github.com/warewulf/warewulf/internal/pkg/node"
	"github.com/warewulf/warewulf/internal/pkg/util"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

func CobraRunE(vars *variables) func(cmd *cobra.Command, args []string) (err error) {
	return func(cmd *cobra.Command, args []string) (err error) {
		t := table.New(cmd.OutOrStdout())
		showSize := vars.size || vars.chroot || vars.compressed
		if showSize || vars.full || vars.kernel {
			sources, err := image.ListSources()
			if err != nil {
				return err
			}

			nodeDB, err := node.New()
			if err != nil {
				return err
			}
			nodes, err := nodeDB.FindAllNodes()
			if err != nil {
				return err
			}
			nodemap := make(map[string]int)
			for _, n := range nodes {
				nodemap[n.ImageName]++
			}

			if vars.full {
				t.AddHeader("IMAGE NAME", "NODES", "KERNEL VERSION", "CREATION TIME", "MODIFICATION TIME", "SIZE")
				for _, name := range sources {
					if len(args) > 0 && !util.InSlice(args, name) {
						continue
					}
					kernelVersion := ""
					if k := kernel.FindKernels(name).Default(); k != nil {
						kernelVersion = k.Version()
					}
					createTime := time.Unix(0, 0)
					if sourceStat, err := os.Stat(image.SourceDir(name)); err == nil {
						createTime = sourceStat.ModTime()
					}
					modTime := time.Unix(0, 0)
					if imageStat, err := os.Stat(image.ImageFile(name)); err == nil {
						modTime = imageStat.ModTime()
					}
					sz := util.ByteToString(int64(image.ImageSize(name)))
					if vars.compressed {
						sz = util.ByteToString(int64(image.CompressedImageSize(name)))
					}
					if vars.chroot {
						size, err := util.DirSize(image.SourceDir(name))
						if err != nil {
							wwlog.Error("%s", err)
							size = 0
						}
						sz = util.ByteToString(size)
					}
					t.AddLine(
						name,
						strconv.Itoa(nodemap[name]),
						kernelVersion,
						createTime.Format(time.RFC822),
						modTime.Format(time.RFC822),
						sz,
					)
				}
			} else if vars.kernel {
				t.AddHeader("IMAGE NAME", "NODES", "KERNEL VERSION")
				for _, name := range sources {
					if len(args) > 0 && !util.InSlice(args, name) {
						continue
					}
					kernelVersion := ""
					if k := kernel.FindKernels(name).Default(); k != nil {
						kernelVersion = k.Version()
					}
					t.AddLine(
						name,
						strconv.Itoa(nodemap[name]),
						kernelVersion,
					)
				}
			} else if showSize {
				t.AddHeader("IMAGE NAME", "NODES", "SIZE")
				for _, name := range sources {
					if len(args) > 0 && !util.InSlice(args, name) {
						continue
					}
					sz := util.ByteToString(int64(image.ImageSize(name)))
					if vars.compressed {
						sz = util.ByteToString(int64(image.CompressedImageSize(name)))
					}
					if vars.chroot {
						size, err := util.DirSize(image.SourceDir(name))
						if err != nil {
							wwlog.Error("%s", err)
							size = 0
						}
						sz = util.ByteToString(size)
					}
					t.AddLine(
						name,
						strconv.Itoa(nodemap[name]),
						sz,
					)
				}
			}
		} else {
			t.AddHeader("IMAGE NAME")
			list, err := image.ListSources()
			if err != nil {
				return err
			}
			for _, name := range list {
				if len(args) > 0 && !util.InSlice(args, name) {
					continue
				}
				t.AddLine(name)
			}
		}
		t.Print()
		return
	}
}
