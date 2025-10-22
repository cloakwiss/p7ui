package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/starfederation/datastar-go/datastar"
)

var (
	//go:embed ui/index.html
	index []byte

	//go:embed ui/style.css
	style []byte

	//go:embed ui/datastar.js
	datastarScript []byte

	//go:embed ui/drag.js
	drag []byte
)

const port = 1337

type (
	target struct {
		Executable string `json:"target"`
		Hookdll    string `json:"hook"`
	}
	data struct {
		name, section string
	}
)

func main() {
	router := chi.NewRouter()

	{ // assets
		router.Get("/", func(w http.ResponseWriter, r *http.Request) {
			w.Write(index)
		})

		router.Get("/style.css", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "text/css")
			w.Write(style)
		})

		router.Get("/datastar.js", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "text/javascript; charset=utf-8")
			w.Write(datastarScript)
		})

		router.Get("/drag.js", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "text/javascript; charset=utf-8")
			w.Write(drag)
		})
	}

	var dataChan = make(chan data)
	var closing = make(chan bool, 1)

	router.Post("/target_pick", func(w http.ResponseWriter, r *http.Request) {
		var target = target{}

		defer r.Body.Close()

		if err := json.NewDecoder(r.Body).Decode(&target); err != nil {
			log.Println("Decode failed:", err)
			http.Error(w, "Invalid JSON", http.StatusBadRequest)
			return
		}

		log.Printf("Data: %+v", target)
	})

	router.Post("/spawnp7", func(w http.ResponseWriter, r *http.Request) {
		mainLoop(w, r, dataChan)
	})

	router.Post("/stop", func(w http.ResponseWriter, r *http.Request) {
		closing <- true
	})

	log.Printf("Starting server on http://localhost:%d", port)

	go feedChannel(dataChan, closing)

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), router); err != nil {
		panic(err)
	}
}

func mainLoop(w http.ResponseWriter, r *http.Request, dataChan <-chan data) {
	sse := datastar.NewSSE(w, r)
	modeOpt := datastar.WithModeAppend()
	container := datastar.WithSelectorID("table-body")
	for message := range dataChan {
		formatted := fmt.Sprintf("<tr><td>%s</td><td>%s</td></tr>\n", message.name, message.section)
		if err := sse.PatchElements(formatted, modeOpt, container); err != nil {
			return
		}
	}

}

func feedChannel(dataChan chan<- data, control chan bool) {
	const message = "Hello, world!"
	i := 0
main:
	for {
		select {
		case <-control:
			break main
		default:
			{
				time.Sleep(250 * time.Millisecond)
				if i < len(message) {
					dataChan <- data{message, message}
				} else {
					dataChan <- data{message, message}
				}
				i++
			}
		}
	}
	close(control)
}
