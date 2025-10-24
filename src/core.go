package src

import (
	"bufio"
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"time"

	"github.com/Microsoft/go-winio"
)

func handleClient(p7 *ApplicationState, dataC chan<- HookData, conn any) {

	defer func() {
		if c, ok := conn.(interface{ Close() error }); ok {
			c.Close()
		}
	}()

	reader, ok := conn.(interface{ Read([]byte) (int, error) })
	if !ok {
		p7.Log.Error("Invalid connection type")
		return
	}

	scanner := bufio.NewScanner(reader)
	lines := make([]string, 0, 8)
	for scanner.Scan() {
		text := scanner.Text()
		lines = append(lines, text)
	}

	dataC <- HookData{lines}

	if err := scanner.Err(); err != nil {
		p7.Log.ErrorWithPayload("Read error", err)
	}
}

// ---------------------------------------------------------------------------------------------- //

// Spawning the core system --------------------------------------------------------------------- //
func Launch(p7 *ApplicationState, dataC chan<- HookData) {
	InjectDLL(p7)
	defer RemoveDLL(p7)

	ctx, cancel := context.WithCancel(context.Background())

	// Listener to read from the InPipe -------------------------------- //
	pipeCfg := &winio.PipeConfig{
		SecurityDescriptor: "",
		MessageMode:        true,
		InputBufferSize:    256,
		OutputBufferSize:   256,
	}

	listener, err := winio.ListenPipe(p7.InPipeName, pipeCfg)
	if err != nil {
		p7.Log.FatalWithPayload("Failed to create pipe", err)
	}
	defer listener.Close()

	// Waiting for Target to spawn the listener for controls ----------- //
	notEnded := true
	go func() {
		for p7.OutPipe == nil && notEnded {
			timeout := 5 * time.Second
			var err error
			p7.OutPipe, err = winio.DialPipe(p7.OutPipeName, &timeout)
			if err == nil {
				p7.Log.Info("Connected the control pipe")
				break
			}
			p7.Log.ErrorWithPayload("Couldn't connect control pipe retrying", err)
			time.Sleep(500 * time.Millisecond)
		}
	}()
	// ----------------------------------------------------------------- //

	go func() {
		spawn := exec.Command(p7.TargetPath)

		stdoutPipe, err := spawn.StdoutPipe()
		if err != nil {
			p7.Log.FatalWithPayload("Failed to get stdout pipe", err)
			return
		}
		stderrPipe, err := spawn.StderrPipe()
		if err != nil {
			p7.Log.FatalWithPayload("Failed to get stderr pipe", err)
			return
		}

		if err := spawn.Start(); err != nil {
			p7.Log.FatalWithPayload("Target Spawn Failed to start", err)
			cancel()
		}

		p7.Log.Info("Target Output")

		// Stdout Handling
		go func() {
			scanner := bufio.NewScanner(stdoutPipe)
			for scanner.Scan() {
				p7.Log.Info(fmt.Sprintf("[STDOUT] %s", scanner.Text()))
			}
			if err := scanner.Err(); err != nil {
				p7.Log.ErrorWithPayload("Stdout scan error", err)
			}
		}()

		// Stderr Handling
		go func() {
			scanner := bufio.NewScanner(stderrPipe)
			for scanner.Scan() {
				p7.Log.Error(fmt.Sprintf("[STDERR] %s", scanner.Text()))
			}
			if err := scanner.Err(); err != nil {
				p7.Log.ErrorWithPayload("Stderr scan error", err)
			}
		}()

		if err := spawn.Wait(); err != nil {
			p7.Log.FatalWithPayload("Target Spawn Failed", err)
		}

		cancel()
	}()

	p7.Log.Info("Waiting for Hook DLL...")
	go func() {
		runtime.LockOSThread()
		for {
			p7.Log.Debug("Looking for hook senders")
			conn, err := listener.Accept()
			if err != nil {
				p7.Log.InfoWithPayload("Listener stopped", err)
				return
			} else {
				p7.Log.Debug("Listener Connected")
			}

			handleClient(p7, dataC, conn)
		}
	}()

	p7.IsCoreRunning = true

	<-ctx.Done()
	notEnded = false
	p7.IsCoreRunning = false
	if p7.OutPipe != nil {
		p7.OutPipe.Close()
		p7.OutPipe = nil
	}
}
