package config

func BoolP(p *bool) bool {
	return p != nil && *p
}
