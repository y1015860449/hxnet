package main

import (
	"github.com/y1015860449/hxnet"
	"log"
)

func main() {
	manager := NewClientManager()
	defer manager.Release()
	svc, err := hxnet.NewServer(&netHandler{manager: manager}, net.Ip("0.0.0.0"), net.Port(6684), net.NumLoops(10), net.CustomProtocol(&MyProtocol{}))
	if err != nil {
		return
	}
	defer svc.Stop()
	if err = svc.Start(); err != nil {
		log.Print(err)
	}
}
