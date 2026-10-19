package node

// ApplyTagChanges applies the tag operations from --tagadd/--tagdel,
// --ipmitagadd/--ipmitagdel, and --nettagadd/--nettagdel to p. Network tags
// target the network device named by add.Net, and are skipped if p does not
// define that network device.
func (p *Profile) ApplyTagChanges(add NodeConfAdd, del NodeConfDel) {
	for _, key := range del.TagsDel {
		delete(p.Tags, key)
	}
	if len(add.TagsAdd) > 0 && p.Tags == nil {
		p.Tags = make(map[string]string)
	}
	for key, val := range add.TagsAdd {
		p.Tags[key] = val
	}

	if p.Ipmi != nil {
		for _, key := range del.IpmiTagsDel {
			delete(p.Ipmi.Tags, key)
		}
	}
	if len(add.IpmiTagsAdd) > 0 {
		if p.Ipmi == nil {
			p.Ipmi = new(IpmiConf)
		}
		if p.Ipmi.Tags == nil {
			p.Ipmi.Tags = make(map[string]string)
		}
	}
	for key, val := range add.IpmiTagsAdd {
		p.Ipmi.Tags[key] = val
	}

	netDev, ok := p.NetDevs[add.Net]
	if !ok || netDev == nil {
		return
	}
	for _, key := range del.NetTagsDel {
		delete(netDev.Tags, key)
	}
	if len(add.NetTagsAdd) > 0 && netDev.Tags == nil {
		netDev.Tags = make(map[string]string)
	}
	for key, val := range add.NetTagsAdd {
		netDev.Tags[key] = val
	}
}
