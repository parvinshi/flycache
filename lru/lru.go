package lru

import (
	"container/list"
	"errors"
)

type Cache struct {
	MaxAllow   int64
    Stats

	ll         *list.List
	items      map[string]*list.Element

	OnEvicted  func(key string, val Value)
}

type Stats struct {
	usedBytes  int64
}

type entry struct {
	key string
	val Value
}

type String string

type Value interface {
	Len() int
}

// New constructs an LRU of the given max allowed bytes
func New(maxAllow int64, onEvicted func(string, Value)) (*Cache, error) {
    if maxAllow < 0 {
    	return nil, errors.New("the maximum allowed value must be greater than or equal to 0")
	}

	c := &Cache{
		MaxAllow:  maxAllow,
		ll:        list.New(),
		items:     make(map[string]*list.Element),
		OnEvicted: onEvicted,
	}

	return c, nil
}

//Get looks up key's value from the cache
func (c *Cache) Get(key string) (val Value, ok bool) {
	if el, ok := c.items[key]; ok {
		c.ll.MoveToFront(el)
		e := el.Value.(*entry)

		return e.val, true
	}

	return nil, false
}

// Set is adds a value to the cache
func (c *Cache) Set(key string, val Value) {
	if el, ok := c.items[key]; ok {
		c.ll.MoveToFront(el)
		e := el.Value.(*entry)
		c.usedBytes += int64(val.Len()) - int64(e.val.Len())
		e.val = val
	} else {
		el := c.ll.PushFront(&entry{key, val})
		c.items[key] = el
		c.usedBytes += int64(len(key)) + int64(val.Len())
	}

	for c.MaxAllow > 0 && c.usedBytes > c.MaxAllow {
		c.RemoveOldest()
	}
}

//Del removes the provided key from the cache
func (c *Cache) Del(key string) (ok bool) {
	if el, ok := c.items[key]; ok {
		c.removeElement(el)

		e := el.Value.(*entry)
		c.updateUsedBytes(e)

		return true
	}

	return false
}

//RemoveOldest removes the oldest item from the cache
func (c *Cache) RemoveOldest() (key, value interface{}, ok bool) {
	el := c.ll.Back()
	if el != nil {
		c.removeElement(el)

		e := el.Value.(*entry)
		c.updateUsedBytes(e)

		return e.key, e.val, true
	}

	return nil, nil, false
}

// removeElement is used to remove a given list element from the cache
func (c *Cache) removeElement(el *list.Element) {
	c.ll.Remove(el)
	e := el.Value.(*entry)
	delete(c.items, e.key)

	if c.OnEvicted != nil {
		c.OnEvicted(e.key, e.val)
	}
}

func (c *Cache) updateUsedBytes(el *entry) {
	oldLen := len(el.key) + el.val.Len()
	c.usedBytes -= int64(oldLen)
}

//GetOldest returns the oldest entry
func (c *Cache) GetOldest() (key, value interface{}, ok bool) {
	el := c.ll.Back()
	if el != nil {
		e := el.Value.(*entry)
		return e.key, e.val, true
	}

	return nil, nil, false
}

//Purge is used to completely clear the cache
func (c *Cache) Purge() {
	for k, e := range c.items {
		if c.OnEvicted != nil {
			e := e.Value.(*entry)
			c.OnEvicted(e.key, e.val)
		}

		delete(c.items, k)
	}

	c.ll.Init()
}

func (c *Cache) Contains(key string) (ok bool) {
	_, ok = c.items[key]
	return ok
}

//Keys returns a slice of the keys in the cache
func (c *Cache) Keys() []string {
	keys := make([]string, len(c.items))
	i := 0
	for e := c.ll.Back(); e != nil; e = e.Prev() {
		keys[i] = e.Value.(*entry).key
		i++
	}

	return keys
}

func (c *Cache) Equal(x, y []string) bool {
	if len(x) != len(y) {
		return false
	}

	for i := range x {
		if x[i] != y[i] {
			return false
		}
	}

	return true
}

func (s String) Len() int {
	return len(s)
}
