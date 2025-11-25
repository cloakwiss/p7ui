package main

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"github.com/cloakwiss/p7ui/p7"
	"github.com/go-chi/chi/v5"
	"github.com/sqweek/dialog"
	"github.com/starfederation/datastar-go/datastar"
)

const port = 13337

func pickFile() (string, error) {
	selected, err := dialog.File().Title("Choose a file").Load()
	if err != nil {
		return "Error in Picking File", nil
	}

	abs, err := filepath.Abs(selected)
	if err != nil {
		return selected, nil
	}

	return abs, nil
}

func main() {
	func() {
		var (
			dbconn, closedb = p7.OpenDB()
			search          = p7.NewSearch(dbconn, 256)
			closing         = make(chan struct{})
			source, sink    = p7.CreateChannelBundle()
			descChan        = make(chan string)

			router = chi.NewRouter()

			logger = p7.NewLogger(source.LogC, closing)
			app    = p7.ApplicationState{
				Log:             logger,
				IsCoreRunning:   false,
				HookChannel:     source.DataC,
				HookPipeName:    `\\.\pipe\P7_HOOKS`,
				ControlPipeName: `\\.\pipe\P7_CONTROLS`,
				LogPipeName:     `\\.\pipe\P7_LOGS`,
				QueryFunction:   search,
			}
		)
		defer func() {
			if er := closedb(); er != nil {
				log.Fatalln("Failed to close the db connection")
			}
		}()

		{ // assets
			router.Get("/", func(w http.ResponseWriter, r *http.Request) {
				http.ServeFile(w, r, "ui/index.html")
			})

			// TODO: First need to find new logo
			// router.Get("/manifest.json", func(w http.ResponseWriter, r *http.Request) {
			// 	w.Header().Add("Content-Type", "text/json")
			// 	http.ServeFile(w, r, "ui/maifest.json")
			// })

			// router.Get("/256.png", func(w http.ResponseWriter, r *http.Request) {
			// 	w.Header().Add("Content-Type", "image/png")
			// 	http.ServeFile(w, r, "ui/index.html")
			// })

			// router.Get("/128.png", func(w http.ResponseWriter, r *http.Request) {
			// 	w.Header().Add("Content-Type", "image/png")
			// })

			// router.Get("/64.png", func(w http.ResponseWriter, r *http.Request) {
			// 	w.Header().Add("Content-Type", "image/png")
			// })

			router.Get("/style.css", func(w http.ResponseWriter, r *http.Request) {
				http.ServeFile(w, r, "ui/style.css")
			})

			router.Get("/datastar.js", func(w http.ResponseWriter, r *http.Request) {
				http.ServeFile(w, r, "ui/datastar.js")
			})

			router.Get("/drag.js", func(w http.ResponseWriter, r *http.Request) {
				http.ServeFile(w, r, "ui/drag.js")
			})

			router.Get("/toggle_stop_resume.js", func(w http.ResponseWriter, r *http.Request) {
				http.ServeFile(w, r, "ui/toggle_stop_resume.js")
			})

			router.Get("/filepicker.js", func(w http.ResponseWriter, r *http.Request) {
				http.ServeFile(w, r, "ui/filepicker.js")
			})

			// Is this name url safe
			router.Get("/ctrl&Shortcut.js", func(w http.ResponseWriter, r *http.Request) {
				http.ServeFile(w, r, "ui/ctrl&Shortcut.js")
			})

			router.Get("/stec-logo.png", func(w http.ResponseWriter, r *http.Request) {
				http.ServeFile(w, r, "ui/stec-logo.png")
			})

			router.Get("/stsc-logo.png", func(w http.ResponseWriter, r *http.Request) {
				http.ServeFile(w, r, "ui/stsc-logo.png")
			})
		}
		{ // routes
			router.Get("/picktarget", func(w http.ResponseWriter, r *http.Request) {
				name, error := pickFile()
				if error != nil {
					app.Log.Error("Error in file picking")
				}
				app.TargetPath = name
				defer r.Body.Close()
				sse := datastar.NewSSE(w, r)
				container1 := datastar.WithSelectorID("target_path")

				input := `<input type="text" id="target_path" placeholder="Selected target executable path will appear here" value="` + name + `" readonly>`
				if err := sse.PatchElements(input, container1); err != nil {
					return
				}
			})
			router.Get("/pickhookdll", func(w http.ResponseWriter, r *http.Request) {
				name, error := pickFile()
				if error != nil {
					app.Log.Error("Error in file picking")
				}
				app.HookDllPath = name
				defer r.Body.Close()
				sse := datastar.NewSSE(w, r)
				container1 := datastar.WithSelectorID("hookdll_path")

				input := `<input type="text" id="hookdll_path" placeholder="Selected hookdll will appear here" value="` + name + `" readonly>`
				if err := sse.PatchElements(input, container1); err != nil {
					return
				}
			})

			router.Get("/spawnp7", func(w http.ResponseWriter, r *http.Request) {
				fmt.Println(r.Header)
				if app.TargetPath != "" && app.HookDllPath != "" {
					if !app.IsCoreRunning {
						go app.Launch()
						app.Log.Info("UI Started")
					} else {
						app.Log.Error("Already Running a P7 instance.")
					}
				} else {
					app.Log.Error("Target Path and HookDll path is empty.")
				}

				p7.MainLoop(w, r, closing, sink, descChan)
			})
			router.Post("/stop", func(w http.ResponseWriter, r *http.Request) {
				app.Log.Info("Stop clicked")
				p7.SendControl(&app, p7.Stop)
			})

			router.Post("/resume", func(w http.ResponseWriter, r *http.Request) {
				app.Log.Info("Resume clicked")
				p7.SendControl(&app, p7.Resume)
			})
			router.Post("/abort", func(w http.ResponseWriter, r *http.Request) {
				app.Log.Info("Abort clicked")
				p7.SendControl(&app, p7.Abort)
				close(closing)
			})
			router.Post("/step", func(w http.ResponseWriter, r *http.Request) {
				app.Log.Info("Step clicked")

				if app.StepState {
					p7.SendControl(&app, p7.STEC)
				} else {
					p7.SendControl(&app, p7.STSC)
				}

				// To alter step to start at end of calls
				app.StepState = !app.StepState
			})
			router.Post("/stec", func(w http.ResponseWriter, r *http.Request) {
				app.Log.Info("STEC clicked")
				p7.SendControl(&app, p7.STEC)

				// To properly fall into the next call start
				app.StepState = false
			})
			router.Post("/stsc", func(w http.ResponseWriter, r *http.Request) {
				app.Log.Info("STSC clicked")
				p7.SendControl(&app, p7.STSC)

				// To properly fall into the next call end
				app.StepState = true
			})

			router.Get("/search/{symbol}", func(w http.ResponseWriter, r *http.Request) {
				functionSymbol := chi.URLParam(r, "symbol")
				funcDesc := p7.Get(&app.QueryFunction, functionSymbol).Description
				descChan <- funcDesc
				app.Log.Debug("Desc sent")
			})
		}

		log.Printf("Starting server on http://localhost:%d", port)

		server := http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: router,
		}

		go server.ListenAndServe()

		// Launch browser window
		// url := fmt.Sprintf("http://localhost:%d", port)
		// browserCmd := launchBrowser(url)
		// fmt.Println(url)

		<-closing

		app.Log.Info("Closing the app.")

		// Kill browser if launched
		// if browserCmd != nil && browserCmd.Process != nil {
		// 	browserCmd.Process.Kill()
		// }

		time.Sleep(500 * time.Millisecond)
		server.Close()

	}()

}

// Launches a browser and returns the command (so we can kill it later)
func launchBrowser(url string) *exec.Cmd {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		for _, browser := range []string{"chrome", "msedge", "firefox", "chromium"} {
			if path, err := exec.LookPath(browser); err == nil {
				cmd = exec.Command(path, "--app="+url)
				cmd.Start()
				return cmd
			}
		}
		exec.Command("cmd", "/c", "start", "", url).Start()

	case "darwin":
		for _, browser := range []string{"/Applications/Google Chrome.app/Contents/MacOS/Google Chrome",
			"/Applications/Microsoft Edge.app/Contents/MacOS/Microsoft Edge",
			"/Applications/Firefox.app/Contents/MacOS/firefox"} {
			if _, err := exec.LookPath(browser); err == nil {
				cmd = exec.Command(browser, "--app="+url)
				cmd.Start()
				return cmd
			}
		}
		exec.Command("open", url).Start()

	default: // Linux
		for _, browser := range []string{"google-chrome", "chromium", "firefox"} {
			if path, err := exec.LookPath(browser); err == nil {
				cmd = exec.Command(path, "--app="+url)
				cmd.Start()
				return cmd
			}
		}
		exec.Command("xdg-open", url).Start()
	}
	return nil
}
