package schema_test

import (
	"os"
	"path/filepath"
	"testing"

	"authz/internal/schema"
	"github.com/stretchr/testify/assert"
)

func TestGetYamlFilesFromEnv(t *testing.T) {
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
