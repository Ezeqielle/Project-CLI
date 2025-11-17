package plugins

import (
	"github.com/ezeqielle/pcli/internal/projecttype"
	goproject "github.com/ezeqielle/pcli/internal/projecttype/go"
)

func RegisterAll() {
	projecttype.Register(goproject.New())
	// later: register typescript, terraform, ...
}
