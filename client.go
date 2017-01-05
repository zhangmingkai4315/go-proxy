package main

import (
	"net/rpc"
	"log"
)
type Empty struct {}
type Stats struct {
	RequestBytes map[string]int64
}
func main(){
	client,err:=rpc.DialHTTP("tcp","127.0.0.1:8079")
	if err!=nil{
		log.Fatalf("Failed to dial : %s\n",err)
	}
	var reply Stats
	err=client.Call("RpcServer.GetStats",&Empty{},&reply)
	if err!=nil{
		log.Fatalf("Failed to getstats:%s\n",err)
		return
	}
	log.Printf("%+v",reply.RequestBytes)
}
