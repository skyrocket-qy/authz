package util

import (
	"os"
	"strings"

	"github.com/skyrocket-qy/erx"
)

type ErrResp struct {
	ReqId string `json:"reqId"`
	Code  string `json:"code"`
}

func TrimToProject(path string) string {
	projectRoot, _ := os.Getwd()
	if rel, ok := strings.CutPrefix(path, projectRoot); ok {
		return rel
	}

	return path
}

func ExtractFuncName(fullFunc string) string {
	// e.g., input: srv/internal/logic/inter.(*Logic).Login
	// output: (*Logic).Login
	if idx := strings.LastIndex(fullFunc, "/"); idx >= 0 {
		return fullFunc[idx+1:]
	}

	return fullFunc
}

func FilterCallerInfos(infos []erx.CallerInfo) []erx.CallerInfo {
	projectPrefix, _ := os.Getwd()

	var filtered []erx.CallerInfo

	for _, ci := range infos {
		if strings.HasPrefix(ci.File, projectPrefix) {
			filtered = append(filtered, ci)
		} else {
			break
		}
	}

	return filtered
}
