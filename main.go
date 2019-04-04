package main

import (
	"log"
	"net/http"
	"os"
	"sync"
)

// DefaultPort is the default port to use if once is not specified by the SERVER_PORT environment variable
const DefaultPort = "9000"

var pathmap = sync.Map{}

func getServerPort() string {
	port := os.Getenv("SERVER_PORT")
	if port != "" {
		return port
	}
	return DefaultPort
}

// initHandler dynamically creates handlers for paths that it sees for the first time
func initHandler(writer http.ResponseWriter, request *http.Request) {
	path := request.URL.Path
	if path == "/" {
		return
	}
	h, ok := pathmap.Load(path)
	// new path detected
	if !ok {
		h = newexpiringHandler(path)
		pathmap.Store(path, h)
	}
	handler := h.(expiringHandler)
	handler.ServeHTTP(writer, request)
}

func main() {
	log.Println("Starting server on port " + getServerPort())
	http.HandleFunc("/", initHandler)
	http.ListenAndServe(":"+getServerPort(), nil)
}
