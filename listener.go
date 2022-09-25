package hxnet

import (
	"net"
	"sync/atomic"
	"syscall"
)

type handleConnected func(fd int)

type listener struct {
	sfd     int
	efd     int
	hc      handleConnected
	running int32
}

func newListener(ip string, port int, hc handleConnected) (*listener, error) {
	// 创建socket
	fd, err := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	if err != nil {
		// todo 添加日志
		return nil, err
	}
	_ = syscall.SetsockoptInt(fd, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, 1)
	_ = syscall.SetNonblock(fd, true)
	// 绑定ip和port
	var addr [4]byte
	copy(addr[:], net.ParseIP(ip).To4())
	if err = syscall.Bind(fd, &syscall.SockaddrInet4{
		Port: port,
		Addr: addr,
	}); err != nil {
		// todo 添加日志
		return nil, err
	}
	// 监听
	if err = syscall.Listen(fd, 5); err != nil {
		// todo 添加日志
		return nil, err
	}
	return &listener{sfd: fd, hc: hc}, nil
}

func (l *listener) Stop() error {
	_ = atomic.SwapInt32(&l.running, 0)
	_ = syscall.Close(l.efd)
	_ = syscall.Close(l.sfd)
	return nil
}

func (l *listener) Run() error {
	//return l.Accept()
	return l.listenerLoop()
}

func (l *listener) Accept() error {
	for {
		nfd, _, err := syscall.Accept(l.sfd)
		if err != nil {
			if err != syscall.EAGAIN {
				return err
			}
			continue
		}
		_ = syscall.SetNonblock(nfd, true)
		l.hc(nfd)
	}
}

func (l *listener) listenerLoop() error {
	efd, err := syscall.EpollCreate(1)
	if err != nil {
		return err
	}
	l.efd = efd
	var event syscall.EpollEvent
	event.Events = syscall.EPOLLIN | syscall.EPOLLPRI
	event.Fd = int32(l.sfd) //设置监听描述符
	if err = syscall.EpollCtl(efd, syscall.EPOLL_CTL_ADD, l.sfd, &event); err != nil {
		return err
	}
	_ = atomic.SwapInt32(&l.running, 1)
	events := make([]syscall.EpollEvent, 10)
	var msec int
	for {
		//msec -1,会一直阻塞,直到有事件可以处理才会返回, n 事件个数
		n, err := syscall.EpollWait(l.efd, events, msec)
		if err != nil && err != syscall.EINTR {
			return err
		}
		if n <= 0 {
			msec = -1
			continue
		}
		for i := 0; i < n; i++ {
			if int(events[i].Fd) == l.sfd {
				nfd, _, err := syscall.Accept(l.sfd)
				if err != nil {
					if err != syscall.EAGAIN {
						return err
					}
					continue
				}
				_ = syscall.SetNonblock(nfd, true)
				l.hc(nfd)
			}
		}
		if atomic.LoadInt32(&l.running) == 0 {
			return nil
		}
	}
	return nil
}
