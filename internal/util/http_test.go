package util

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/skyrocket-qy/erx"
	"github.com/stretchr/testify/assert"
)

func TestTrimToProject(t *testing.T) {
	projectRoot, err := os.Getwd()
	assert.NoError(t, err)

	t.Run("should trim project root prefix", func(t *testing.T) {
		path := filepath.Join(projectRoot, "internal", "logic", "logic.go")
		expected := filepath.Join("/", "internal", "logic", "logic.go")
		assert.Equal(t, expected, trimToProject(path))
	})

	t.Run("should return original path if not in project", func(t *testing.T) {
		path := "/some/other/project/main.go"
		assert.Equal(t, path, trimToProject(path))
	})

	t.Run("should handle path equal to project root", func(t *testing.T) {
		assert.Equal(t, "", trimToProject(projectRoot))
	})
}

func TestExtractFuncName(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "fully qualified function name",
			input:    "srv/internal/logic/inter.(*Logic).Login",
			expected: "inter.(*Logic).Login",
		},
		{
			name:     "function name without slashes",
			input:    "main.main",
			expected: "main.main",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "name with single part",
			input:    "myFunction",
			expected: "myFunction",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, extractFuncName(tc.input))
		})
	}
}

func TestFilterCallerInfos(t *testing.T) {
	projectRoot, err := os.Getwd()
	assert.NoError(t, err)

	infoInProject1 := erx.CallerInfo{File: filepath.Join(projectRoot, "main.go")}
	infoInProject2 := erx.CallerInfo{File: filepath.Join(projectRoot, "internal", "util.go")}
	infoOutsideProject := erx.CallerInfo{File: "/usr/lib/go/src/runtime/panic.go"}

	t.Run("should return only infos within the project", func(t *testing.T) {
		infos := []erx.CallerInfo{infoInProject1, infoInProject2}
		filtered := filterCallerInfos(infos)
		assert.Equal(t, infos, filtered)
	})

	t.Run("should stop filtering after first non-project file", func(t *testing.T) {
		infos := []erx.CallerInfo{infoInProject1, infoOutsideProject, infoInProject2}
		filtered := filterCallerInfos(infos)
		assert.Len(t, filtered, 1)
		assert.Equal(t, infoInProject1, filtered[0])
	})

	t.Run("should return empty slice if first file is outside project", func(t *testing.T) {
		infos := []erx.CallerInfo{infoOutsideProject, infoInProject1}
		filtered := filterCallerInfos(infos)
		assert.Empty(t, filtered)
	})

	t.Run("should handle empty input slice", func(t *testing.T) {
		infos := []erx.CallerInfo{}
		filtered := filterCallerInfos(infos)
		assert.Empty(t, filtered)
	})

	t.Run("should handle nil input slice", func(t *testing.T) {
		filtered := filterCallerInfos(nil)
		assert.Empty(t, filtered)
	})
}
