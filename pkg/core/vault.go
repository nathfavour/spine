package core

import (
	"errors"
	"fmt"
	"os"
	"sync/atomic"
	"unsafe"

	"golang.org/x/sys/unix"
)

const (
	VaultMagic   uint64 = 0x5350494E // "SPIN"
	VaultSize    int    = 4096       // Minimum page size usually
	SegmentSize  int    = 512        // Size of one intention segment
	MaxSegments  int    = VaultSize / SegmentSize

	// Offsets within a segment
	OffsetMagic     = 0x0000
	OffsetMutex     = 0x0008
	OffsetNamespace = 0x0010
	OffsetWakeTime  = 0x0020
	OffsetPayload   = 0x0028
)

type Vault struct {
	file *os.File
	data []byte
}

func OpenVault(path string) (*Vault, error) {
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return nil, err
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}

	if fi.Size() < int64(VaultSize) {
		if err := f.Truncate(int64(VaultSize)); err != nil {
			return nil, err
		}
	}

	data, err := unix.Mmap(int(f.Fd()), 0, VaultSize, unix.PROT_READ|unix.PROT_WRITE, unix.MAP_SHARED)
	if err != nil {
		f.Close()
		return nil, err
	}

	v := &Vault{file: f, data: data}
	return v, nil
}

func (v *Vault) Close() error {
	unix.Munmap(v.data)
	return v.file.Close()
}

func (v *Vault) getSegment(index int) ([]byte, error) {
	if index < 0 || index >= MaxSegments {
		return nil, errors.New("invalid segment index")
	}
	start := index * SegmentSize
	return v.data[start : start+SegmentSize], nil
}

func (v *Vault) Lock(index int) bool {
	seg, err := v.getSegment(index)
	if err != nil {
		return false
	}
	ptr := (*uint64)(unsafe.Pointer(&seg[OffsetMutex]))
	return atomic.CompareAndSwapUint64(ptr, 0, 1)
}

func (v *Vault) Unlock(index int) {
	seg, err := v.getSegment(index)
	if err != nil {
		return
	}
	ptr := (*uint64)(unsafe.Pointer(&seg[OffsetMutex]))
	atomic.StoreUint64(ptr, 0)
}

func (v *Vault) WriteState(index int, namespace [16]byte, wakeTime int64, payload []byte) error {
	seg, err := v.getSegment(index)
	if err != nil {
		return err
	}

	// Write Magic
	*(*uint64)(unsafe.Pointer(&seg[OffsetMagic])) = VaultMagic

	// Write Namespace
	copy(seg[OffsetNamespace:OffsetNamespace+16], namespace[:])

	// Write WakeTime
	*(*int64)(unsafe.Pointer(&seg[OffsetWakeTime])) = wakeTime

	// Write Payload (max 256 bytes per spec, though segment has more space)
	if len(payload) > 256 {
		payload = payload[:256]
	}
	copy(seg[OffsetPayload:OffsetPayload+256], payload)

	return nil
}

func (v *Vault) ReadState(index int) ([16]byte, int64, []byte, error) {
	seg, err := v.getSegment(index)
	if err != nil {
		return [16]byte{}, 0, nil, err
	}

	magic := *(*uint64)(unsafe.Pointer(&seg[OffsetMagic]))
	if magic != VaultMagic {
		return [16]byte{}, 0, nil, fmt.Errorf("invalid magic: %x", magic)
	}

	var namespace [16]byte
	copy(namespace[:], seg[OffsetNamespace:OffsetNamespace+16])

	wakeTime := *(*int64)(unsafe.Pointer(&seg[OffsetWakeTime]))

	payload := make([]byte, 256)
	copy(payload, seg[OffsetPayload:OffsetPayload+256])

	return namespace, wakeTime, payload, nil
}
