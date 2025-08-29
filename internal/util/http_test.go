package util_test

import (
	"os"
	"path/filepath"
	"testing"

	"authz/internal/util"
	"github.com/skyrocket-qy/erx"
	"github.com/stretchr/testify/assert"
)

func TestTrimToProject(t *testing.T) {
	projectRoot, err := os.Getwd()
	assert.NoError(t, err)

	t.Run("should trim project root prefix", func(t *testing.T) {
		path := filepath.Join(projectRoot, "internal", "logic", "logic.go")
		expected := filepath.Join("/", "internal", "logic", "logic.go")
		assert.Equal(t, expected, util.TrimToProject(path))
	})

	t.Run("should return original path if not in project", func(t *testing.T) {
		path := "/some/other/project/main.go"
		assert.Equal(t, path, util.TrimToProject(path))
	})

	t.Run("should handle path equal to project root", func(t *testing.T) {
		assert.Empty(t, util.TrimToProject(projectRoot))
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
			assert.Equal(t, tc.expected, util.ExtractFuncName(tc.input))
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
		filtered := util.FilterCallerInfos(infos)
		assert.Equal(t, infos, filtered)
	})

	t.Run("should stop filtering after first non-project file", func(t *testing.T) {
		infos := []erx.CallerInfo{infoInProject1, infoOutsideProject, infoInProject2}
		filtered := util.FilterCallerInfos(infos)
		assert.Len(t, filtered, 1)
		assert.Equal(t, infoInProject1, filtered[0])
	})

	t.Run("should return empty slice if first file is outside project", func(t *testing.T) {
		infos := []erx.CallerInfo{infoOutsideProject, infoInProject1}
		filtered := util.FilterCallerInfos(infos)
		assert.Empty(t, filtered)
	})

	t.Run("should handle empty input slice", func(t *testing.T) {
		infos := []erx.CallerInfo{}
		filtered := util.FilterCallerInfos(infos)
		assert.Empty(t, filtered)
	})

	t.Run("should handle nil input slice", func(t *testing.T) {
		filtered := util.FilterCallerInfos(nil)
		assert.Empty(t, filtered)
	})
}
