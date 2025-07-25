package main

import (
	"authz/cmd"
	_ "authz/docs/openapi"

	_ "ariga.io/atlas-provider-gorm/gormschema"
)

// @securityDefinitions.apikey	Bearer
// @in							header
// @name						Authorization.
func main() {
	cmd.Execute()
}
