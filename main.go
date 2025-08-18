package main

import (
	"authz/cmd"

	_ "ariga.io/atlas-provider-gorm/gormschema"
)

func main() {
	cmd.Execute()
}
