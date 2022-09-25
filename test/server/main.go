package main

import (
	"github.com/y1015860449/hxnet"
	"log"
)

func main() {
	manager := NewClientManager()
	svc, err := hxnet.NewServer(&netHandler{manager: manager}, hxnet.Ip("0.0.0.0"), hxnet.Port(6684), hxnet.NumLoops(10), hxnet.CustomProtocol(&MyProtocol{}))
	if err != nil {
		return
	}
	defer svc.Stop()
	if err = svc.Start(); err != nil {
		log.Print(err)
	}
}
