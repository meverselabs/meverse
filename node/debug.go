package node

import "github.com/fletaio/fleta_v2/cmd/config/go-toml"

var (
	DEBUG = false
)

func init() {
	defer func() {
		recover()
	}()
	config, err := toml.LoadFile("config.toml")
	if err == nil {
		debug := config.Get("debug").(string)
		DEBUG = debug == "true"
	}
}
