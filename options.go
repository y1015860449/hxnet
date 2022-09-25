package hxnet

import (
	"time"
)

// Options 服务配置
type Options struct {
	Ip       string
	Port     int
	NumLoops int
	IdleTime time.Duration
	Protocol Protocol
	Strategy LoadBalanceStrategy

	tick      time.Duration
	wheelSize int64
}

// Option ...
type Option func(*Options)

func newOptions(opt ...Option) *Options {
	opts := Options{}

	for _, o := range opt {
		o(&opts)
	}

	if opts.Ip == "" {
		opts.Ip = "0.0.0.0"
	}
	if opts.Port == 0 {
		opts.Port = 8188
	}
	if opts.tick == 0 {
		opts.tick = 1 * time.Millisecond
	}
	if opts.wheelSize == 0 {
		opts.wheelSize = 1000
	}
	if opts.Protocol == nil {
		opts.Protocol = &DefaultProtocol{}
	}
	if opts.Strategy == nil {
		opts.Strategy = RoundRobin()
	}
	if opts.IdleTime == 0 {
		opts.IdleTime = 60
	}

	return &opts
}

// Ip 监听地址
func Ip(ip string) Option {
	return func(o *Options) {
		o.Ip = ip
	}
}

func Port(port int) Option {
	return func(o *Options) {
		o.Port = port
	}
}

// NumLoops work eventloop 的数量
func NumLoops(n int) Option {
	return func(o *Options) {
		o.NumLoops = n
	}
}

// CustomProtocol 数据包处理
func CustomProtocol(p Protocol) Option {
	return func(o *Options) {
		o.Protocol = p
	}
}

// IdleTime 最大空闲时间（秒）
func IdleTime(t time.Duration) Option {
	return func(o *Options) {
		o.IdleTime = t
	}
}

func LoadBalance(strategy LoadBalanceStrategy) Option {
	return func(o *Options) {
		o.Strategy = strategy
	}
}
