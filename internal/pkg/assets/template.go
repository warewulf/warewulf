package assets



type WWTemplate struct {
	GroupName      string
	HostName       string
	DomainName     string
	Fqdn           string
	Vnfs           string
	VnfsDir        string
	Ipxe           string
	SystemOverlay  string
	RuntimeOverlay string
	KernelVersion  string
	NetDevs        map[string]netDevs
}

func (t WWTemplate) Import(file string) {

}