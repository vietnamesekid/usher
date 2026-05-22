package registry

type MCPAuth struct {
	EnvVar string `json:"envVar"`
	URL    string `json:"url"`
}

type MCPRegistryEntry struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Command     string   `json:"command"`
	Args        []string `json:"args"`
	Auth        MCPAuth  `json:"auth"`
	Tags        []string `json:"tags"`
}

type SkillSource struct {
	Type string `json:"type"` // "url" | "local"
	URL  string `json:"url"`
}

type SkillRegistryEntry struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Source      SkillSource `json:"source"`
	Versions    []string    `json:"versions"`
	Latest      string      `json:"latest"`
	Tags        []string    `json:"tags"`
}

type Registry interface {
	GetMCP(name string) (MCPRegistryEntry, error)
	GetSkill(name string) (SkillRegistryEntry, error)
	ListMCPs() []MCPRegistryEntry
	ListSkills() []SkillRegistryEntry
}
