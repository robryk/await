package await

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"path"
	"sync"
)

const idLen = 16

func generateId() string {
	buf := make([]byte, idLen)
	_, err := rand.Read(buf)
	if err != nil {
		panic(err)
	}
	return hex.EncodeToString(buf)
}

type AwaitServer struct {
	urlPrefix string // readonly

	sleepers map[string]chan<- struct{} // protected by mu
	mu       sync.Mutex
}

func (as *AwaitServer) wakeUp(id string) bool {
	as.mu.Lock()
	defer as.mu.Unlock()
	ch, ok := as.sleepers[id]
	if !ok {
		return false
	}
	close(ch)
	delete(as.sleepers, id)
	return true
}

func (as *AwaitServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := path.Base(r.URL.Path)
	if !as.wakeUp(id) {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (as *AwaitServer) New() Await {
	ch := make(chan struct{})
	var id string
	for {
		id = generateId()
		as.mu.Lock()
		if _, ok := as.sleepers[id]; !ok {
			break
		}
		as.mu.Unlock()
	}
	as.sleepers[id] = ch
	as.mu.Unlock()
	return Await{
		ch: ch,
		id: id,
		as: as,
	}
}

func NewAwaitServer(urlPrefix string) *AwaitServer {
	as := &AwaitServer{
		urlPrefix: urlPrefix,
		sleepers:  make(map[string]chan<- struct{}),
	}
	return as
}

type Await struct {
	ch <-chan struct{}
	id string
	as *AwaitServer
}

func (a Await) Chan() <-chan struct{} {
	return a.ch
}

func (a Await) Cancel() {
	as := a.as
	as.mu.Lock()
	defer as.mu.Unlock()

	if ch, ok := as.sleepers[a.id]; ok {
		close(ch)
		delete(as.sleepers, a.id)
	}
}

func (a Await) Url() string {
	return fmt.Sprintf("%s/%s", a.as.urlPrefix, a.id)
}
