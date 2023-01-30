package container

import (
	"io/ioutil"
	"path"
	"strings"

	"github.com/hpcng/warewulf/internal/pkg/buildconfig"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"gopkg.in/yaml.v2"
)

/*
Describe a mount point for a container exec
*/
type MntDetails struct {
	Dest     string `yaml:"Dest"`
	Source   string `yaml:"Source,omitempty"`
	ReadOnly bool   `yaml:"Readonly,omitempty" default:"false"`
	Options  string `yaml:"Options,omitempty"` // ignored at the moment
}

/*
Holds the containers, so that they can be marshalled
*/
type MountPoints struct {
	MountPnts []*MntDetails `yaml:"Mount Points"`
}

/*
Create a slice iof MntDetails from a string slice with following
format "source:[:destination][:readonly]" if destination is not
given, the source is used as destination
*/
func InitMountPnts(binds []string) (mounts []MntDetails) {
	for _, b := range binds {
		bind := strings.Split(b, ":")
		dest := bind[0]
		if len(bind) >= 2 {
			dest = bind[1]
		}
		readonly := false
		if len(bind) >= 3 && bind[2] == "ro" {
			readonly = true
		}
		mntPnt := MntDetails{
			Source:   bind[0],
			Dest:     dest,
			ReadOnly: readonly,
		}
		mounts = append(mounts, mntPnt)
	}
	return mounts
}

/*
Read in the default bind mounts from the configuration file

	warewulf/mounts.conf
*/
func DefaultMntPts() (mounts []MntDetails) {
	data, err := ioutil.ReadFile(path.Join(buildconfig.SYSCONFDIR(), "warewulf/mounts.conf"))
	if err != nil {
		wwlog.Verbose("No default bind mounts for containers: %s", err)
		return mounts
	}
	wwlog.Debug("Unmarshaling the mounts configuration")
	err = yaml.Unmarshal(data, &mounts)
	if err != nil {
		wwlog.Verbose("Couldn't unmarshall default bind mounts for containers: %s", err)
	}
	return mounts
}
