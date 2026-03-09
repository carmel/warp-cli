//go:generate openapi-generator generate -i openapi-spec.yml -g go -o openapi --additional-properties=disallowAdditionalPropertiesIfNotPresent=false

package main

import (
	"log"

	"github.com/carmel/warp-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		log.Fatalf("%+v\n", err)
	}
}
