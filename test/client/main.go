package main

import (
	"fmt"
	"net"
	"time"
)

func main() {
	for i := 0; i < 10000; i++ {
		go func() {
			conn, err := net.Dial("tcp", "192.168.20.129:6684")
			if err != nil {
				fmt.Println(err)
				return
			}
			//go func() {
			//	data := make([]byte, 100)
			//	for {
			//		_, err = conn.Read(data)
			//		if err != nil {
			//			fmt.Println(conn.LocalAddr(), err)
			//			break
			//			// panic(err)
			//		}
			//	}
			//}()
			//for {
			//	if _, err = conn.Write([]byte("hello world!")); err != nil {
			//		fmt.Println(conn.LocalAddr(), err)
			//		break
			//	}
			//	time.Sleep(2 * time.Millisecond)
			//}
			data := make([]byte, 1024)
			for {
				if _, err = conn.Write([]byte("hello world!")); err != nil {
					fmt.Println(conn.LocalAddr(), err)
					break
				}
				//var buf []byte
				//for {
				_, err := conn.Read(data)
				if err != nil {
					fmt.Println(conn.LocalAddr(), err)
					break
					// panic(err)
				}
				//if n <= 0 {
				//	break
				//}
				//	buf = append(buf, data[:n]...)
				//}

				//fmt.Println(string(data[:n]))
				//time.Sleep(2 * time.Millisecond)
			}
			//time.Sleep(1 * time.Second)
			//_ = conn.Close()
		}()
		time.Sleep(5 * time.Millisecond)
	}
	for {
		time.Sleep(10 * time.Second)
	}
}
