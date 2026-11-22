package image

import (
	"image"
	"sync"
)

type Cache struct {
	once  sync.Once
	value image.Image
}

var (
	cacheMu  sync.RWMutex
	cacheMap = make(map[string]*Cache)
)

func cacheOrFetch(key string, fetch func() image.Image) image.Image {
	cacheMu.RLock()
	e, ok := cacheMap[key]
	cacheMu.RUnlock()

	if !ok {
		cacheMu.Lock()
		e, ok = cacheMap[key]
		if !ok {
			e = &Cache{}
			cacheMap[key] = e
		}
		cacheMu.Unlock()
	}

	e.once.Do(func() {
		e.value = fetch()
	})

	return e.value
}
