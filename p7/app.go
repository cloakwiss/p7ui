package p7

import (
	// "encoding/hex"
	"errors"
	"fmt"
	"net"

	deserialize "github.com/cloakwiss/project-seven/deserialize"
)

type Page string

const (
	IndexPage  Page = "index.html"
	ReportPage Page = "report.html"
	ErrorPage  Page = "error.html"
)

type ApplicationState struct {
	TargetPath      string
	HookDllPath     string
	IsCoreRunning   bool
	HookPipeName    string
	ControlPipeName string
	LogPipeName     string
	StepState       bool
	HookChannel     chan<- HookData
	Log             Logger
	ControlPipe     net.Conn
	Page            Page
	Hooks           HookList
}

type HookList struct {
	CallList   []HookCall
	ReturnList []HookReturns
}

type HookCall struct {
	id    string
	depth uint64
	args  []deserialize.Values
}

type HookReturns struct {
	id      string
	depth   uint64
	time    float64
	returns []deserialize.Values
}

type Control byte

const (
	Start  Control = 0x21
	Stop   Control = 0x22
	Resume Control = 0x23
	Abort  Control = 0x24
	STEC   Control = 0x25
	STSC   Control = 0x26
)

const (
	HOOK_CALL_ID byte = 0x82
	HOOK_RET_ID  byte = 0x28
)

// Send Control Signal to hool dll
//
//	p7.Start  0x21
//	p7.Stop   0x22
//	p7.Resume 0x23
//	p7.Abort  0x24
//	p7.STEC   0x25
//	p7.STSC   0x26
func (p7 *ApplicationState) SendControl(controlSignal Control) {
	if p7.ControlPipe != nil {
		b := []byte{byte(controlSignal)}
		go func() {
			_, err := p7.ControlPipe.Write(b)

			if err != nil {
				p7.Log.Error("Write error: %v\n", err)
			}
			// } else {
			// 	p7.Log.Debug("Wrote Signal %d", controlSignal)
			// }
		}()

	} else {
		if !p7.IsCoreRunning && (p7.ControlPipe == nil) {
			p7.Log.Error("P7 is not running")
		} else if p7.ControlPipe == nil {
			p7.Log.Error("OutPipe is not connected")
		}
	}
}

// --------------------------------------------------------------------------------------------- //
// Template Getters (to be implemented generally for all funcs) -------------------------------- //
// --------------------------------------------------------------------------------------------- //
func GetCallStructure(id string) ([]deserialize.Values, error) {
	switch id {
	case "MessageBoxA":
		{
			args := make([]deserialize.Values, 4)

			args[0].Name = "hWnd"
			args[0].Val = uint64(0)

			args[1].Name = "lpText"
			args[1].Val = string("")

			args[2].Name = "lpCaption"
			args[2].Val = string("")

			args[3].Name = "uType"
			args[3].Val = uint32(0)

			return args, nil
		}

	case "Sleep":
		{
			args := make([]deserialize.Values, 1)

			args[0].Name = "dwMilliseconds"
			args[0].Val = uint32(0)

			return args, nil
		}

	default:
		{
			return nil, errors.New("Unimplemented function id")
		}
	}
}

func GetReturnStructure(id string) ([]deserialize.Values, error) {
	switch id {
	case "MessageBoxA":
		{
			args := make([]deserialize.Values, 1)

			args[0].Name = "result"
			args[0].Val = int32(0)

			return args, nil
		}

	case "Sleep":
		{
			args := make([]deserialize.Values, 0)
			return args, nil
		}

	default:
		{
			return nil, errors.New("Unimplemented function id")
		}
	}
}

// --------------------------------------------------------------------------------------------- //
// Add Hooks ----------------------------------------------------------------------------------- //
// --------------------------------------------------------------------------------------------- //

// deserializes the buffer accoriding to the id and adds hook call
// to the Call list in hooks
func (Hooks *HookList) AddCall(p7 *ApplicationState, buffer []byte) {
	head := int(0)
	var call HookCall

	depth, err := deserialize.Decode(buffer, &head, uint64(0))
	if err != nil {
		p7.Log.Error("Desirialization of call depth failed: %v", err)
	} else {
		p7.Log.Debug("Call depth :%v", depth.(uint64))
	}

	id, err := deserialize.Decode(buffer, &head, "")
	if err != nil {
		p7.Log.Error("Desirialization of call id failed: %v", err)
	} else {
		p7.Log.Debug("Call id :%s", id.(string))
	}

	args, err := GetCallStructure(id.(string))
	if err != nil {
		p7.Log.Error("Construction of args Value List failed for this id %s, %v", id, err)
	}

	err = deserialize.DecodeValue(&args, buffer, &head)
	if err != nil {
		p7.Log.Error("Desirialization of call args failed: %v", err)
	}

	call.id = id.(string)
	call.depth = depth.(uint64)
	call.args = args

	p7.Log.Info("Call struct:\n%v", call)

	callDataString := fmt.Sprintf("%v", call)
	p7.HookChannel <- HookData{len(Hooks.CallList), callDataString}

	Hooks.CallList = append(Hooks.CallList, call)
}

// deserializes the buffer accoriding to the id and adds hook return
// to the Return list in hooks
func (Hooks *HookList) AddReturn(p7 *ApplicationState, buffer []byte) {
	head := int(0)
	var ret HookReturns

	depth, err := deserialize.Decode(buffer, &head, uint64(0))
	if err != nil {
		p7.Log.Error("Desirialization of call depth failed: %v", err)
	} else {
		p7.Log.Debug("Call depth :%v", depth.(uint64))
	}

	timing, err := deserialize.Decode(buffer, &head, float64(0))
	if err != nil {
		p7.Log.Error("Desirialization of return timing failed: %v", err)
	} else {
		p7.Log.Debug("Return time :%v", timing.(float64))
	}

	id, err := deserialize.Decode(buffer, &head, "")
	if err != nil {
		p7.Log.Error("Desirialization of return id failed: %v", err)
	} else {
		p7.Log.Debug("Return id :%s", id.(string))
	}

	returns, err := GetReturnStructure(id.(string))
	if err != nil {
		p7.Log.Error("Construction of return Value List failed for this id %s, %v", id, err)
	}

	err = deserialize.DecodeValue(&returns, buffer, &head)
	if err != nil {
		p7.Log.Error("Desirialization of returns failed: %v", err)
	}

	ret.id = id.(string)
	ret.depth = depth.(uint64)
	ret.returns = returns

	p7.Log.Info("Return struct:\n%v", ret)

	retDataString := fmt.Sprintf("%v", ret)
	p7.HookChannel <- HookData{len(Hooks.ReturnList), retDataString}

	Hooks.ReturnList = append(Hooks.ReturnList, ret)
}

// deserializes the buffer and adds the hook into the hooks list
func (p7 *ApplicationState) AddHook(buffer []byte) {
	// p7.Log.Debug("Just Hook Buffer: \n%v", hex.Dump(buffer[1:]))

	switch buffer[0] {
	case HOOK_CALL_ID:
		{
			p7.Hooks.AddCall(p7, buffer[1:])
			break
		}

	case HOOK_RET_ID:
		{
			p7.Hooks.AddReturn(p7, buffer[1:])
			break
		}

	default:
		{
			p7.Log.Fatal("Neither call nor return id found in the first byte in buffer")
			break
		}
	}
}

// --------------------------------------------------------------------------------------------- //
