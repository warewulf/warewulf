package vnfs

import (
	"github.com/hpcng/warewulf/internal/pkg/config"
	"github.com/hpcng/warewulf/internal/pkg/errors"
	"github.com/hpcng/warewulf/internal/pkg/util"
	"github.com/hpcng/warewulf/internal/pkg/wwlog"
	"gopkg.in/yaml.v2"
	"io/ioutil"
	"os"
	"path"
	"strings"
)

type VnfsObject struct {
	Name string
	Source string
	Chroot string
	Image string
	Config string
}

func Load (name string) (VnfsObject, error) {
	config := config.New()
	var ret VnfsObject

	if name == "" {
		wwlog.Printf(wwlog.DEBUG, "Called vnfs.Load() without a name, returning error\n")
		return ret, errors.New("Called vnfs.Load() without a VNFS name")
	}

	pathFriendlyName := CleanName(name)

	configFile := path.Join(config.VnfsImageDir(pathFriendlyName), "config.yaml")

	if util.IsFile(configFile) == false {
		return ret, errors.New("VNFS has not been imported: " + name)
	}

	data, err := ioutil.ReadFile(configFile)
	if err != nil {
		return ret, errors.New("Error reading VNFS configuration file: " + name)
	}

	err = yaml.Unmarshal(data, &ret)
	if err != nil {
		return ret, err
	}

	return ret, nil
}

func CleanName(source string) string {
	var tmp string

	if strings.HasPrefix(source, "/") == true {
		tmp = path.Base(source)
	} else {
		tmp = source
	}

	tmp = strings.ReplaceAll(tmp, "://", "-")
	tmp = strings.ReplaceAll(tmp, "/", ".")
	tmp = strings.ReplaceAll(tmp, ":", "-")

	return tmp
}

func New(source string) (VnfsObject, error) {
	var ret VnfsObject
	config := config.New()

	if source == "" {
		wwlog.Printf(wwlog.DEBUG, "Called vnfs.Load() without a name, returning error\n")
		return ret, errors.New("Called vnfs.Load() without a VNFS name")
	}

	pathFriendlyName := CleanName(source)

	if strings.HasPrefix(source, "/") == true {
		ret.Source = source
		ret.Name = pathFriendlyName
	} else {
		tmp := strings.ReplaceAll(source, "://", "-")
		tmp = strings.ReplaceAll(tmp, "/", ".")
		tmp = strings.ReplaceAll(tmp, ":", ".")
		ret.Name = source
		ret.Source = source
	}

	ret.Chroot = config.VnfsChroot(pathFriendlyName)
	ret.Image = config.VnfsImage(pathFriendlyName)
	ret.Config = path.Join(config.VnfsImageDir(pathFriendlyName), "config.yaml")

	if util.IsFile(ret.Config) {
		return Load(source)
	}

	return ret, nil
}

func (self *VnfsObject) SaveConfig() error {

	out, err := yaml.Marshal(self)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(self.Config, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		wwlog.Printf(wwlog.ERROR, "%s\n", err)
		os.Exit(1)
	}

	defer file.Close()

	_, err = file.WriteString(string(out))
	if err != nil {
		return err
	}

	return nil
}






func Build(name string, force bool) error {

	vnfs, err := New(name)
	if err != nil {
		return err
	}

	wwlog.Printf(wwlog.VERBOSE, "Building VNFS: %s\n", vnfs.Name)
	if strings.HasPrefix(vnfs.Source, "/") {
		if strings.HasSuffix(vnfs.Source, "tar.gz") {
			//wwlog.Printf(wwlog.WARN, "Building VNFS from local tarball: %s\n", uri)
			wwlog.Printf(wwlog.WARN, "Building VNFS from local tarball is not supported yet: %s\n", vnfs.Name)
		} else {
			BuildContainerdir(vnfs, force)
		}
	} else {
		BuildDocker(vnfs, force)
	}

	err = vnfs.SaveConfig()
	if err != nil {
		return err
	}

	return nil
}







func (self *VnfsObject) Nameold() string {
	if self.Source == "" {
		return ""
	}

	if strings.HasPrefix(self.Source, "/") {
		return path.Base(self.Source)
	}

	return self.Source
}

func NameClean1(SourcePath string) string {
	if SourcePath == "" {
		return ""
	}

	if strings.HasPrefix(SourcePath, "/") {
		return path.Base(SourcePath)
	}
	uri := strings.Split(SourcePath, "://")

	return strings.ReplaceAll(uri[0]+":"+uri[1], "/", "_")
}