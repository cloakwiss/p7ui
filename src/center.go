package src

import (
	"fmt"
	"log"
	"net/http"

	"github.com/starfederation/datastar-go/datastar"
)

type (
	LogLine struct {
		timestamp    string
		level        LogLevel
		msg, payload string
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
	return LogLine{timestamp, level, msg, ""}
}

func NewLogLineWithPayload(timestamp string, level LogLevel, msg string, payload string) LogLine {
	return LogLine{timestamp, level, msg, payload}
}

func (l LogLine) String() string {
	if l.payload == "" {
		return fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%s</td></tr>\n", l.timestamp, l.level, l.msg)
	} else {
		return fmt.Sprintf("<tr><td>%s</td><td>%s</td><td><details><summary>%s</summary>%s</details></td></tr>\n", l.timestamp, l.level, l.msg, l.payload)
	}
}

func (h HookData) String() string {
	var lines string
	for i, line := range h.lines {
		lines += fmt.Sprintf("<tr><td>%d</td><td>%s</td></tr>\n", i, line)
	}
	return lines
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
				if err := sse.PatchElements(logLine.String(), modeOpt, container1); err != nil {
					return
				}
			}
		case data := <-sink.DataC:
			{
				if err := sse.PatchElements(data.String(), modeOpt, container2); err != nil {
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
				p7.Log.ErrorWithPayload("Write error: %v\n", err)
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
