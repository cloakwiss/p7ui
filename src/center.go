package src

import (
	"fmt"
	"log"
	"net/http"

	"github.com/starfederation/datastar-go/datastar"
)

type (
	LogLine struct {
		timestamp, msg string
		level          LogLevel
	}

	HookData struct {
		lines []string
	}

	ChannelBundleSink struct {
		LogC  <-chan LogLine
		DataC <-chan HookData
	}

	ChannelBundleSource struct {
		LogC  chan<- LogLine
		DataC chan<- HookData
	}
)

func NewLogLine(timestamp string, level LogLevel, msg string) LogLine {
	return LogLine{timestamp, msg, level}
}

func CreateChannelBundle() (ChannelBundleSource, ChannelBundleSink) {
	var (
		logC  = make(chan LogLine, 1000)
		dataC = make(chan HookData, 1000)
	)
	return ChannelBundleSource{logC, dataC}, ChannelBundleSink{logC, dataC}
}

func MainLoop(w http.ResponseWriter, r *http.Request, control <-chan struct{}, sink ChannelBundleSink) {
	sse := datastar.NewSSE(w, r)
	modeOpt := datastar.WithModeAppend()
	container1 := datastar.WithSelectorID("console")
	container2 := datastar.WithSelectorID("hooks")
	for {
		select {
		case <-control:
			return
		case logLine := <-sink.LogC:
			{
				formatted := fmt.Sprintf("<tr><td>%s</td></tr>\n", logLine)
				if err := sse.PatchElements(formatted, modeOpt, container1); err != nil {
					return
				}
			}
		case data := <-sink.DataC:
			{
				formatted := fmt.Sprintf("<tr><td>%s</td></tr>\n", data)
				if err := sse.PatchElements(formatted, modeOpt, container2); err != nil {
					return
				}
			}
		}
	}
}

func SendControl(p7 *ApplicationState, controlSignal Control) {

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
