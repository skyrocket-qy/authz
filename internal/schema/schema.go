package schema

import (
	"os"

	"gopkg.in/yaml.v2"
)

type SchemaConfig struct {
	Namespaces map[string]*Namespace `yaml:"namespaces" json:"namespaces"`
}

type Namespace struct {
	Relations   map[string]*Relation   `yaml:"relations" json:"relations"`
	Permissions map[string]*Permission `yaml:"permissions" json:"permissions"`
}

type Relation struct {
	Types []string `yaml:"types" json:"types"` // e.g., "user", "group"
}

type Permission struct {
	// One of: union, intersection, exclusion, or a simple string relation
	Union        []string `yaml:"union,omitempty" json:"union,omitempty"`
	Intersection []string `yaml:"intersection,omitempty" json:"intersection,omitempty"`
	Exclusion    []string `yaml:"exclusion,omitempty" json:"exclusion,omitempty"`
	Relation     string   `yaml:"relation,omitempty" json:"relation,omitempty"`
}

func Load() {
	data, err := os.ReadFile("schema.yaml")
	if err != nil {
		panic(err)
	}

	var cfg SchemaConfig
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		panic(err)
	}
}
