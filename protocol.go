package net

import "github.com/y1015860449/gotoolkit/ringBuffer"

// Protocol 自定义协议编解码接口
type Protocol interface {
	UnPacket(c *Connection, buffer *ringBuffer.RingBuffer) (interface{}, []byte)
	Packet(c *Connection, data interface{}) []byte
}

var _ Protocol = &DefaultProtocol{}

// DefaultProtocol 默认 Protocol
type DefaultProtocol struct{}

// UnPacket 拆包
func (d *DefaultProtocol) UnPacket(c *Connection, buffer *ringBuffer.RingBuffer) (interface{}, []byte) {
	s, e := buffer.PeekAllBuffer()
	if len(e) > 0 {
		size := len(s) + len(e)
		buf := make([]byte, size, size)
		copy(buf, s)
		copy(buf[len(s):], e)
		buffer.RetrieveAll()

		return nil, buf
	} else {
		buffer.RetrieveAll()

		return nil, s
	}
}

// Packet 封包
func (d *DefaultProtocol) Packet(c *Connection, data interface{}) []byte {
	return data.([]byte)
}
