package hxnet

import (
	"github.com/y1015860449/hxnet/poller"
	"sync"
)

type EventLoop struct {
	poll        *poller.Poller
	cLock       sync.Mutex
	connections map[int]*Connection
}

// New 创建一个 EventLoop
func New() (*EventLoop, error) {
	p, err := poller.Create()
	if err != nil {
		return nil, err
	}

	return &EventLoop{
		poll:        p,
		connections: make(map[int]*Connection),
	}, nil
}

// AddSocketAndEnableRead 增加 Socket 到事件循环中，并注册可读事件
func (e *EventLoop) AddSocketAndEnableRead(fd int, c *Connection) error {
	if err := e.poll.AddRead(fd); err != nil {
		return err
	}
	e.cLock.Lock()
	e.connections[fd] = c
	e.cLock.Unlock()
	return nil
}

// EnableReadWrite 注册可读可写事件
func (e *EventLoop) EnableReadWrite(fd int) error {
	return e.poll.EnableReadWrite(fd)
}

// EnableRead 只注册可读事件
func (e *EventLoop) EnableRead(fd int) error {
	return e.poll.EnableRead(fd)
}

func (e *EventLoop) RemoveEvent(fd int) {
	e.poll.Delete(fd)
	e.cLock.Lock()
	if _, ok := e.connections[fd]; ok {
		delete(e.connections, fd)
	}
	e.cLock.Unlock()
}

// Run 启动事件循环
func (e *EventLoop) Run() {
	e.poll.PollLoop(e.handlerEvent)
}

func (e *EventLoop) Stop() error {
	e.cLock.Lock()
	for _, v := range e.connections {
		v.Close()
	}
	e.connections = nil
	e.cLock.Unlock()
	e.poll.Stop()
	return nil
}

func (e *EventLoop) handlerEvent(fd int, event uint32) {
	e.cLock.Lock()
	c, ok := e.connections[fd]
	e.cLock.Unlock()
	if ok {
		c.HandlerEvent(fd, event)
		return
	}
}
