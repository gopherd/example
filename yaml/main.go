package main

import (
	"github.com/gopherd/core/service"
	"gopkg.in/yaml.v3"

	_ "github.com/gopherd/example/components"
)

func main() {
	service.Run(service.WithEncoder(yaml.Marshal), service.WithDecoder(yaml.Unmarshal))
}
