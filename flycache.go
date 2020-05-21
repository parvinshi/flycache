package main

import (
	"flycache/lru"
	"flycache/singleflight"
	"fmt"
	"log"
	"sync"
)

var (
	mutex  sync.RWMutex
	groups = make(map[string]*Group)
)

type Group struct {
	name      string
	getter    Getter
	mainCache LRUCache
	peers     PeerPicker
	loader    *singleflight.Group
}

type Getter interface {
	Get(key string) ([]byte, error)
}

type GetterFunc func(key string) ([]byte, error)

func (f GetterFunc) Get(key string) ([]byte, error) {
	return f(key)
}

func (g *Group) RegisterPeers(peers PeerPicker) {
	if g.peers != nil {
		panic("RegisterPeerPicker called more than once")
	}
	g.peers = peers
}

func NewGroup(name string, cacheBytes int64, getter Getter) *Group {
	if getter == nil {
		panic("nil getter")
	}

	mutex.Lock()
	defer mutex.Unlock()

	g := &Group{
		name:      name,
		getter:    getter,
		mainCache: LRUCache{cacheBytes: cacheBytes},
		loader:    &singleflight.Group{},
	}

	groups[name] = g

	return g
}

func GetGroup(name string) *Group {
	mutex.RLock()
	g := groups[name]
	mutex.RUnlock()

	return g
}

func (g *Group) Get(key string) (ByteView, error) {
	if key == "" {
		return ByteView{}, fmt.Errorf("key is required")
	}

	if v, ok := g.mainCache.get(key); ok {
		log.Println("cache hit")
		return v, nil
	}

	return g.load(key)
}

func (g *Group) load(key string) (val ByteView, err error) {
	v, err := g.loader.Do(key, func() (i interface{}, e error) {
		if g.peers != nil {
			if peer, ok := g.peers.PickPeer(key); ok {
				if val, err = g.getFromPeer(peer, key); err == nil {
					return val, nil
				}
				log.Println("Failed to get from peer", err)
			}
		}

		return g.getLocally(key)
	})

	if err == nil {
		return v.(ByteView), nil
	}

	return
}

func (g *Group) getFromPeer(peer PeerGetter, key string) (ByteView, error) {
	bytes, err := peer.Get(g.name, key)
	if err != nil {
		return ByteView{}, err
	}
	return ByteView{b: bytes}, nil
}

func (g *Group) getLocally(key string) (ByteView, error) {
	bytes, err := g.getter.Get(key)
	if err != nil {
		return ByteView{}, err
	}

	val := ByteView{b: cloneBytes(bytes)}
	g.populateCache(key, val)
	return val, nil
}

func (g *Group) populateCache(key string, val ByteView) {
	g.mainCache.set(key, val)
}

func (g *Group) GetStats() lru.Stats {
	return g.mainCache.stats()
}
