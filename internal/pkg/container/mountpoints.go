package container

import (
	"strings"

	warewulfconf "github.com/warewulf/warewulf/internal/pkg/config"
	"github.com/warewulf/warewulf/internal/pkg/wwlog"
)

/*
Create a slice iof MntDetails from a string slice with following
format "source:[:destination][:readonly]" if destination is not
given, the source is used as destination
*/
func InitMountPnts(binds []string) (mounts []*warewulfconf.MountEntry) {
	wwlog.Debug("Trying to mount following mount points: %s", mounts)
	for _, b := range binds {
		bind := strings.Split(b, ":")
		dest := bind[0]
		if len(bind) >= 2 {
			dest = bind[1]
		}
		readonly := false
		copy_ := false
		if len(bind) >= 3 {
			if bind[2] == "ro" {
				readonly = true
			} else if bind[2] == "copy" {
				copy_ = true
			}
		}
		mntPnt := warewulfconf.MountEntry{
			Source:    bind[0],
			Dest:      dest,
			ReadOnlyP: &readonly,
			CopyP:     &copy_,
		}
		mounts = append(mounts, &mntPnt)
	}
	return mounts
}
