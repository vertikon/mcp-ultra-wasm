package models

// TemplateMetadata descreve um template disponível para geração.
type TemplateMetadata struct {
	Name        string              `yaml:"name" json:"name"`
	DisplayName string              `yaml:"display_name" json:"display_name"`
	Description string              `yaml:"description" json:"description"`
	Version     string              `yaml:"version" json:"version"`
	Variables   []TemplateVariable  `yaml:"variables" json:"variables"`
	Tags        []string            `yaml:"tags" json:"tags"`
	Defaults    map[string]string   `yaml:"defaults" json:"defaults"`
}

// TemplateVariable define os campos parametrizáveis.
type TemplateVariable struct {
	Key         string `yaml:"key" json:"key"`
	Description string `yaml:"description" json:"description"`
	Required    bool   `yaml:"required" json:"required"`
}

