package main

import (
	"flycache/chash"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const (
	basePath        = "/cache/"
	defaultReplicas = 5
)

var (
	_ PeerPicker = (*cacheHandler)(nil)
	_ PeerGetter = (*httpGetter)(nil)
)

type Peer interface {
	Set(nodes ...string)
	Get(group string, key string)
}

type HTTPServer struct {
	host     string
	basePath string
	//Peer
	mutex       sync.Mutex
	peers       *chash.Map
	httpGetters map[string]*httpGetter
}

type cacheHandler struct {
	*HTTPServer
	mutex       sync.Mutex
	peers       *chash.Map
	httpGetters map[string]*httpGetter
}

func (h *HTTPServer) PickPeer(key string) (peer PeerGetter, ok bool) {
	//panic("implement me")
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if peer := h.peers.Get(key); peer != "" && peer != h.host {
		h.LogPrintf("Pick peer %s", peer)
		return h.httpGetters[peer], true
	}

	return nil, false
}

/*type statusHandler struct {
	*HTTPServer
}*/

func NewHTTPServer(host string) *HTTPServer {
	return &HTTPServer{host: host, basePath: basePath}
}

func (h *HTTPServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, basePath) {
		panic("unexpected path: " + r.URL.Path)
	}

	parts := strings.SplitN(r.URL.EscapedPath()[len(basePath):], "/", 2)
	if len(parts) == 0 || len(parts) != 2 {
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := GetGroup(groupName)
	if group == nil {
		group = NewGroup(groupName, 2<<10, GetterFunc(
			func(key string) ([]byte, error) {
				return []byte(key), nil
			}))

		/*http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return*/
	}

	view, err := group.Get(key)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}

/*func (s *HTTPServer) cacheHandler() http.Handler {
	return &cacheHandler{}
}*/

/*func (s *HTTPServer) Listen() {
	http.Handle(basePath, s.cacheHandler())
	//http.Handle("/status", s.statusHandler())
	http.ListenAndServe(":8090", nil)
}*/

type httpGetter struct {
	baseURL string
}

func (h *HTTPServer) Set(peers ...string) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.peers = chash.New(defaultReplicas, nil)
	h.peers.Add(peers...)
	h.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers {
		h.httpGetters[peer] = &httpGetter{baseURL: peer + h.basePath}
	}
}

func (h *httpGetter) Get(group string, key string) ([]byte, error) {
	u := fmt.Sprintf("%v%v/%v", h.baseURL, url.QueryEscape(group), url.QueryEscape(key))

	res, err := http.Get(u)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}

	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response body: %v", err)
	}

	return bytes, nil
}

func (h *HTTPServer) LogPrintf(format string, v ...interface{}) {
	log.Printf("[Server %s] %s", h.host, fmt.Sprintf(format, v...))
}
