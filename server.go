package hxnet

import (
	"errors"
	"github.com/RussellLuo/timingwheel"
	"log"
	"runtime"
	"time"
)

// Handler Server 注册接口
type Handler interface {
	ConnectionCallBack
	OnConnect(c *Connection)
}

type Server struct {
	listener    *listener
	eventLoops  []*EventLoop
	callback    Handler
	timingWheel *timingwheel.TimingWheel
	opts        *Options
}

// NewServer 创建 Server
func NewServer(handler Handler, opts ...Option) (server *Server, err error) {
	if handler == nil {
		return nil, errors.New("handler is nil")
	}
	options := newOptions(opts...)
	server = new(Server)
	server.callback = handler
	server.opts = options
	server.timingWheel = timingwheel.NewTimingWheel(server.opts.tick, server.opts.wheelSize)

	if server.opts.NumLoops <= 0 {
		server.opts.NumLoops = runtime.NumCPU()
	}

	eloops := make([]*EventLoop, server.opts.NumLoops)
	for i := 0; i < server.opts.NumLoops; i++ {
		loop, err := New()
		if err != nil {
			for j := 0; j < i; j++ {
				_ = eloops[j].Stop()
			}
			return nil, err
		}
		eloops[i] = loop
	}
	server.eventLoops = eloops

	server.listener, err = newListener(server.opts.Ip, server.opts.Port, server.handleNewConnection)
	if err != nil {
		return nil, err
	}
	return
}

func (s *Server) handleNewConnection(fd int) {
	loop := s.opts.Strategy(s.eventLoops)
	c := NewConnection(fd, loop, s.opts.Protocol, s.timingWheel, s.opts.IdleTime*time.Second, s.callback)
	s.callback.OnConnect(c)
	if err := loop.AddSocketAndEnableRead(fd, c); err != nil {
		log.Printf("handleNewConnection AddSocketAndEnableRead err(%v)", err)
	}
}

// Start 启动 Server
func (s *Server) Start() error {
	s.timingWheel.Start()
	for i := 0; i < s.opts.NumLoops; i++ {
		go s.eventLoops[i].Run()
	}
	return s.listener.Run()
}

// Stop 关闭 Server
func (s *Server) Stop() {
	s.timingWheel.Stop()
	if err := s.listener.Stop(); err != nil {
		log.Print(err)
	}

	for k := range s.eventLoops {
		if err := s.eventLoops[k].Stop(); err != nil {
			log.Print(err)
		}
	}
	s.listener.Stop()
}
