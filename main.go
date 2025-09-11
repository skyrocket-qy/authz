package main

import (
	"authz/cmd"
	_ "net/http/pprof"

	_ "ariga.io/atlas-provider-gorm/gormschema"
)

func main() {
	cmd.Execute()
}
