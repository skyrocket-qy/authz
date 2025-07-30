package schema

import "errors"

type Schema struct {
	Namespaces map[string]*Namespace `yaml:"namespaces"`
}

type Namespace struct {
	Relations map[string]*Relation `yaml:"relations"`
}

type Relation = UsersetRewrite

// Recursive AST-like structure
type UsersetRewrite struct {
	Union           []*UsersetRewrite `yaml:"union,omitempty"`
	Intersection    []*UsersetRewrite `yaml:"intersection,omitempty"`
	Exclusion       *ExclusionNode    `yaml:"exclusion,omitempty"`
	ComputedUserSet *ComputedUserset  `yaml:"computed_userset,omitempty"`
	TupleToUserset  *TupleToUserset   `yaml:"tuple_to_userset,omitempty"`
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

type Userset struct {
	Relation string `yaml:"relation"` // e.g. "owner""
}

type ExclusionNode struct {
	Base     *UsersetRewrite `yaml:"base"`
	Subtract *UsersetRewrite `yaml:"subtract"`
}

type ComputedUserset struct {
	Relation string `yaml:"relation"`
}

type TupleToUserset struct {
	Tupleset        *Userset         `yaml:"tupleset"`         // e.g. parent
	ComputedUserset *ComputedUserset `yaml:"computed_userset"` // e.g. viewer
}
