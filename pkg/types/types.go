package types

import "time"

type OpCode byte

const (
	OpPark        OpCode = 0x01
	OpWake        OpCode = 0x02
	OpPushIntent  OpCode = 0x03
	OpRegister    OpCode = 0x04
	OpRequestPace OpCode = 0x05
	OpHeartbeat   OpCode = 0x06
)

type Purpose string

const (
	PurposeSocial   Purpose = "social"
	PurposeResearch Purpose = "research"
	PurposeBurst    Purpose = "burst"
	PurposeIdle     Purpose = "idle"
)

type RegistrationRequest struct {
	AgentID   string   `json:"agent_id"`
	Namespace [16]byte `json:"namespace"`
	Purpose   Purpose  `json:"purpose"`
}

type PaceRequest struct {
	AgentID string  `json:"agent_id"`
	Weight  float64 `json:"weight"` // Multiplier for the base purpose lambda
}

type PaceResponse struct {
	IntervalMs int64 `json:"interval_ms"`
}

type Intention struct {
	Namespace [16]byte `json:"namespace"`
	ID        [16]byte `json:"id"`
	Priority  uint8    `json:"priority"`
	Payload   []byte   `json:"payload"`
}

type SleepRequest struct {
	Namespace [16]byte      `json:"namespace"`
	Duration  time.Duration `json:"duration"`
	StateBlob []byte        `json:"state_blob"`
}

type WakeNotification struct {
	Namespace [16]byte  `json:"namespace"`
	Timestamp time.Time `json:"timestamp"`
	Reason    uint8     `json:"reason"` // 0=Timeout, 1=External Signal
}

const MagicByte byte = 0xAA

type Frame struct {
	Opcode  OpCode
	Payload []byte
}
