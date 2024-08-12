package main

import (
	"github.com/BurntSushi/toml"
	"github.com/gopherd/core/service"

	_ "github.com/gopherd/example/components"
)

func main() {
	service.Run(service.WithEncoder(toml.Marshal), service.WithDecoder(toml.Unmarshal))
}
