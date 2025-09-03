package schema_test

import (
	"os"
	"path/filepath"
	"testing"

	"authz/internal/schema"
	"github.com/stretchr/testify/assert"
)

func TestGetYamlFilesFromEnv(t *testing.T) {
	t.Run("should return yaml files from a directory", func(t *testing.T) {
		tempDir := "test_schemas_get"
		err := os.Mkdir(tempDir, 0o755)
		assert.NoError(t, err)

		defer os.RemoveAll(tempDir)

		// Create test yaml files
		file1, err := os.Create(filepath.Join(tempDir, "test1.yaml"))
		assert.NoError(t, err)
		file1.Close()

		file2, err := os.Create(filepath.Join(tempDir, "test2.yml"))
		assert.NoError(t, err)
		file2.Close()

		// Set environment variable
		t.Setenv("SCHEMA_PATH", tempDir)

		files, err := schema.GetYamlFilesFromEnv()
		assert.NoError(t, err)
		assert.Len(t, files, 2)
	})

	t.Run("should return error if SCHEMA_PATH is not set", func(t *testing.T) {
		t.Setenv("SCHEMA_PATH", "")
		_, err := schema.GetYamlFilesFromEnv()
		assert.Error(t, err)
	})

	t.Run("should return error if SCHEMA_PATH is an absolute path", func(t *testing.T) {
		t.Setenv("SCHEMA_PATH", "/etc/passwd")
		_, err := schema.GetYamlFilesFromEnv()
		assert.Error(t, err)
	})

	t.Run("should return error if path does not exist", func(t *testing.T) {
		t.Setenv("SCHEMA_PATH", "non_existent_dir")
		_, err := schema.GetYamlFilesFromEnv()
		assert.Error(t, err)
	})
}

func TestNewSchema(t *testing.T) {
	tempDir := "test_schemas_new"
	err := os.Mkdir(tempDir, 0o755)
	assert.NoError(t, err)

	defer os.RemoveAll(tempDir)

	// Create test yaml files
	schema1 := `
namespaces:
  user:
    type: subject
  group:
    type: subject
    relations:
      member: {}
`
	err = os.WriteFile(filepath.Join(tempDir, "schema1.yaml"), []byte(schema1), 0o644)
	assert.NoError(t, err)

	schema2 := `
namespaces:
  document:
    type: resource
    relations:
      owner: {}
      viewer:
        union:
          - computed_userset:
              relation: owner
`
	err = os.WriteFile(filepath.Join(tempDir, "schema2.yaml"), []byte(schema2), 0o644)
	assert.NoError(t, err)

	// Set environment variable
	t.Setenv("SCHEMA_PATH", tempDir)

	s, err := schema.NewSchema()
	assert.NoError(t, err)
	assert.NotNil(t, s)
	assert.Len(t, s.Namespaces, 3)
	assert.Contains(t, s.Namespaces, "user")
	assert.Contains(t, s.Namespaces, "group")
	assert.Contains(t, s.Namespaces, "document")
}

func TestNewSchema_DuplicateNamespace(t *testing.T) {
	tempDir := "test_schemas_dup"
	err := os.Mkdir(tempDir, 0o755)
	assert.NoError(t, err)

	defer os.RemoveAll(tempDir)

	schema1 := `
namespaces:
  user:
    type: subject
`
	err = os.WriteFile(filepath.Join(tempDir, "schema1.yaml"), []byte(schema1), 0o644)
	assert.NoError(t, err)

	schema2 := `
namespaces:
  user:
    type: subject
`
	err = os.WriteFile(filepath.Join(tempDir, "schema2.yaml"), []byte(schema2), 0o644)
	assert.NoError(t, err)

	t.Setenv("SCHEMA_PATH", tempDir)

	_, err = schema.NewSchema()
	assert.Error(t, err)
}

func TestSchema_Build(t *testing.T) {
	s := &schema.Schema{
		Namespaces: map[string]*schema.Namespace{
			"document": {
				Relations: map[string]*schema.Relation{
					"viewer": {
						UsersetRewrite: schema.UsersetRewrite{
							TupleToUserset: &schema.TupleToUserset{
								ComputedUserset: &schema.ComputedUserset{
									Relation: "owner",
								},
							},
						},
					},
				},
			},
		},
	}

	s.Build()

	assert.NotNil(t, s.Namespaces["document"].Relations["viewer"].TupleToUserset.Tupleset)
	assert.Equal(t, "viewer", *s.Namespaces["document"].Relations["viewer"].TupleToUserset.Tupleset.Relation)
}

func TestUsersetRewrite_Validate(t *testing.T) {
	t.Run("should return error if more than one rewrite type is set", func(t *testing.T) {
		rewrite := &schema.UsersetRewrite{
			Union:        []*schema.UsersetRewrite{{}},
			Intersection: []*schema.UsersetRewrite{{}},
		}
		err := rewrite.Validate()
		assert.Error(t, err)
	})

	t.Run("should return error if no rewrite type is set", func(t *testing.T) {
		rewrite := &schema.UsersetRewrite{}
		err := rewrite.Validate()
		assert.Error(t, err)
	})

	t.Run("should return error if exclusion base is nil", func(t *testing.T) {
		rewrite := &schema.UsersetRewrite{
			Exclusion: &schema.ExclusionNode{
				Subtract: &schema.UsersetRewrite{
					ComputedUserSet: &schema.ComputedUserset{Relation: "test"},
				},
			},
		}
		err := rewrite.Validate()
		assert.Error(t, err)
	})

	t.Run("should return error if exclusion subtract is nil", func(t *testing.T) {
		rewrite := &schema.UsersetRewrite{
			Exclusion: &schema.ExclusionNode{
				Base: &schema.UsersetRewrite{
					ComputedUserSet: &schema.ComputedUserset{Relation: "test"},
				},
			},
		}
		err := rewrite.Validate()
		assert.Error(t, err)
	})

	t.Run("should return error if computed_userset relation is empty", func(t *testing.T) {
		rewrite := &schema.UsersetRewrite{
			ComputedUserSet: &schema.ComputedUserset{},
		}
		err := rewrite.Validate()
		assert.Error(t, err)
	})

	t.Run("should return error if tuple_to_userset tupleset is nil", func(t *testing.T) {
		rewrite := &schema.UsersetRewrite{
			TupleToUserset: &schema.TupleToUserset{
				ComputedUserset: &schema.ComputedUserset{Relation: "test"},
			},
		}
		err := rewrite.Validate()
		assert.Error(t, err)
	})

	t.Run("should return error if tuple_to_userset computed_userset is nil", func(t *testing.T) {
		relation := "test"
		rewrite := &schema.UsersetRewrite{
			TupleToUserset: &schema.TupleToUserset{
				Tupleset: &schema.Userset{Relation: &relation},
			},
		}
		err := rewrite.Validate()
		assert.Error(t, err)
	})

	t.Run("should return error if tuple_to_userset tupleset relation is nil", func(t *testing.T) {
		rewrite := &schema.UsersetRewrite{
			TupleToUserset: &schema.TupleToUserset{
				Tupleset:        &schema.Userset{},
				ComputedUserset: &schema.ComputedUserset{Relation: "test"},
			},
		}
		err := rewrite.Validate()
		assert.Error(t, err)
	})

	t.Run("should return error if tuple_to_userset computed_userset relation is empty", func(t *testing.T) {
		relation := "test"
		rewrite := &schema.UsersetRewrite{
			TupleToUserset: &schema.TupleToUserset{
				Tupleset:        &schema.Userset{Relation: &relation},
				ComputedUserset: &schema.ComputedUserset{},
			},
		}
		err := rewrite.Validate()
		assert.Error(t, err)
	})

	t.Run("should return nil if valid", func(t *testing.T) {
		rewrite := &schema.UsersetRewrite{
			ComputedUserSet: &schema.ComputedUserset{Relation: "test"},
		}
		err := rewrite.Validate()
		assert.NoError(t, err)
	})
}

func TestNewSchema_InvalidYaml(t *testing.T) {
	tempDir := "test_schemas_invalid"
	err := os.Mkdir(tempDir, 0o755)
	assert.NoError(t, err)

	defer os.RemoveAll(tempDir)

	invalidSchema := `
namespaces:
  user:
    type: subject
  group
`
	err = os.WriteFile(filepath.Join(tempDir, "invalid.yaml"), []byte(invalidSchema), 0o644)
	assert.NoError(t, err)

	t.Setenv("SCHEMA_PATH", tempDir)

	_, err = schema.NewSchema()
	assert.Error(t, err)
}
