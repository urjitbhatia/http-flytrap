package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// DefaultPort is the default port to use if once is not specified by the SERVER_PORT environment variable
const DefaultPort = "9000"

var pathmap = sync.Map{}
var dataStore = newMemStore()

type templateData struct {
	HandlerTTL  string
	HandlerData []handlerData
}

var tdata = templateData{HandlerTTL: getHandlerTTL().String()}

type handlerData struct {
	Path string
	Reqs [][]string
}

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
			serveTemplate(w, r)
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

func serveTemplate(w http.ResponseWriter, r *http.Request) {
	lp := filepath.Join("templates", "layout.html")
	fp := filepath.Join("templates", "components.html")

	tmpl, err := template.ParseFiles(lp, fp)
	if err != nil {
		// Log the detailed error
		log.Println(err.Error())
		// Return a generic "Internal Server Error" message
		http.Error(w, http.StatusText(500), 500)
		return
	}

	data := []handlerData{}
	dataStore.foreach(func(key string, values []interface{}) bool {
		d := handlerData{}
		d.Path = key
		for _, v := range values {
			// restore the body for future reads
			displayVal := fmt.Sprintf("Request: %s", v)
			// for better formatting
			lines := strings.Split(displayVal, "\n")
			d.Reqs = append(d.Reqs, lines)
		}
		data = append(data, d)
		return true
	})
	tdata.HandlerData = data

	if err := tmpl.ExecuteTemplate(w, "layout", tdata); err != nil {
		log.Println(err.Error())
		http.Error(w, http.StatusText(500), 500)
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
