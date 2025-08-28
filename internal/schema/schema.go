package schema

import (
	"os"
	"path/filepath"
	"strings"

	"authz/internal/util"

	"github.com/rs/zerolog/log"
	"github.com/skyrocket-qy/erx"
	"gopkg.in/yaml.v3"
)

func NewSchema() (*Schema, error) {
	filePaths, err := getYamlFilesFromEnv()
	if err != nil {
		return nil, erx.W(err)
	}

	log.Info().Msgf("schema files: %v", filePaths)

	schema := Schema{Namespaces: map[string]*Namespace{}}
	for _, path := range filePaths {
		tmpSchema := Schema{}

		clean := filepath.Clean(path)

		f, err := os.ReadFile(clean)
		if err != nil {
			return nil, erx.W(err)
		}

		err = yaml.Unmarshal(f, &tmpSchema)
		if err != nil {
			return nil, erx.W(err)
		}

		tmpSchema.Build()

		if err := schema.Union(&tmpSchema); err != nil {
			return nil, erx.W(err)
		}
	}

	return &schema, nil
}

func getYamlFilesFromEnv() ([]string, error) {
	pathStr := os.Getenv("SCHEMA_PATH")
	if strings.HasPrefix(pathStr, "/") {
		return nil, erx.Newf(util.ErrBadRequest, "schema path %s is not allowed", pathStr)
	}

	if pathStr == "" {
		return nil, erx.New(util.ErrBadRequest, "CONFIG_PATHS not set")
	}

	paths := strings.Split(pathStr, ",")

	var yamlFiles []string

	for _, p := range paths {
		p = strings.TrimSpace(p)

		fi, err := os.Stat(p)
		if err != nil {
			return nil, erx.W(err)
		}

		if fi.IsDir() {
			// only scan files directly in directory (non-recursive)
			entries, err := os.ReadDir(p)
			if err != nil {
				return nil, erx.W(err)
			}

			for _, entry := range entries {
				if !entry.IsDir() &&
					(strings.HasSuffix(entry.Name(), ".yaml") || strings.HasSuffix(entry.Name(), ".yml")) {
					yamlFiles = append(yamlFiles, filepath.Join(p, entry.Name()))
				}
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
	UsersetRewrite `yaml:"inline"`

	AllowSubjectNamespaces []string `yaml:"allow_subject_namespaces,omitempty"`
}

// Recursive AST-like structure.
type UsersetRewrite struct {
	Union           []*UsersetRewrite `yaml:"union,omitempty"`
	Intersection    []*UsersetRewrite `yaml:"intersection,omitempty"`
	Exclusion       *ExclusionNode    `yaml:"exclusion,omitempty"`
	ComputedUserSet *ComputedUserset  `yaml:"computed_userset,omitempty"`
	TupleToUserset  *TupleToUserset   `yaml:"tuple_to_userset,omitempty"`
}

type Userset struct {
	Namespace *string `yaml:"namespace,omitempty"`
	Id        *string `yaml:"id,omitempty"`
	Relation  *string `yaml:"relation,omitempty"` // e.g. "owner""
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
		t.Tupleset = &Userset{Relation: &rel}
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
				return erx.W(err)
			}
		}
	}

	if r.Intersection != nil {
		count++

		for _, child := range r.Intersection {
			if err := child.Validate(); err != nil {
				return erx.W(err)
			}
		}
	}

	if r.Exclusion != nil {
		count++

		if r.Exclusion.Base == nil || r.Exclusion.Subtract == nil {
			return erx.New(util.ErrBadRequest, "exclusion must have both base and subtract")
		}

		if err := r.Exclusion.Base.Validate(); err != nil {
			return erx.W(err)
		}

		if err := r.Exclusion.Subtract.Validate(); err != nil {
			return erx.W(err)
		}
	}

	if r.ComputedUserSet != nil {
		count++

		if r.ComputedUserSet.Relation == "" {
			return erx.New(util.ErrBadRequest, "computed_userset must have relation")
		}
	}

	if r.TupleToUserset != nil {
		count++

		if r.TupleToUserset.Tupleset == nil || r.TupleToUserset.ComputedUserset == nil {
			return erx.New(util.ErrBadRequest, "tuple_to_userset must have both tupleset and computed_userset")
		}

		if r.TupleToUserset.Tupleset.Relation == nil ||
			r.TupleToUserset.ComputedUserset.Relation == "" {
			return erx.New(util.ErrBadRequest, "tuple_to_userset fields must have non-empty relation")
		}
	}

	if count > 1 {
		return erx.New(util.ErrBadRequest, "only one rewrite type can be set in UsersetRewrite")
	}

	if count == 0 {
		return erx.New(util.ErrBadRequest, "at least one rewrite type must be set in UsersetRewrite")
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

// TODO: union deeper.
func (s *Schema) Union(schema2 *Schema) error {
	for ns, nsEntry := range schema2.Namespaces {
		if _, exists := s.Namespaces[ns]; exists {
			return erx.Newf(util.ErrDuplicate, "namespace %s already exists", ns)
		}

		s.Namespaces[ns] = nsEntry
	}

	return nil
}
