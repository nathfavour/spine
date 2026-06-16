package types

import "time"

type OpCode byte

const (
	OpPark       OpCode = 0x01
	OpWake       OpCode = 0x02
	OpPushIntent OpCode = 0x03
)

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
