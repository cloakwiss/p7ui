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

const port = 13337

type (
	target struct {
		Executable string `json:"target"`
		Hookdll    string `json:"hook"`
	}

	channelBundleSink struct {
		logC  <-chan string
		dataC <-chan string
	}

	channelBundleSource struct {
		logC  chan<- string
		dataC chan<- string
	}
)

func createChannelBundle() (channelBundleSource, channelBundleSink) {
	var (
		logC  = make(chan string, 1000)
		dataC = make(chan string, 1000)
	)
	return channelBundleSource{logC, dataC}, channelBundleSink{logC, dataC}
}

func main() {
	for {
		func() {

			var (
				closing      = make(chan struct{})
				source, sink = createChannelBundle()

				router = chi.NewRouter()

				logger = src.NewLogger(source.logC, closing)
				app    = src.ApplicationState{
					Log:           logger,
					IsCoreRunning: false,
					InPipeName:    `\\.\pipe\P7_HOOKS`,
					OutPipeName:   `\\.\pipe\P7_CONTROLS`,
				}
			)

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

				router.Post("/spawnp7", func(w http.ResponseWriter, r *http.Request) {
					if app.TargetPath != "" && app.HookDllPath != "" {
						if !app.IsCoreRunning {
							go src.Launch(&app, source.dataC)
							app.Log.Info("UI Started")
						} else {
							app.Log.Error("Already Running a P7 instance.")
						}
					} else {
						app.Log.Fatal("Target Path and HookDll path is empty.")
					}

					mainLoop(w, r, closing, sink)
				})

				router.Post("/stop", func(w http.ResponseWriter, r *http.Request) {
					app.Log.Info("Stop clicked")
					SendControl(&app, src.Stop)
					close(closing)
				})

				router.Post("/resume", func(w http.ResponseWriter, r *http.Request) {
					app.Log.Info("Resume clicked")
					SendControl(&app, src.Resume)
				})
				router.Post("/abort", func(w http.ResponseWriter, r *http.Request) {
					app.Log.Info("Abort clicked")
					SendControl(&app, src.Abort)
				})
				router.Post("/step", func(w http.ResponseWriter, r *http.Request) {
					app.Log.Info("Step clicked")

					if app.StepState {
						SendControl(&app, src.STEC)
					} else {
						SendControl(&app, src.STSC)
					}

					// To alter step to start at end of calls
					app.StepState = !app.StepState
				})
				router.Post("/stec", func(w http.ResponseWriter, r *http.Request) {
					app.Log.Info("STEC clicked")
					SendControl(&app, src.STEC)

					// To properly fall into the next call start
					app.StepState = false
				})
				router.Post("/stsc", func(w http.ResponseWriter, r *http.Request) {
					app.Log.Info("STSC clicked")
					SendControl(&app, src.STSC)

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

func mainLoop(w http.ResponseWriter, r *http.Request, control <-chan struct{}, sink channelBundleSink) {
	sse := datastar.NewSSE(w, r)
	modeOpt := datastar.WithModeAppend()
	container1 := datastar.WithSelectorID("console")
	container2 := datastar.WithSelectorID("hooks")
	for {
		select {
		case <-control:
			return
		case logLine := <-sink.logC:
			{
				formatted := fmt.Sprintf("<tr><td>%s</td></tr>\n", logLine)
				if err := sse.PatchElements(formatted, modeOpt, container1); err != nil {
					return
				}
			}
		case data := <-sink.dataC:
			{
				formatted := fmt.Sprintf("<tr><td>%s</td></tr>\n", data)
				if err := sse.PatchElements(formatted, modeOpt, container2); err != nil {
					return
				}
			}
		}
	}
}

func SendControl(p7 *src.ApplicationState, controlSignal src.Control) {

	if p7.OutPipe != nil {
		b := []byte{byte(controlSignal)}
		//TODO: why this go routine, it fells like channel
		// can do it
		go func() {
			_, err := p7.OutPipe.Write(b)

			if err != nil {
				p7.Log.Error("Write error: %v\n", err)
			}
			// } else {
			// 	p7.Log.Debug("Wrote Signal %d", controlSignal)
			// }
		}()

	} else {
		if !p7.IsCoreRunning && (p7.OutPipe == nil) {
			p7.Log.Error("P7 is not running")
		} else if p7.OutPipe == nil {
			p7.Log.Error("OutPipe is not connected")
		}
	}
	log.Println("Sent Control")
}
