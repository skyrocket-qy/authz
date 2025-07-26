package pkg

import "authz/docs/openapi"

func InitSwagger() {
	openapi.SwaggerInfo.Title = "OpenAPI"
	openapi.SwaggerInfo.Schemes = []string{"http"}
}
