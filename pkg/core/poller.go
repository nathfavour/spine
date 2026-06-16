package core

import (
	"fmt"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

type Poller struct {
	epfd int
	// map timerfd to callback or channel
	timers map[int]chan struct{}
	mu     sync.Mutex
}

func NewPoller() (*Poller, error) {
	epfd, err := unix.EpollCreate1(unix.EPOLL_CLOEXEC)
	if err != nil {
		return nil, err
	}
	return &Poller{
		epfd:   epfd,
		timers: make(map[int]chan struct{}),
	}, nil
}

func (p *Poller) Close() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	for fd := range p.timers {
		unix.Close(fd)
	}
	return unix.Close(p.epfd)
}

func (p *Poller) AddTimer(duration time.Duration) (chan struct{}, error) {
	tfd, err := unix.TimerfdCreate(unix.CLOCK_MONOTONIC, unix.TFD_NONBLOCK|unix.TFD_CLOEXEC)
	if err != nil {
		return nil, err
	}

	// Set the timer
	nsec := duration.Nanoseconds()
	ts := unix.ItimerSpec{
		Value: unix.Timespec{
			Sec:  nsec / 1e9,
			Nsec: nsec % 1e9,
		},
	}
	if err := unix.TimerfdSettime(tfd, 0, &ts, nil); err != nil {
		unix.Close(tfd)
		return nil, err
	}

	ch := make(chan struct{}, 1)
	p.mu.Lock()
	p.timers[tfd] = ch
	p.mu.Unlock()

	// Add to epoll
	if err := unix.EpollCtl(p.epfd, unix.EPOLL_CTL_ADD, tfd, &unix.EpollEvent{
		Events: unix.EPOLLIN,
		Fd:     int32(tfd),
	}); err != nil {
		p.mu.Lock()
		delete(p.timers, tfd)
		p.mu.Unlock()
		unix.Close(tfd)
		return nil, err
	}

	return ch, nil
}

func (p *Poller) Wait() error {
	events := make([]unix.EpollEvent, 10)
	for {
		n, err := unix.EpollWait(p.epfd, events, -1)
		if err != nil {
			if err == unix.EINTR {
				continue
			}
			return err
		}

		for i := 0; i < n; i++ {
			fd := int(events[i].Fd)
			p.mu.Lock()
			ch, ok := p.timers[fd]
			if ok {
				// Read from timerfd to clear it
				var buf [8]byte
				unix.Read(fd, buf[:])
				
				ch <- struct{}{}
				close(ch)
				delete(p.timers, fd)
				
				// Remove from epoll and close
				unix.EpollCtl(p.epfd, unix.EPOLL_CTL_DEL, fd, nil)
				unix.Close(fd)
			}
			p.mu.Unlock()
		}
	}
}

func (p *Poller) Run() {
	if err := p.Wait(); err != nil {
		fmt.Printf("Poller error: %v\n", err)
	}
}
