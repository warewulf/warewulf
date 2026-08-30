package unset

// Vars holds the parsed flags for an unset command. It is shared between
// node unset and profile unset, which are structurally identical.
type Vars struct {
	UnsetYes       bool
	UnsetForce     bool
	UnsetFields    map[string]*bool
	UnsetScopes    map[string]string
	Netname        string
	Diskname       string
	Partname       string
	Fsname         string
	NetnameChanged bool
	Tags           []string
	IpmiTags       []string
	NetTags        []string
}
