package postplugin

import (
	"sync"
)

var (
	mu      sync.RWMutex
	plugins []Plugin
)

func Register(p Plugin) {
	mu.Lock()
	defer mu.Unlock()
	plugins = append(plugins, p)
}

func All() []Plugin {
	mu.RLock()
	defer mu.RUnlock()

	out := make([]Plugin, len(plugins))
	copy(out, plugins)
	return out
}
