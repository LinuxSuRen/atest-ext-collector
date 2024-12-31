package gateway

import (
	"gopkg.in/yaml.v3"
	"os"
)

func ParseGateway(config string) (gw *Gateway, err error) {
	gw = &Gateway{}
	var data []byte
	if data, err = os.ReadFile(config); err == nil {
		err = yaml.Unmarshal(data, gw)
	}
	return
}
