package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/cloakwiss/p7ui/src"
	"github.com/go-chi/chi/v5"
)

var (

	//go:embed ui/64.png
	l64 []byte
	//go:embed ui/128.png
	l128 []byte
	//go:embed ui/256.png
	l256 []byte

	//go:embed ui/manifest.json
	manifest []byte

	//go:embed ui/index.html
	index []byte
	//go:embed ui/style.css
	style []byte
	//go:embed ui/datastar.js
	datastarScript []byte
	//go:embed ui/drag.js
	drag []byte
)

const port = 13337

type (
	target struct {
		Executable string `json:"target"`
		Hookdll    string `json:"hook"`
	}
)

func main() {
	for {
		func() {

			var (
				closing      = make(chan struct{})
				source, sink = src.CreateChannelBundle()

				router = chi.NewRouter()

				logger = src.NewLogger(source.LogC, closing)
				app    = src.ApplicationState{
					Log:             logger,
					IsCoreRunning:   false,
					HookPipeName:    `\\.\pipe\P7_HOOKS`,
					ControlPipeName: `\\.\pipe\P7_CONTROLS`,
					LogPipeName:     `\\.\pipe\P7_LOGS`,
				}
			)

			{ // assets
				router.Get("/", func(w http.ResponseWriter, r *http.Request) {
					w.Write(index)
				})

				router.Get("/manifest.json", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Add("Content-Type", "text/json")
					w.Write(manifest)
				})

				router.Get("/256.png", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Add("Content-Type", "image/png")
					w.Write(l256)
				})

				router.Get("/128.png", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Add("Content-Type", "image/png")
					w.Write(l128)
				})

				router.Get("/64.png", func(w http.ResponseWriter, r *http.Request) {
					w.Header().Add("Content-Type", "image/png")
					w.Write(l64)
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
			{ // routes
				router.Post("/target_pick", func(w http.ResponseWriter, r *http.Request) {
					var target = target{}

					defer r.Body.Close()

					if err := json.NewDecoder(r.Body).Decode(&target); err != nil {
						log.Println("Decode failed:", err)
						http.Error(w, "Invalid JSON", http.StatusBadRequest)
						return
					}
					// valid the target data
					if target.Executable == "" {
						app.Log.Error("Target not Picked.")
					} else {
						app.TargetPath = target.Executable
					}

					if target.Hookdll == "" {
						app.Log.Error("Hookdll not Picked.")
					} else {
						app.HookDllPath = target.Hookdll
					}

					log.Printf("Data: %+v", target)
				})

				router.Get("/spawnp7", func(w http.ResponseWriter, r *http.Request) {
					if app.TargetPath != "" && app.HookDllPath != "" {
						if !app.IsCoreRunning {
							go app.Launch(source.DataC)
							app.Log.Info("UI Started")
						} else {
							app.Log.Error("Already Running a P7 instance.")
						}
					} else {
						app.Log.Fatal("Target Path and HookDll path is empty.")
					}

					src.MainLoop(w, r, closing, sink)
				})

				router.Post("/stop", func(w http.ResponseWriter, r *http.Request) {
					app.Log.Info("Stop clicked")
					src.SendControl(&app, src.Stop)
					close(closing)
				})

				router.Post("/resume", func(w http.ResponseWriter, r *http.Request) {
					app.Log.Info("Resume clicked")
					src.SendControl(&app, src.Resume)
				})
				router.Post("/abort", func(w http.ResponseWriter, r *http.Request) {
					app.Log.Info("Abort clicked")
					src.SendControl(&app, src.Abort)
				})
				router.Post("/step", func(w http.ResponseWriter, r *http.Request) {
					app.Log.Info("Step clicked")

					if app.StepState {
						src.SendControl(&app, src.STEC)
					} else {
						src.SendControl(&app, src.STSC)
					}

					// To alter step to start at end of calls
					app.StepState = !app.StepState
				})
				router.Post("/stec", func(w http.ResponseWriter, r *http.Request) {
					app.Log.Info("STEC clicked")
					src.SendControl(&app, src.STEC)

					// To properly fall into the next call start
					app.StepState = false
				})
				router.Post("/stsc", func(w http.ResponseWriter, r *http.Request) {
					app.Log.Info("STSC clicked")
					src.SendControl(&app, src.STSC)

					// To properly fall into the next call end
					app.StepState = true
				})
			}

			log.Printf("Starting server on http://localhost:%d", port)

			server := http.Server{
				Addr:    fmt.Sprintf(":%d", port),
				Handler: router,
			}

			go server.ListenAndServe()

			<-closing

			log.Println("Closing the app")
			app.Log.Info("Closing the app.")
			// I dont know why this grace period
			// just felt like
			time.Sleep(500 * time.Millisecond)
			server.Close()

		}()
	}
}
