package main

import (
	_ "ariga.io/atlas-provider-gorm/gormschema"
	"authz/cmd"
)

func main() {
	cmd.Execute()
}
