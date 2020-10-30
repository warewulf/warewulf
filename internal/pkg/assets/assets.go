
package assets



import (
    "fmt"
    "io/ioutil"

    "gopkg.in/yaml.v2"
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
    NetDevs       map[string]netDevs
}


func FindAllNodes() []NodeInfo {
    var c nodeYaml
    var ret []NodeInfo

    fd, err := ioutil.ReadFile(ConfigFile)
    if err != nil {
        fmt.Println(err)
    }

    err = yaml.Unmarshal(fd, &c)
    if err != nil {
        fmt.Println(err)
    }

    for groupname, group := range c.NodeGroups {
        for _, node := range group.Nodes {
            var n NodeInfo

            n.GroupName       = groupname
            n.HostName        = node.Hostname

            n.Vnfs            = group.Vnfs
            n.Overlay         = group.Overlay
            n.KernelVersion   = group.KernelVersion
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
            ret = append(ret, n)
        }
    }

    return ret
}


func FindByHwaddr(hwa string) NodeInfo{
    var ret NodeInfo

    for _, node := range FindAllNodes() {
        for _, dev := range node.NetDevs {
            if dev.Hwaddr == hwa {
                return node
            }
        }
    }

    return ret
}


func FindAllVnfs() []string {
    var ret []string
    set := make(map[string]bool)

    for _, node := range FindAllNodes() {
        if node.Vnfs != "" {
            set[node.Vnfs] = true
        }
    }

    for entry := range set {
        ret = append(ret, entry)
    }

    return ret
}



func FindAllKernels() []string {
    var ret []string
    set := make(map[string]bool)

    for _, node := range FindAllNodes() {
        if node.KernelVersion != "" {
            set[node.KernelVersion] = true
        }
    }

    for entry := range set {
        ret = append(ret, entry)
    }

    return ret
}

//FindAllOverlays
func FindAllOverlays() []string {
    var ret []string
    set := make(map[string]bool)

    for _, node := range FindAllNodes() {
        if node.Overlay != "" {
            set[node.Overlay] = true
        }
    }

    for entry := range set {
        ret = append(ret, entry)
    }

    return ret
}


