package main

import (
	"fmt"
	"log"
	"reflect"
	"testing"
)

var tuples = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func TestGetter(t *testing.T) {
	var f Getter = GetterFunc(func(key string) ([]byte, error) {
		return []byte(key), nil
	})

	expect := []byte("key")
	v, _ := f.Get("key")
	log.Printf("v=%s\n", v)
	if !reflect.DeepEqual(v, expect) {
		t.Fatal("callback failed")
	}
}

func TestGet(t *testing.T) {
	loadCounts := make(map[string]int, len(tuples))
	fly := NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := tuples[key]; ok {
				if _, ok := loadCounts[key]; !ok {
					loadCounts[key] = 0
				}
				loadCounts[key]++
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exists", key)
		}))

	for k, v := range tuples {
		if view, err := fly.Get(k); err != nil || view.String() != v {
			t.Fatal("failed to get value of tom")
		}

		if _, err := fly.Get(k); err != nil || loadCounts[k] > 1 {
			t.Fatalf("cache %s miss", k)
		}
	}

	if view, err := fly.Get("undefined"); err == nil {
		t.Fatalf("the value of undefined should be empty, but %s got", view)
	}
}

func TestGetGroup(t *testing.T) {
	groupName := "students"
	NewGroup(groupName, 2<<10, GetterFunc(
		func(key string) (bytes []byte, err error) { return }))

	if group := GetGroup(groupName); group == nil || group.name != groupName {
		t.Fatalf("group %s not exists", groupName)
	}

	if group := GetGroup(groupName + "123"); group != nil {
		t.Fatalf("expect nil, but %s got", group.name)
	}
}
