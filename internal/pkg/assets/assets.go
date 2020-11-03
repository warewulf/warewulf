
package assets

import (
    "gopkg.in/yaml.v2"
    "io/ioutil"

    "github.com/hpcng/warewulf/internal/pkg/errors"
)


const ConfigFile = "/etc/warewulf/nodes.yaml"


func init() {


}

type nodeYaml struct {
    NodeGroups map[string]nodeGroup //`yaml:"nodegroups"`
}

type nodeGroup struct {
    Comment       string
    Vnfs          string
    Overlay       string
    DomainSuffix  string `yaml:"domain suffix"`
    KernelVersion string `yaml:"kernel version"`
    Nodes         map[string]nodeEntry
}

type nodeEntry struct {
    Hostname      string
    Vnfs          string
    Overlay       string
    DomainSuffix  string `yaml:"domain suffix"`
    KernelVersion string `yaml:"kernel version"`
    IpmiIpaddr    string `yaml:"ipmi ipaddr"`
    NetDevs       map[string]netDevs
}

type netDevs struct {
    Type          string
    Hwaddr        string
    Ipaddr        string
    Netmask       string
    Gateway       string
}

type NodeInfo struct {
    GroupName     string
    HostName      string
    Fqdn          string
    Vnfs          string
    Overlay       string
    KernelVersion string
    DomainName    string
    NetDevs       map[string]netDevs
}


func FindAllNodes() ([]NodeInfo, error) {
    var c nodeYaml
    var ret []NodeInfo

    fd, err := ioutil.ReadFile(ConfigFile)
    if err != nil {
        return nil, err
    }

    err = yaml.Unmarshal(fd, &c)
    if err != nil {
        return nil, err
    }

    for groupname, group := range c.NodeGroups {
        for _, node := range group.Nodes {
            var n NodeInfo

            n.GroupName       = groupname
            n.HostName        = node.Hostname

            n.Vnfs            = group.Vnfs
            n.Overlay         = group.Overlay
            n.KernelVersion   = group.KernelVersion
            n.DomainName      = group.DomainSuffix
            n.NetDevs         = node.NetDevs

            if group.DomainSuffix != "" {
                n.Fqdn = node.Hostname + "." + group.DomainSuffix
            } else {
                n.Fqdn = node.Hostname
            }
            if node.KernelVersion != "" {
                n.KernelVersion = node.KernelVersion
            }
            if node.Vnfs != "" {
                n.Vnfs = node.Vnfs
            }
            if node.Overlay != "" {
                n.Overlay = node.Overlay
            }
            if node.DomainSuffix != "" {
                n.DomainName = node.DomainSuffix
            }
            ret = append(ret, n)
        }
    }

    return ret, nil
}


func FindByHwaddr(hwa string) (NodeInfo, error) {
    var ret NodeInfo

    nodeList, err := FindAllNodes()
    if err != nil {
        return ret, err
    }

    for _, node := range nodeList {
        for _, dev := range node.NetDevs {
            if dev.Hwaddr == hwa {
                return node, nil
            }
        }
    }

    return ret, errors.New("No nodes found with HW Addr: " + hwa)
}


func FindAllVnfs() ([]string, error) {
    var ret []string
    set := make(map[string]bool)

    nodeList, err := FindAllNodes()
    if err != nil {
        return ret, err
    }

    for _, node := range nodeList {
        if node.Vnfs != "" {
            set[node.Vnfs] = true
        }
    }

    for entry := range set {
        ret = append(ret, entry)
    }

    return ret, nil
}



func FindAllKernels() ([]string, error) {
    var ret []string
    set := make(map[string]bool)

    nodeList, err := FindAllNodes()
    if err != nil {
        return ret, err
    }

    for _, node := range nodeList {
        if node.KernelVersion != "" {
            set[node.KernelVersion] = true
        }
    }

    for entry := range set {
        ret = append(ret, entry)
    }

    return ret, nil
}

//FindAllOverlays
func FindAllOverlays() ([]string, error) {
    var ret []string
    set := make(map[string]bool)

    nodeList, err := FindAllNodes()
    if err != nil {
        return ret, err
    }

    for _, node := range nodeList {
        if node.Overlay != "" {
            set[node.Overlay] = true
        }
    }

    for entry := range set {
        ret = append(ret, entry)
    }

    return ret, nil
}


