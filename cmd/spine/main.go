package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/nathfavour/spine/pkg/core"
	"github.com/nathfavour/spine/pkg/ipc"
	"github.com/nathfavour/spine/pkg/types"
)

func main() {
	fmt.Println("spine: Agentic Foundation Engine starting...")

	// 1. Topography Discipline
	if err := core.PinToEfficiencyCores(); err != nil {
		log.Printf("Warning: Failed to pin to efficiency cores: %v", err)
	}
	if err := core.SetIOPriority(core.IOPRIO_CLASS_IDLE, 7); err != nil {
		log.Printf("Warning: Failed to set IO priority: %v", err)
	}

	// 2. Initialize Vault
	tmpDir := "/tmp/spine"
	os.MkdirAll(tmpDir, 0755)
	vaultPath := filepath.Join(tmpDir, "vault.mmap")
	vault, err := core.OpenVault(vaultPath)
	if err != nil {
		log.Fatalf("Failed to open vault: %v", err)
	}
	defer vault.Close()

	// 3. Initialize Poller
	poller, err := core.NewPoller()
	if err != nil {
		log.Fatalf("Failed to initialize poller: %v", err)
	}
	defer poller.Close()
	go poller.Run()

	// 4. IPC Server
	socketPath := "/tmp/spine.sock"
	
	// We need a slightly modified handler that can send delayed responses
	ln, err := net.Listen("unix", socketPath)
	if err != nil {
		log.Fatalf("Failed to listen: %v", err)
	}
	os.Chmod(socketPath, 0666)

	for {
		conn, err := ln.Accept()
		if err != nil {
			continue
		}
		go handleConnection(conn, vault, poller)
	}
}

func handleConnection(conn net.Conn, vault *core.Vault, poller *core.Poller) {
	defer conn.Close()

	for {
		frame, err := ipc.ReadFrame(conn)
		if err != nil {
			return
		}

		switch frame.Opcode {
		case types.OpPark:
			var req types.SleepRequest
			if err := json.Unmarshal(frame.Payload, &req); err != nil {
				return
			}

			fmt.Printf("Parking namespace: %s for %v\n", req.Namespace, req.Duration)

			index := int(req.Namespace[0]) % core.MaxSegments
			wakeTime := time.Now().Add(req.Duration).UnixNano()
			vault.WriteState(index, req.Namespace, wakeTime, req.StateBlob)

			ch, err := poller.AddTimer(req.Duration)
			if err != nil {
				return
			}

			// Wait for timer and then notify back on the SAME connection
			<-ch
			
			resp := types.WakeNotification{
				Namespace: req.Namespace,
				Timestamp: time.Now(),
				Reason:    0,
			}
			data, _ := json.Marshal(resp)
			ipc.WriteFrame(conn, types.OpWake, data)
			return // One park per connection usually

		case types.OpPushIntent:
			var intent types.Intention
			json.Unmarshal(frame.Payload, &intent)
			index := int(intent.Namespace[0]) % core.MaxSegments
			vault.WriteState(index, intent.Namespace, 0, intent.Payload)
			// Acknowledge?
		}
	}
}
