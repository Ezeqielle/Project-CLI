package postplugin

import (
	global "github.com/ezeqielle/pcli/internal/postplugin/global"
)

func RegisterAll() {
	Register(global.New())
}
