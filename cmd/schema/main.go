package main

import (
	"os"

	"authz/internal/schema"

	"github.com/k0kubun/pp/v3"
	"gopkg.in/yaml.v3"
)

func main() {
	s := schema.Schema{}

	f, _ := os.ReadFile("schema.yaml")
	err := yaml.Unmarshal(f, &s)
	if err != nil {
		panic(err)
	}

	mypp := pp.New()
	mypp.SetOmitEmpty(true)
	mypp.Print(&s)
}
