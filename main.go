package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
)

var rows = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func createGroup() *Group {
	return NewGroup("scores", 2<<10, GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := rows[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))
}

func startCacheServer(node string, nodes []string, g *Group) {
	peers := NewHTTPServer(node)
	peers.Set(nodes...)
	g.RegisterPeers(peers)
	log.Println("flycache is running at", node)
	log.Fatal(http.ListenAndServe(node[7:], peers))
}

func startAPIServer(apiAddr string, g *Group) {
	http.Handle("/api", http.HandlerFunc(
		func(w http.ResponseWriter, r *http.Request) {
			key := r.URL.Query().Get("key")
			view, err := g.Get(key)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Write(view.ByteSlice())

		}))
	log.Println("frontend server is running at", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))

}

func main() {
	//NewHTTPServer("127.0.0.0").Listen()
	var port int
	var api bool
	flag.IntVar(&port, "port", 8001, "cache server port")
	flag.BoolVar(&api, "api", false, "Start a api server?")
	flag.Parse()

	apiHost := "http://localhost:8888"
	hostMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var hosts []string
	for _, v := range hostMap {
		hosts = append(hosts, v)
	}

	g := createGroup()
	if api {
		go startAPIServer(apiHost, g)
	}
	startCacheServer(hostMap[port], hosts, g)
}
