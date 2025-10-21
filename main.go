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

//go:embed ui/static/index.html
var index []byte

//go:embed ui/static/style.css
var style []byte

//go:embed ui/static/datastar.js
var datastarScript []byte

const port = 1337

type target struct {
	Executable string `json:"target"`
	Hookdll    string `json:"hook"`
}

func main() {
	router := chi.NewRouter()

	const message = "Hello, world!"

	var dataChan = make(chan string)
	var closing = make(chan bool, 1)

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
	}

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
		sse := datastar.NewSSE(w, r)
		modeOpt := datastar.WithModeAppend()
		container := datastar.WithSelectorID("console-output")
		for message := range dataChan {
			if err := sse.PatchElements(`<div>`+message+`</div>`, modeOpt, container); err != nil {
				return
			}
		}
	})

	router.Post("/stop", func(w http.ResponseWriter, r *http.Request) {
		closing <- true
	})

	log.Printf("Starting server on http://localhost:%d", port)

	go func() {
		i := 0
	main:
		for {
			select {
			case <-closing:
				close(closing)
				break main
			default:
				{
					time.Sleep(1 * time.Second)
					if i < len(message) {
						dataChan <- message[:i]
					} else {
						dataChan <- message
					}
					i++
				}
			}
		}
		// dataChan <- "Done"
		close(dataChan)
	}()

	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), router); err != nil {
		panic(err)
	}
}
