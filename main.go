package main

import (
	"net"
	"fmt"
	"bufio"
	"net/http"
	"log"
	"io"
)

func handleConnection(conn net.Conn){
	defer conn.Close();
	reader:=bufio.NewReader(conn)
	//this goroutine will keep alive util closed with error or eof
	for{
		req,err:=http.ReadRequest(reader)
		if err!=nil{
			if err!=io.EOF{
				log.Printf("Failed to read request:%s",err.Error())
			}
			return
		}
		if be,err:=net.Dial("tcp","localhost:8081");err==nil{
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
