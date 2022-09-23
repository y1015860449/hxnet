package net

import (
	"github.com/RussellLuo/timingwheel"
	"github.com/y1015860449/gotoolkit/ringBuffer"
	"github.com/y1015860449/hxnet/define"
	"log"
	"net"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"
)

type ConnectionCallBack interface {
	OnMessage(c *Connection, ctx interface{}, data []byte) interface{}
	OnClose(c *Connection)
}

type Connection struct {
	fd        int
	outBuffer *ringBuffer.RingBuffer // write buffer
	inBuffer  *ringBuffer.RingBuffer // read buffer
	loop      *EventLoop
	callBack  ConnectionCallBack
	protocol  Protocol
	ctx       interface{}

	idleTime    time.Duration            // 空闲时间
	timingWheel *timingwheel.TimingWheel // 时间轮
	activeTime  int64
	timer       atomic.Value
	peerAddr    string
}

func NewConnection(fd int,
	loop *EventLoop,
	protocol Protocol,
	tw *timingwheel.TimingWheel,
	idleTime time.Duration,
	callBack ConnectionCallBack) *Connection {
	conn := &Connection{
		fd:          fd,
		loop:        loop,
		timingWheel: tw,
		protocol:    protocol,
		idleTime:    idleTime,
		callBack:    callBack,
		inBuffer:    ringBuffer.NewBuffer(4096),
		outBuffer:   ringBuffer.NewBuffer(4096),
	}
	conn.peerAddr, _ = conn.getPeerAddr()
	if conn.idleTime > 0 {
		_ = atomic.SwapInt64(&conn.activeTime, time.Now().Unix())
		timer := conn.timingWheel.AfterFunc(conn.idleTime, conn.closeTimeoutConn())
		conn.timer.Store(timer)
	}
	return conn
}

// Context 获取 Context
func (c *Connection) Context() interface{} {
	return c.ctx
}

// SetContext 设置 Context
func (c *Connection) SetContext(ctx interface{}) {
	c.ctx = ctx
}

// Send 发送
func (c *Connection) Send(data interface{}) {
	c.sendData(c.protocol.Packet(c, data))
}

func (c *Connection) closeTimeoutConn() func() {
	return func() {
		now := time.Now()
		intervals := now.Sub(time.Unix(atomic.LoadInt64(&c.activeTime), 0))
		if intervals >= c.idleTime {
			log.Printf("closeTimeoutConn")
			c.handlerClose(c.fd)
		} else {
			timer := c.timingWheel.AfterFunc(c.idleTime-intervals, c.closeTimeoutConn())
			c.timer.Store(timer)
		}
	}
}

func (c *Connection) HandlerEvent(fd int, event uint32) {
	if c.idleTime > 0 {
		_ = atomic.SwapInt64(&c.activeTime, time.Now().Unix())
	}
	if event&define.EventClose != 0 {
		c.handlerClose(fd)
		return
	}
	if !c.outBuffer.IsEmpty() {
		if event&define.EventWrite != 0 {
			c.handlerWrite(fd)
		}
	} else if event&define.EventRead != 0 {
		c.handlerRead(fd)
	}
}

func (c *Connection) PeerAddr() string {
	return c.peerAddr
}

func (c *Connection) Close() {
	log.Printf("Close")
	c.handlerClose(c.fd)
}

func (c *Connection) handlerRead(fd int) {
	log.Printf("handlerRead")
	for {
		buf := make([]byte, 8192)
		n, err := syscall.Read(fd, buf)
		log.Printf("handlerRead Read n(%v) err(%v)", n, err)
		if err != nil && err != syscall.EAGAIN {
			log.Printf("handlerRead close")
			c.handlerClose(fd)
			return
		}
		if n == 0 { // 客户端主动关闭
			log.Printf("handlerRead2 close")
			c.handlerClose(fd)
			return
		} else if n < 0 {
			break
		} else {
			c.inBuffer.WriteBuffer(buf[0:n])
		}

	}
	if c.inBuffer.IsEmpty() {
		return
	}
	outBuf := c.handlerProtocol(c.inBuffer)
	if len(outBuf) > 0 {
		n, err := c.writeByte(c.fd, outBuf)
		if err == nil {
			if n < len(outBuf) {
				c.outBuffer.WriteBuffer(outBuf[n:])
			}
			if !c.outBuffer.IsEmpty() {
				_ = c.loop.EnableReadWrite(c.fd)
			}
		}
	}
}

func (c *Connection) handlerWrite(fd int) {
	log.Printf("handlerWrite")
	f, s := c.outBuffer.PeekAllBuffer()
	n, err := c.writeByte(fd, f)
	if err != nil || n == 0 {
		return
	}
	c.outBuffer.Retrieve(n)
	if len(f) == n && len(s) > 0 {
		n, err = c.writeByte(fd, s)
		if err != nil || n == 0 {
			return
		}
		c.outBuffer.Retrieve(n)
	}
	if c.outBuffer.IsEmpty() {
		_ = c.loop.EnableRead(fd)
	}
}

func (c *Connection) writeByte(fd int, data []byte) (int, error) {
	n, err := syscall.Write(fd, data)
	if err != nil {
		if err == syscall.EAGAIN {
			return 0, nil
		}
		log.Printf("writeByte close")
		c.handlerClose(fd)
		return 0, err
	}
	return n, nil
}

func (c *Connection) handlerClose(fd int) {
	log.Printf("handlerClose")
	c.callBack.OnClose(c)
	c.loop.RemoveEvent(fd)
	_ = syscall.Close(fd)
	if v := c.timer.Load(); v != nil {
		timer := v.(*timingwheel.Timer)
		timer.Stop()
	}
}

func (c *Connection) handlerProtocol(inBuffer *ringBuffer.RingBuffer) []byte {
	var outBuffer []byte
	for {
		ctx, rcvData := c.protocol.UnPacket(c, inBuffer)
		if ctx == nil && len(rcvData) <= 0 {
			break
		}
		sendData := c.callBack.OnMessage(c, ctx, rcvData)
		if sendData != nil {
			outBuffer = append(outBuffer, c.protocol.Packet(c, sendData)...)
		}
	}
	return outBuffer
}

func (c *Connection) getPeerAddr() (string, error) {
	sa, err := syscall.Getpeername(c.fd)
	if err != nil {
		return "", err
	}
	ipv4Sa := sa.(*syscall.SockaddrInet4)
	return net.JoinHostPort(net.IP(ipv4Sa.Addr[:]).String(), strconv.Itoa(ipv4Sa.Port)), nil
}

func (c *Connection) sendData(data []byte) {
	if !c.outBuffer.IsEmpty() {
		_, _ = c.outBuffer.WriteBuffer(data)
	} else {
		n, err := c.writeByte(c.fd, data)
		if err != nil {
			return
		}
		if n <= 0 {
			_, _ = c.outBuffer.WriteBuffer(data)
		} else if n < len(data) {
			_, _ = c.outBuffer.WriteBuffer(data[n:])
		}
		if !c.outBuffer.IsEmpty() {
			_ = c.loop.EnableReadWrite(c.fd)
		}
	}
}
