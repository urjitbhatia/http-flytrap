package internal

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

var pathmap = sync.Map{}
var dataStore = newMemStore()

type templateData struct {
	CapturePort string
	HandlerTTL  string
	HandlerData []handlerData
}

var tdata = templateData{HandlerTTL: getHandlerTTL().String()}

type handlerData struct {
	Path string
	Reqs [][]string
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

func createQueryHandler(fsHandler http.Handler) http.HandlerFunc {
	defaultPaths := []string{"css", "logos"}
	favico := func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/favicon.ico")
	}
	return func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		if path == "/favicon.ico" {
			favico(w, r)
			return
		}
		for _, defaultPath := range defaultPaths {
			if strings.Contains(path, defaultPath) {
				fsHandler.ServeHTTP(w, r)
				return
			}
		}
		// otherwise, serve the index
		if path == "/" {
			serveTemplate(w, r)
			return
		}
	}
}

func serveTemplate(w http.ResponseWriter, r *http.Request) {
	lp := filepath.Join("templates", "layout.html")

	tmpl, err := template.ParseFiles(lp)
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

// Trap starts the flytrap capture
func Trap(capturePort, queryPort string, ttl time.Duration) {
	// set capture port for template
	tdata.CapturePort = capturePort
	go pruneHandlers(ttl, &pathmap)

	// query server
	log.Printf("Starting query server on port %s", queryPort)
	fs := http.FileServer(http.Dir("static"))
	querySrv := http.NewServeMux()
	// querySrv.Handle("/", createQueryHandler(fs))
	querySrv.Handle("/", createQueryHandler(http.StripPrefix("/static/", fs)))
	go func() {
		log.Printf("Query server exiting with error: %s", http.ListenAndServe(":"+queryPort, querySrv).Error())
	}()

	// capture server
	log.Printf("Laying trap on port %s", capturePort)
	captureSrv := http.NewServeMux()
	captureSrv.Handle("/", http.HandlerFunc(dynamicHandler))
	log.Printf("Capture server exiting with error: %s", http.ListenAndServe(":"+capturePort, captureSrv).Error())
}
