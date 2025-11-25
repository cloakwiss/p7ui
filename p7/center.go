package p7

import (
	"fmt"
	"log"
	"net/http"

	"github.com/starfederation/datastar-go/datastar"
)

type HookType uint8

const (
	Call   HookType = 0x01
	Return HookType = 0x02
)

type (
	LogLine struct {
		timestamp string
		level     LogLevel
		msg       string
	}

	HookData struct {
		HookType
		srno int
		call HookCall
		ret  HookReturns
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
	return LogLine{timestamp, level, msg}
}

func (l LogLine) String() string {
	return fmt.Sprintf("<tr><td>%s</td><td>%s</td><td>%s</td></tr>\n", l.timestamp, l.level, l.msg)
}

func (h HookData) String() string {
	switch h.HookType {
	case Call:
		return fmt.Sprintf(
			`<tr>
				<td>%d</td>
				<td>%s</td>
				<td>
					<details>
						<summary><a data-on-click="@get('/search/%s')">%s</a></summary>
						<pre>%v</pre>
					</details>
				</td>
				</tr>
			`,
			h.srno,
			"Call",
			h.call.id,
			h.call.id,
			h.call)
	case Return:
		return fmt.Sprintf(
			`<tr>
				<td>%d</td>
				<td>%s</td>
				<td>
					<details>
						<summary><a data-on-click="@get('/search/%s')">%s</a></summary>
						<pre>%v</pre>
					</details>
				</td>
				</tr>
			`,
			h.srno,
			"Return",
			h.ret.id,
			h.ret.id,
			h.ret)
	default:
		return "Error"
	}
}

func CreateChannelBundle() (ChannelBundleSource, ChannelBundleSink) {
	var (
		logC  = make(chan LogLine, 1000)
		dataC = make(chan HookData, 1000)
	)
	return ChannelBundleSource{logC, dataC}, ChannelBundleSink{logC, dataC}
}

func MainLoop(w http.ResponseWriter, r *http.Request, control <-chan struct{}, sink ChannelBundleSink, descChan <-chan string) {
	sse := datastar.NewSSE(w, r)
	modeOpt := datastar.WithModeAppend()
	modeWithInner := datastar.WithModeInner()
	container1 := datastar.WithSelectorID("console")
	container2 := datastar.WithSelectorID("hooks")
	container3 := datastar.WithSelectorID("description")
	for {
		select {
		case <-control:
			return
		case logLine := <-sink.LogC:
			{
				if err := sse.PatchElements(logLine.String(), modeOpt, container1); err != nil {
					log.Panicf("LogC: %s", err)
				}
			}
		case data := <-sink.DataC:
			{
				if err := sse.PatchElements(data.String(), modeOpt, container2); err != nil {
					log.Panicf("DataC: %s", err)
				}
			}
		case desc :=  <-descChan:
			{
				if err := sse.PatchElements(desc, modeWithInner,container3); err != nil {
					log.Panicf("Description error: %s", err)
				}
			}
		}
	}
}

func SendControl(p7 *ApplicationState, controlSignal Control) {

	if p7.ControlPipe != nil {
		b := []byte{byte(controlSignal)}
		//TODO: why this go routine, it fells like channel
		// can do it
		go func() {
			_, err := p7.ControlPipe.Write(b)

			if err != nil {
				p7.Log.Error("Write error: %v", err)
			}
		}()

	} else {
		if !p7.IsCoreRunning && (p7.ControlPipe == nil) {
			p7.Log.Error("P7 is not running")
		} else if p7.ControlPipe == nil {
			p7.Log.Error("OutPipe is not connected")
		}
	}
	log.Println("Sent Control")
}
