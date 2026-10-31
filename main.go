package main

import (
	"custom_tcp_stack_go_game/server"
)

func main() {
	go server.Server()
	select{}
}