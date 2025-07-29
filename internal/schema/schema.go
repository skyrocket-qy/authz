package schema

import (
	"os"

	"gopkg.in/yaml.v2"
)

type Schema struct {
	Namespaces map[string]*Namespace `yaml:"namespaces" json:"namespaces"`
}

type Namespace struct {
	Relations   map[string]*Relation       `yaml:"relations" json:"relations"`
	Permissions map[string]*PermissionExpr `yaml:"permissions" json:"permissions"`
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

type PermissionExpr struct {
	// Exactly one of these should be set
	Relation string            `yaml:"relation,omitempty"`
	Or       []*PermissionExpr `yaml:"union,omitempty"`
	And      []*PermissionExpr `yaml:"intersection,omitempty"`
	Not      []*PermissionExpr `yaml:"exclusion,omitempty"`
}

func Load() (*Schema, error) {
	data, err := os.ReadFile("schema.yaml")
	if err != nil {
		return nil, err
	}

	var schema Schema
	if err := yaml.Unmarshal(data, &schema); err != nil {
		return nil, err
	}

	return &schema, nil
}
