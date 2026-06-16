package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/nathfavour/spine/pkg/ipc"
	"github.com/nathfavour/spine/pkg/types"
)

func main() {
	socketPath := "/tmp/spine.sock"
	conn, err := net.Dial("unix", socketPath)
	if err != nil {
		log.Fatalf("Failed to connect to spine: %v", err)
	}
	defer conn.Close()

	agentID := "example-agent-01"

	// 1. Register OTA
	reg := types.RegistrationRequest{
		AgentID: agentID,
		Purpose: types.PurposeSocial,
	}
	regData, _ := json.Marshal(reg)
	fmt.Println("Registering as Social agent...")
	if err := ipc.WriteFrame(conn, types.OpRegister, regData); err != nil {
		log.Fatalf("Registration failed: %v", err)
	}

	// 2. Request Pacing in a loop
	for i := 1; i <= 3; i++ {
		fmt.Printf("Action %d: Requesting human-mimic pace...\n", i)
		paceReq := types.PaceRequest{
			AgentID: agentID,
			Weight:  1.0,
		}
		paceData, _ := json.Marshal(paceReq)
		if err := ipc.WriteFrame(conn, types.OpRequestPace, paceData); err != nil {
			log.Fatalf("Pace request failed: %v", err)
		}

		// Wait for response
		frame, err := ipc.ReadFrame(conn)
		if err != nil {
			log.Fatalf("Read failed: %v", err)
		}

		var paceResp types.PaceResponse
		json.Unmarshal(frame.Payload, &paceResp)
		
		interval := time.Duration(paceResp.IntervalMs) * time.Millisecond
		fmt.Printf("Spine says wait: %v. Mimicking human delay...\n", interval)
		time.Sleep(interval)
		
		fmt.Println("Performing action...")
	}

	fmt.Println("Done.")
}
