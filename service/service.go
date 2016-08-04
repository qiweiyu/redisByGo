package service

import (
	"../core"
	"net"
)

func Server(listener net.Listener) {
	for {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		go doServer(conn)
	}
}

func doServer(conn net.Conn) {
	core.Handle(conn)
}
