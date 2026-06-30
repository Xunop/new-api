package playground_attachment

import (
	"github.com/QuantumNous/new-api/service/playground_attachment/driver"
	"github.com/QuantumNous/new-api/service/playground_attachment/driver/local"
)

func NewLocalDriver(basePath string) driver.Driver {
	return local.New(basePath)
}
