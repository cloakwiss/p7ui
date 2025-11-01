package src

import (
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

// // deserializes the buffer accoriding to the id and adds hook return
// // to the Return list in hooks
// func (Hooks *HookList) AddReturn(p7 *ApplicationState, buffer []byte) {
// 	head := int(0)
// 	var ret HookReturns

// 	depth, err := deserialize.Decode(buffer, &head, uint64(0))
// 	if err != nil {
// 		p7.Log.Error("Desirialization of call depth failed: %v", err)
// 	} else {
// 		p7.Log.Debug("Call depth :%v", depth.(uint64))
// 	}

// 	timing, err := deserialize.Decode(buffer, &head, float64(0))
// 	if err != nil {
// 		p7.Log.Error("Desirialization of return timing failed: %v", err)
// 	} else {
// 		p7.Log.Debug("Return time :%v", timing.(float64))
// 	}

// 	id, err := deserialize.Decode(buffer, &head, "")
// 	if err != nil {
// 		p7.Log.Error("Desirialization of return id failed: %v", err)
// 	} else {
// 		p7.Log.Debug("Return id :%s", id.(string))
// 	}

// 	returns, err := GetReturnStructure(id.(string))
// 	if err != nil {
// 		p7.Log.Error("Construction of return Value List failed for this id %s, %v", id, err)
// 	}

// 	err = deserialize.DecodeValue(&returns, buffer, &head)
// 	if err != nil {
// 		p7.Log.Error("Desirialization of returns failed: %v", err)
// 	}

// 	ret.id = id.(string)
// 	ret.depth = depth.(uint64)
// 	ret.returns = returns
// 	p7.Log.Info("Return struct:\n%v", ret)
// 	Hooks.ReturnList = append(Hooks.ReturnList, ret)
// }
