package lru

import (
	"testing"
)

func TestGet(t *testing.T) {
	c, _ := New(int64(0), nil)
	c.Set("k1", String("111"))
	if v, ok := c.Get("k1"); !ok || string(v.(String)) != "111" {
		t.Fatalf("cache hit key1=111 failed")
	}
	if _, ok := c.Get("k2"); ok {
		t.Fatalf("cache miss key2 failed")
	}
}

func TestAdd(t *testing.T) {
	c, _ := New(int64(0), nil)

	c.Set("k3", String("33333"))
	c.Set("k3", String("333333"))
	c.Set("k4", String("4"))
	c.Set("k4", String("4444"))
	//t.Log(c.Get("k3"))
	//t.Log(c.Get("k4"))
	if c.usedBytes != int64(len("k3")+len("333333")+len("k4")+len("4444")) {
		t.Fatal("expected 14 but got", c.usedBytes)
	}
}

func TestDel(t *testing.T) {
	c, _ := New(int64(0), nil)
	c.Set("k1", String("111"))

	c.Del("k1")
	if _, ok := c.Get("k1"); ok {
		t.Fatalf("cache delete k1 failed")
	}
}

func TestRemoveOldest(t *testing.T) {
	k1, k2, k3 := "k1", "k2", "k3"
	v1, v2, v3 := "v1", "v2", "v3"
	cap_ := len(k1 + k2 + v1 + v2)
	c, _ := New(int64(cap_), nil)
	c.Set(k1, String(v1))
	c.Set(k2, String(v2))
	c.Set(k3, String(v3))

	if _, ok := c.Get("k1"); ok || c.ll.Len() != 2 {
		t.Fatalf("Removeoldest key1 failed")
	}
}

func TestKeys(t *testing.T) {
	c, _ := New(int64(0), nil)

	c.Set("k5", String("5"))
	c.Set("k6", String("6"))

	keys := []string{"k5", "k6"}
    //fmt.Printf("%#v\n", c.Keys())
	if !c.Equal(c.Keys(), keys) {
		t.Fatalf("keys test failed")
	}
}

func TestClear(t *testing.T) {
	c, _ := New(int64(0), nil)
	c.Purge()
}



