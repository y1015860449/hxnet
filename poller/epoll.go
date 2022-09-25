//go:build linux
// +build linux

package poller

import (
	"github.com/y1015860449/hxnet/define"
	"log"
	"runtime"
	"sync/atomic"
	"syscall"
)

type Poller struct {
	fd      int
	running int32
}

func Create() (*Poller, error) {
	fd, err := syscall.EpollCreate(1)
	if err != nil {
		return nil, err
	}
	return &Poller{fd: fd}, nil
}

func (p *Poller) Stop() {
	p.running = atomic.SwapInt32(&p.running, 0)
	_ = syscall.Close(p.fd)
}

func (p *Poller) add(fd int, events uint32) error {
	var ev syscall.EpollEvent
	ev.Fd = int32(fd)
	ev.Events = events
	ev.Pad = 0
	return syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_ADD, fd, &ev)
}

func (p *Poller) Delete(fd int) error {
	return syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_DEL, fd, nil)
}

// AddRead 注册fd到epoll，并注册可读事件
func (p *Poller) AddRead(fd int) error {
	return p.add(fd, syscall.EPOLLIN|syscall.EPOLLPRI)
}

// AddWrite 注册fd到epoll，并注册可写事件
func (p *Poller) AddWrite(fd int) error {
	return p.add(fd, syscall.EPOLLOUT)
}

// Del 从epoll中删除fd
func (p *Poller) Del(fd int) error {
	return syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_DEL, fd, nil)
}

func (p *Poller) mod(fd int, events uint32) error {
	return syscall.EpollCtl(p.fd, syscall.EPOLL_CTL_MOD, fd, &syscall.EpollEvent{
		Events: events,
		Fd:     int32(fd),
	})
}

// EnableReadWrite 修改fd注册事件为可读可写事件
func (p *Poller) EnableReadWrite(fd int) error {
	return p.mod(fd, syscall.EPOLLIN|syscall.EPOLLPRI|syscall.EPOLLOUT)
}

// EnableWrite 修改fd注册事件为可写事件
func (p *Poller) EnableWrite(fd int) error {
	return p.mod(fd, syscall.EPOLLOUT)
}

// EnableRead 修改fd注册事件为可读事件
func (p *Poller) EnableRead(fd int) error {
	return p.mod(fd, syscall.EPOLLIN|syscall.EPOLLPRI)
}

func (p *Poller) PollLoop(handler func(fd int, event uint32)) error {
	events := make([]syscall.EpollEvent, define.WaitEvents)
	_ = atomic.SwapInt32(&p.running, 1)
	var msec int
	//在死循环中处理epoll
	for {
		n, err := syscall.EpollWait(p.fd, events, msec)
		if err != nil && err != syscall.EINTR {
			return err
		}
		if n <= 0 {
			msec = -1
			runtime.Gosched()
			continue
		}
		for i := 0; i < n; i++ {
			var event uint32
			log.Printf("PollLoop fd(%v) event(%v)", events[i].Fd, events[i].Events)
			if (events[i].Events&syscall.EPOLLHUP != 0) && (events[i].Events&syscall.EPOLLIN != 0) {
				event |= define.EventClose
			} else if events[i].Events&(syscall.EPOLLIN|syscall.EPOLLPRI|syscall.EPOLLHUP) != 0 {
				event |= define.EventRead
			} else if (events[i].Events&syscall.EPOLLOUT != 0) || (events[i].Events&syscall.EPOLLERR != 0) {
				event |= define.EventWrite
			}
			handler(int(events[i].Fd), event)
		}
		if atomic.LoadInt32(&p.running) == 0 {
			return nil
		}
		if n == len(events) {
			events = make([]syscall.EpollEvent, n+n/2)
		}
	}
	return nil
}
