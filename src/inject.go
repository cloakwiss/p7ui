package src

import (
	"errors"
	"fmt"
	"os/exec"
)

const InjecterEXE string = "../builds/debug/main.exe"

func InjectDLL(p7 *ApplicationState) {
	HookdllPath := fmt.Sprintf("-d%s", p7.HookDllPath)
	TargetPath := fmt.Sprintf("-e%s", p7.TargetPath)

	spawn := exec.Command(
		InjecterEXE,
		HookdllPath,
		TargetPath,
	)

	output, err := spawn.CombinedOutput()
	// This just any ugly hack
	p7.Log.InfoWithPayload("InjectDLL Output:", errors.New(string(output)))
	if err != nil {
		p7.Log.FatalWithPayload("InjectDLL Spawn Failed for some reason", err)
	}
}

func RemoveDLL(p7 *ApplicationState) {
	TargetPath := fmt.Sprintf("-e%s", p7.TargetPath)
	spawn := exec.Command(
		InjecterEXE,
		TargetPath,
		"-r",
	)

	output, err := spawn.CombinedOutput()
	p7.Log.InfoWithPayload("RemoveDLL Output:", errors.New(string(output)))
	if err != nil {
		p7.Log.FatalWithPayload("RemoveDLL Spawn Failed for some reason %v", err)
	}
}
