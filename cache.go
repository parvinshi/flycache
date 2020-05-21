package main

import (
	"flycache/lru"
	"sync"
)

type LRUCache struct {
	lru        *lru.Cache
	mutex      sync.RWMutex
	cacheBytes int64
}

func (c *LRUCache) set(k string, v ByteView) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.lru == nil {
		c.lru, _ = lru.New(c.cacheBytes, nil)
	}

	c.lru.Set(k, v)
}

func (c *LRUCache) get(k string) (value ByteView, ok bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if c.lru == nil {
		return
	}

	if v, ok := c.lru.Get(k); ok {
		return v.(ByteView), ok
	}

	return
}

func (c *LRUCache) del(k string) (ok bool) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if ok := c.lru.Del(k); ok {
		return true
	}

	return false
}

//get cache stats
func (c *LRUCache) stats() lru.Stats {
	return c.lru.Stats
}
