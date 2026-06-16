package ipc

import (
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"

	"github.com/nathfavour/spine/pkg/types"
)

type Server struct {
	path string
	ln   net.Listener
}

func NewServer(path string) *Server {
	return &Server{path: path}
}

func (s *Server) Start(handler func(types.Frame) ([]byte, error)) error {
	if _, err := os.Stat(s.path); err == nil {
		os.Remove(s.path)
	}

	ln, err := net.Listen("unix", s.path)
	if err != nil {
		return err
	}
	s.ln = ln

	for {
		conn, err := s.ln.Accept()
		if err != nil {
			continue
		}
		go s.handleConnection(conn, handler)
	}
}

func (s *Server) handleConnection(conn net.Conn, handler func(types.Frame) ([]byte, error)) {
	defer conn.Close()

	for {
		frame, err := ReadFrame(conn)
		if err != nil {
			if err != io.EOF {
				fmt.Printf("Read error: %v\n", err)
			}
			return
		}

		resp, err := handler(frame)
		if err != nil {
			fmt.Printf("Handler error: %v\n", err)
			// Maybe send error frame back?
			continue
		}

		if resp != nil {
			// Write response if needed
			// For now, most ops might be fire-and-forget or sync
		}
	}
}

func ReadFrame(r io.Reader) (types.Frame, error) {
	var head [6]byte
	if _, err := io.ReadFull(r, head[:]); err != nil {
		return types.Frame{}, err
	}

	if head[0] != types.MagicByte {
		return types.Frame{}, fmt.Errorf("invalid magic byte: %x", head[0])
	}

	opcode := types.OpCode(head[1])
	length := binary.BigEndian.Uint32(head[2:6])

	payload := make([]byte, length)
	if _, err := io.ReadFull(r, payload); err != nil {
		return types.Frame{}, err
	}

	return types.Frame{
		Opcode:  opcode,
		Payload: payload,
	}, nil
}

func WriteFrame(w io.Writer, opcode types.OpCode, payload []byte) error {
	var head [6]byte
	head[0] = types.MagicByte
	head[1] = byte(opcode)
	binary.BigEndian.PutUint32(head[2:6], uint32(len(payload)))

	if _, err := w.Write(head[:]); err != nil {
		return err
	}
	if _, err := w.Write(payload); err != nil {
		return err
	}
	return nil
}
