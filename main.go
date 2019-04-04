package main

import (
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"
)

// DefaultPort is the default port to use if once is not specified by the SERVER_PORT environment variable
const DefaultPort = "9000"

var pathmap = sync.Map{}
var dataStore = newMemStore()

func getServerPort() string {
	port := os.Getenv("SERVER_PORT")
	if port != "" {
		return port
	}
	return DefaultPort
}

func getHandlerTTL() time.Duration {
	ttl := os.Getenv("HANDLER_TTL")
	if ttl != "" {
		ttlDur, err := time.ParseDuration(ttl)
		if err != nil {
			log.Printf("Invalid handler TTL duration: %s "+
				"using default: %v. (Use go style duration string)",
				ttl,
				DefaultHandlerTTL)
			return DefaultHandlerTTL
		}
		return ttlDur
	}
	return DefaultHandlerTTL
}

// dynamicHandler dynamically creates handlers for paths that it sees for the first time
func dynamicHandler(writer http.ResponseWriter, request *http.Request) {
	path := request.URL.Path
	h, ok := pathmap.Load(path)
	// new path detected
	if !ok {
		h = newexpiringHandler(path, dataStore)
		pathmap.Store(path, h)
	}
	handler := h.(expiringHandler)
	handler.ServeHTTP(writer, request)
}

func createMainHandler(fsHandler http.Handler) http.HandlerFunc {
	defaultPaths := []string{"favicon.ico", "css", "html"}
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/" {
			dataStore.foreach(func(key string, values []interface{}) bool {
				log.Printf("Handler: %s", key)
				for _, v := range values {
					log.Printf("\t Request: %s", v)
				}
				return true
			})
			fsHandler.ServeHTTP(w, r)
			return
		}

		for _, defaultPath := range defaultPaths {
			if strings.Contains(path, defaultPath) {
				fsHandler.ServeHTTP(w, r)
				return
			}
		}

		dynamicHandler(w, r)
	}
}

func main() {
	log.Println("Starting server on port " + getServerPort())
	go pruneHandlers(DefaultHandlerTTL, &pathmap)
	fs := http.FileServer(http.Dir("static"))
	ffs := http.StripPrefix("/static/", fs)
	http.HandleFunc("/", createMainHandler(ffs))
	http.ListenAndServe(":"+getServerPort(), nil)
}
