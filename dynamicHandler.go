package main

import (
	"net/http"
	"net/http/httputil"
	"sync"
	"time"

	"github.com/google/uuid"
)

// DefaultHandlerTTL is the default TTL after which a dynamic path handler will the uninstalled if it is inactive
// for at least that duration
const DefaultHandlerTTL = time.Hour * 5

type expiringHandler struct {
	*http.HandlerFunc           // the actual handler
	lastAccessed      time.Time // the last time this handler was accessed
	id                string
	store             storage
}

func newexpiringHandler(path string, store storage) expiringHandler {
	eh := expiringHandler{id: uuid.New().String(), store: store}
	h := http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {

		// Capture the request body
		rbuf, _ := httputil.DumpRequest(request, true)
		eh.store.append(path, rbuf)
		request.Header.Add("X-Echo-Handler", eh.id)
		request.Write(writer)
		eh.lastAccessed = time.Now()
	})
	eh.HandlerFunc = &h
	eh.lastAccessed = time.Now()

	return eh
}

func pruneHandlers(ttl time.Duration, handlers *sync.Map) {
	for range time.NewTicker(time.Second * 30).C {
		handlers.Range(func(key, value interface{}) bool {
			h := value.(expiringHandler)
			age := h.lastAccessed.Sub(time.Now())
			if age >= ttl {
				// delete
				println("Pruning old handler for path: %s. Age: %s", key, age)
				handlers.Delete(key)
			}
			return true
		})
	}
}
