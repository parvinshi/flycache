package chash

import (
	"hash/crc32"
	"sort"
	"strconv"
)

type Hash func(data []byte) uint32

type Map struct {
	hash     Hash
	replicas int
	keys     []int
	hashMap  map[int]string
}

func New(replicas int, fn Hash) *Map {
	m := &Map{
		replicas: replicas, // 每个真实节点和虚拟节点的个数
		hash:     fn,
		hashMap:  make(map[int]string),
	}
	if m.hash == nil {
		m.hash = crc32.ChecksumIEEE
	}
	return m
}

// Add adds some keys to the hash.
func (m *Map) Add(keys ...string) {
	for _, key := range keys {
		for i := 0; i < m.replicas; i++ {
			str := strconv.Itoa(i) + key
			data := []byte(str)
			hash := int(m.hash(data))
			//log.Printf("str: %#v\n", str) // "06"
			//log.Printf("data: %#v\n", []byte(strconv.Itoa(i) + key)) // []byte{0x30, 0x36}
			//log.Printf("hash data: %#v\n", m.hash([]byte(strconv.Itoa(i) + key))) // 0x6
			//log.Printf("hash: %d key: %s\n", hash, key) // 6 6
			m.keys = append(m.keys, hash)
			m.hashMap[hash] = key
		}
	}
	sort.Ints(m.keys)
}

// Get gets the closest item in the hash to the provided key.
func (m *Map) Get(key string) string {
	if len(m.keys) == 0 {
		return ""
	}

	hash := int(m.hash([]byte(key)))

	// Binary search for appropriate replica.
	idx := sort.Search(len(m.keys), func(i int) bool { return m.keys[i] >= hash })

	return m.hashMap[m.keys[idx % len(m.keys)]]
}
