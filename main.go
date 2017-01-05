package main

import (
	"net"
	"fmt"
)

func handleConnection(conn net.Conn){
	fmt.Println("New Connection")
	conn.Close();
}
func main() {
	ln,err:=net.Listen("tcp",":8080")
	if err!=nil{
		fmt.Println(err.Error())
	}
	for {
		conn,err:=ln.Accept()
		if err!=nil{
			fmt.Println(err.Error())
		}
		go handleConnection(conn)
	}
}
