package schema

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

func NewSchema() (*Schema, error) {
	filePaths, err := getYamlFilesFromEnv()
	if err != nil {
		return nil, err
	}

	schema := Schema{Namespaces: map[string]*Namespace{}}
	for _, path := range filePaths {
		tmpSchema := Schema{}
		f, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		err = yaml.Unmarshal(f, &tmpSchema)
		if err != nil {
			return nil, err
		}
		tmpSchema.Build()

		if err := schema.Union(&tmpSchema); err != nil {
			return nil, err
		}
	}

	return &schema, nil
}

func getYamlFilesFromEnv() ([]string, error) {
	pathStr := os.Getenv("SCHEMA_PATH")
	if pathStr == "" {
		return nil, fmt.Errorf("CONFIG_PATHS not set")
	}

	paths := strings.Split(pathStr, ",")
	var yamlFiles []string

	for _, p := range paths {
		p = strings.TrimSpace(p)
		fi, err := os.Stat(p)
		if err != nil {
			return nil, err
		}

		if fi.IsDir() {
			// scan all .yaml files in directory
			err := filepath.WalkDir(p, func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if !d.IsDir() && (strings.HasSuffix(d.Name(), ".yaml") || strings.HasSuffix(d.Name(), ".yml")) {
					yamlFiles = append(yamlFiles, path)
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
		} else {
			yamlFiles = append(yamlFiles, p)
		}
	}

	return yamlFiles, nil
}

type Schema struct {
	Namespaces map[string]*Namespace `yaml:"namespaces"`
}

type Namespace struct {
	AllowSubjectNamespaces []string             `yaml:"allow_subject_namespaces,omitempty"`
	Type                   string               `yaml:"type"`
	Relations              map[string]*Relation `yaml:"relations"`
}

type Relation struct {
	AllowSubjectNamespaces []string `yaml:"allow_subject_namespaces,omitempty"`
	UsersetRewrite         `yaml:",inline"`
}

// Recursive AST-like structure
type UsersetRewrite struct {
	Union           []*UsersetRewrite `yaml:"union,omitempty"`
	Intersection    []*UsersetRewrite `yaml:"intersection,omitempty"`
	Exclusion       *ExclusionNode    `yaml:"exclusion,omitempty"`
	ComputedUserSet *ComputedUserset  `yaml:"computed_userset,omitempty"`
	TupleToUserset  *TupleToUserset   `yaml:"tuple_to_userset,omitempty"`
}

type Userset struct {
	Namespace *string `yaml:"namespace"`
	Id        string  `yaml:"id"`
	Relation  string  `yaml:"relation"` // e.g. "owner""
}

type ExclusionNode struct {
	Base     *UsersetRewrite `yaml:"base"`
	Subtract *UsersetRewrite `yaml:"subtract"`
}

type ComputedUserset struct {
	Relation string `yaml:"relation"`
}

type TupleToUserset struct {
	Tupleset        *Userset         `yaml:"tupleset,omitempty"`
	ComputedUserset *ComputedUserset `yaml:"computed_userset"`
}

func (t *TupleToUserset) Build(rel string) {
	if t.Tupleset == nil {
		t.Tupleset = &Userset{Relation: rel}
	}
}

func (u *UsersetRewrite) Build(rel string) {
	if u.TupleToUserset != nil {
		u.TupleToUserset.Build(rel)
	}
}

func (r *UsersetRewrite) Validate() error {
	count := 0
	if r.Union != nil {
		count++
		for _, child := range r.Union {
			if err := child.Validate(); err != nil {
				return err
			}
		}
	}
	if r.Intersection != nil {
		count++
		for _, child := range r.Intersection {
			if err := child.Validate(); err != nil {
				return err
			}
		}
	}
	if r.Exclusion != nil {
		count++
		if r.Exclusion.Base == nil || r.Exclusion.Subtract == nil {
			return errors.New("exclusion must have both base and subtract")
		}
		if err := r.Exclusion.Base.Validate(); err != nil {
			return err
		}
		if err := r.Exclusion.Subtract.Validate(); err != nil {
			return err
		}
	}
	if r.ComputedUserSet != nil {
		count++
		if r.ComputedUserSet.Relation == "" {
			return errors.New("computed_userset must have relation")
		}
	}
	if r.TupleToUserset != nil {
		count++
		if r.TupleToUserset.Tupleset == nil || r.TupleToUserset.ComputedUserset == nil {
			return errors.New("tuple_to_userset must have both tupleset and computed_userset")
		}
		if r.TupleToUserset.Tupleset.Relation == "" || r.TupleToUserset.ComputedUserset.Relation == "" {
			return errors.New("tuple_to_userset fields must have non-empty relation")
		}
	}
	if count > 1 {
		return errors.New("only one rewrite type can be set in UsersetRewrite")
	}
	if count == 0 {
		return errors.New("at least one rewrite type must be set in UsersetRewrite")
	}
	return nil
}

func (s *Schema) Build() {
	for _, ns := range s.Namespaces {
		for name, rel := range ns.Relations {
			if rel.TupleToUserset != nil {
				rel.TupleToUserset.Build(name)
			}

			if rel.Union != nil {
				for _, rewrite := range rel.Union {
					rewrite.Build(name)
				}
			}

			if rel.Intersection != nil {
				for _, rewrite := range rel.Intersection {
					rewrite.Build(name)
				}
			}
		}
	}
}

// TODO: union deeper
func (s *Schema) Union(schema2 *Schema) error {
	for ns, nsEntry := range schema2.Namespaces {
		if _, exists := s.Namespaces[ns]; exists {
			return fmt.Errorf("namespace %s already exists", ns)
		}
		s.Namespaces[ns] = nsEntry
	}

	return nil
}
