package main

import (
	"./service"
	"fmt"
	"net"
)

func main() {
	fmt.Println("Starting Server")
	listener, err := net.Listen("tcp", "0.0.0.0:5000")
	if err != nil {
		fmt.Println("Error listening", err.Error())
		return
	}
	service.Server(listener)
}
