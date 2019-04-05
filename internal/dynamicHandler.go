package internal

import (
	"log"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"
)

// DefaultHandlerTTL is the default TTL after which a dynamic path handler will the uninstalled if it is inactive
// for at least that duration
const DefaultHandlerTTL = time.Minute * 30

// DefaultPruneTicker sets how often we should check for stale handlers
const DefaultPruneTicker = time.Minute * 1

type expiringHandler struct {
	*http.HandlerFunc           // the actual handler
	lastAccessed      time.Time // the last time this handler was accessed
	path              string
	store             storage
}

func newexpiringHandler(path string, store storage) expiringHandler {
	eh := expiringHandler{path: path, store: store}
	h := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		// Capture the request
		request.Header.Add("X-Recv-Timestamp", time.Now().Format(time.StampMilli))
		rbuf, _ := httputil.DumpRequest(request, true)
		eh.store.append(eh.path, rbuf)
		eh.lastAccessed = time.Now()
	})
	eh.HandlerFunc = &h
	eh.lastAccessed = time.Now()

	return eh
}

func pruneHandlers(ttl time.Duration, handlers *sync.Map) {
	for range time.NewTicker(getHandlerTTL()).C {
		handlers.Range(func(key, value interface{}) bool {
			h := value.(expiringHandler)
			age := time.Now().Sub(h.lastAccessed)
			if age >= ttl {
				// delete
				log.Printf("Pruning old handler for path: %s Age: %v", h.path, age)
				handlers.Delete(key)
				h.store.delete(h.path)
			}
			return true
		})
	}
}
