package main

import (
	"github.com/y1015860449/hxnet"
	"log"
)

type netHandler struct {
	manager *ClientManager
}

func (h *netHandler) OnMessage(c *hxnet.Connection, ctx interface{}, data []byte) interface{} {
	log.Printf("OnMessage data(%v)", string(data))
	h.manager.Invoke(data)
	return data
}

func (h *netHandler) OnClose(c *hxnet.Connection) {
	log.Printf("OnClose addr(%v)", c.PeerAddr())
	h.manager.DeleteConnection(c)
}

func (h *netHandler) OnConnect(c *hxnet.Connection) {
	log.Printf("OnConnect addr(%v)", c.PeerAddr())
	h.manager.AddConnection(c)
}
