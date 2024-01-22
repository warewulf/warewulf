package profile

type ProfileListEntry interface {
	GetHeader() []string
	GetValue() []string
}

type ProfileListResponse struct {
	Profiles map[string][]ProfileListEntry `yaml:"Profiles" json:"Profiles"`
}

type ProfileListSimpleEntry struct {
	CommentDesc string `yaml:"Comment/Description" json:"Comment/Description"`
}

func (p *ProfileListSimpleEntry) GetHeader() []string {
	return []string{"PROFILE NAME", "COMMENT/DESCRIPTION"}
}

func (p *ProfileListSimpleEntry) GetValue() []string {
	return []string{p.CommentDesc}
}

type ProfileListLongEntry struct {
	Field   string `yaml:"Field" json:"Field"`
	Profile string `yaml:"Profile" json:"Profile"`
	Value   string `yaml:"Value" json:"Value"`
}

func (p *ProfileListLongEntry) GetHeader() []string {
	return []string{"PROFILE", "FIELD", "PROFILE", "VALUE"}
}

func (p *ProfileListLongEntry) GetValue() []string {
	return []string{p.Field, p.Profile, p.Value}
}
