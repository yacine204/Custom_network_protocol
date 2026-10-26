package main

import (
	"custom_tcp_stack_go_game/client"
	"custom_tcp_stack_go_game/server"
	"fmt"
)

func main(){
	ready := make(chan error, 1)

	go server.Server(ready)
	if err:=<-ready; err!=nil{
		fmt.Printf("[ERR] server failed to start: %s\n", err)
		return
	}
	go client.Conn()

	select{}
}