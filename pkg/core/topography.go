package core

import (
	"runtime"
	"syscall"

	"golang.org/x/sys/unix"
)

// SetCPUAffinity pins the current goroutine/thread to specific CPUs.
// Note: In Go, we should call runtime.LockOSThread() before calling this.
func SetCPUAffinity(cpus []int) error {
	var mask unix.CPUSet
	mask.Zero()
	for _, cpu := range cpus {
		mask.Set(cpu)
	}
	return unix.SchedSetaffinity(0, &mask)
}

const (
	IOPRIO_CLASS_NONE = 0
	IOPRIO_CLASS_RT   = 1
	IOPRIO_CLASS_BE   = 2
	IOPRIO_CLASS_IDLE = 3

	IOPRIO_WHO_PROCESS = 1
	IOPRIO_WHO_PGRP    = 2
	IOPRIO_WHO_USER    = 3
)

func SetIOPriority(class, priority int) error {
	// ioprio = (class << 13) | priority
	ioprio := (class << 13) | (priority & 0x1fff)
	_, _, err := syscall.Syscall(unix.SYS_IOPRIO_SET, uintptr(IOPRIO_WHO_PROCESS), 0, uintptr(ioprio))
	if err != 0 {
		return err
	}
	return nil
}

func PinToEfficiencyCores() error {
	// This is a heuristic. On many modern systems, higher-indexed cores are often efficiency cores.
	// A better way would be to parse /proc/cpuinfo or use lscpu, but for now we'll allow the user to specify
	// or provide a helper to guess.
	numCPU := runtime.NumCPU()
	if numCPU <= 4 {
		return nil // Not enough cores to safely pin
	}
	
	// Assume the last half are E-cores (common on Intel Alder Lake and later)
	var eCores []int
	for i := numCPU / 2; i < numCPU; i++ {
		eCores = append(eCores, i)
	}
	
	runtime.LockOSThread()
	return SetCPUAffinity(eCores)
}
