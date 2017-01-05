package main

import (
	"net"
	"fmt"
	"bufio"
	"net/http"
	"log"
)

func handleConnection(conn net.Conn){
	defer conn.Close();
	reader:=bufio.NewReader(conn)
	if req,err:=http.ReadRequest(reader);err==nil{
		  //connect to backend
		  if be,err:=net.Dial("tcp","jsmean.com:80");err==nil{
			  beReader:=bufio.NewReader(be)
			  if err:=req.Write(be);err==nil{
					//read the resonse from the backend
				  if resp,err:=http.ReadResponse(beReader,req);err==nil{
					  resp.Close=true
					  if err:=resp.Write(conn);err==nil{
						  log.Printf("%s:%d",req.URL.Path,resp.StatusCode)
					  }
					}
				}
		  }
	}
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
		}else{
			go handleConnection(conn)
		}

	}
}
