package main

import (
	"github.com/y1015860449/gotoolkit/ringBuffer"
	"github.com/y1015860449/hxnet"
)

type MyProtocol struct{}

// UnPacket 拆包
func (d *MyProtocol) UnPacket(c *hxnet.Connection, buffer *ringBuffer.RingBuffer) (interface{}, []byte) {
	s, e, length := buffer.PeekBuffer(12)
	if length >= 12 {
		if len(e) > 0 {
			size := len(s) + len(e)
			buf := make([]byte, size, size)
			copy(buf, s)
			copy(buf[len(s):], e)
			buffer.Retrieve(12)

			return nil, buf
		} else {
			buffer.Retrieve(12)
			return nil, s
		}
	}
	return nil, nil
}

// Packet 封包
func (d *MyProtocol) Packet(c *hxnet.Connection, data interface{}) []byte {
	return data.([]byte)
}
