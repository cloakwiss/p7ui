package src

import (
	"bufio"
	"context"
	"encoding/hex"
	"net"
	"os/exec"
	"runtime"
	"time"

	"github.com/Microsoft/go-winio"
)

// HookData is assumed to be defined elsewhere, e.g., type HookData struct { Lines []string }

// readAndBatch accepts a bufio.Scanner and batches lines based on a pause (timeout).
func readAndBatch(scanner *bufio.Scanner, dataC chan<- HookData, timeout time.Duration) {

	// 1. Channel for individual lines
	// This is still necessary to decouple the blocking Scan() from the select logic.
	lineC := make(chan string)

	// 2. Start the Scanner Goroutine (Blocking I/O)
	// This goroutine performs the blocking scan and sends lines immediately.
	go func() {
		// The loop uses the provided scanner directly.
		for scanner.Scan() {
			lineC <- scanner.Text()
		}
		close(lineC) // Signal EOF/Error to the main logic

		// Optional: Check and handle scanner errors
		// if err := scanner.Err(); err != nil {
		//     // Log or send error notification
		// }
	}()

	// 3. Main Logic (Non-Blocking Batching)
	lines := make([]string, 0, 8)
	timer := time.NewTimer(timeout)

	for {
		// Reset the timer for the select statement
		var timerC <-chan time.Time
		if len(lines) > 0 {
			// Enable the timeout case only if we have buffered data.
			timer.Reset(timeout)
			timerC = timer.C
		} else {
			// If buffer is empty, disable the timer channel to wait indefinitely for the first line.
			timerC = nil
		}

		select {
		case line, isOpen := <-lineC:
			if !isOpen {
				// The scanner goroutine closed the channel (EOF/Error).
				// Send any remaining data and then exit.
				if len(lines) > 0 {
					dataC <- HookData{lines}
				}
				return
			}

			// A line arrived: add it to the buffer.
			lines = append(lines, line)

		case <-timerC:
			// Timeout hit: no new lines arrived for 'timeout' duration.
			// Send the collected burst and reset the buffer.
			dataC <- HookData{lines}
			lines = make([]string, 0, 8)
		}
	}
}

func handleClient(p7 *ApplicationState, dataC chan<- HookData, conn net.Conn) {
	p7.Log.Debug("Started handle client")

	defer conn.Close()

	var (
		buffer      = make([]byte, 1024*1024*1)
		srno   uint = 1
	)

	for {
		n, err := conn.Read(buffer)
		if n > 0 {
			dump := hex.Dump(buffer[:n])
			dataC <- HookData{srno, dump}
			srno += 1
		}
		if err != nil {
			p7.Log.Error("Read error or EOF for hook: %v\n", err)
			break
		}
	}
}

// ---------------------------------------------------------------------------------------------- //

// Spawning the core system --------------------------------------------------------------------- //
func Launch(p7 *ApplicationState, dataC chan<- HookData) {
	p7.Log.Debug("Started Launch")
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
		p7.Log.Fatal("Failed to create pipe %v", err)
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
			p7.Log.Error("Couldn't connect control pipe retrying %v", err)
			time.Sleep(500 * time.Millisecond)
		}
	}()
	// ----------------------------------------------------------------- //

	go func() {
		spawn := exec.Command(p7.TargetPath)

		stdoutPipe, err := spawn.StdoutPipe()
		if err != nil {
			p7.Log.Fatal("Failed to get stdout pipe %v", err)
			return
		}
		stderrPipe, err := spawn.StderrPipe()
		if err != nil {
			p7.Log.Fatal("Failed to get stderr pipe %v", err)
			return
		}

		if err := spawn.Start(); err != nil {
			p7.Log.Fatal("Target Spawn Failed to start %v", err)
			cancel()
		}

		p7.Log.Info("Target Output")

		// Stdout Handling
		go func() {
			scanner := bufio.NewScanner(stdoutPipe)
			for scanner.Scan() {
				p7.Log.Info("[STDOUT] %s", scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				p7.Log.Error("Stdout scan error %v", err)
			}
		}()

		// Stderr Handling
		go func() {
			scanner := bufio.NewScanner(stderrPipe)
			for scanner.Scan() {
				p7.Log.Error("[STDERR] %s", scanner.Text())
			}
			if err := scanner.Err(); err != nil {
				p7.Log.Error("Stderr scan error %v", err)
			}
		}()

		if err := spawn.Wait(); err != nil {
			p7.Log.Fatal("Target Spawn Failed %v", err)
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
				p7.Log.Info("Listener stopped: %v", err)
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
