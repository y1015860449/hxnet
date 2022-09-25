package main

import (
	"github.com/y1015860449/gotoolkit/coPool"
	"github.com/y1015860449/hxnet"
	"math/rand"
	"sync"
	"time"
)

type ClientManager struct {
	cliLock sync.Mutex
	clients map[string]*hxnet.Connection
	addrs   []string
	rtPool  *coPool.SharePool
}

func NewClientManager() *ClientManager {
	rand.Seed(time.Now().UnixNano())
	manager := &ClientManager{
		clients: make(map[string]*hxnet.Connection),
	}
	manager.rtPool = coPool.NewSharePool(5, 2, 1024, 30*time.Second)

	return manager
}

func (p *ClientManager) AddConnection(c *hxnet.Connection) {
	p.cliLock.Lock()
	p.clients[c.PeerAddr()] = c
	p.addrs = append(p.addrs, c.PeerAddr())
	p.cliLock.Unlock()
}

func (p *ClientManager) DeleteConnection(c *hxnet.Connection) {
	p.cliLock.Lock()
	delete(p.clients, c.PeerAddr())
	for i, v := range p.addrs {
		if c.PeerAddr() == v {
			p.addrs = p.addrs[:i+copy(p.addrs[i:], p.addrs[i+1:])]
		}
	}
	p.cliLock.Unlock()
}

func (p *ClientManager) Invoke(data interface{}) {
	p.rtPool.Submit(data, p.handlerMessage)
}

func (p *ClientManager) handlerMessage(data interface{}) error {
	message := data.([]byte)
	if len(p.addrs) > 0 {
		index := rand.Intn(len(p.addrs))
		p.cliLock.Lock()
		c, ok := p.clients[p.addrs[index]]
		p.cliLock.Unlock()
		if ok {
			c.Send(message)
		}
	}
	return nil
}
