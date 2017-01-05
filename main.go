package main

import (
	"net"
	"fmt"
	"bufio"
	"net/http"
	"log"
	"io"
	"sync"
	"strconv"
	"time"
)

var requestBytes map[string]int64
var requestLock sync.Mutex

func init(){
	requestBytes=make(map[string]int64)
}

func updateStats(req *http.Request,resp *http.Response) int64{
	requestLock.Lock()
	defer requestLock.Unlock()
	var bytesLength int64
	//usually -1 mean unknown length, here we just add 0 to the original length.
	if resp.ContentLength==-1 {
		bytesLength=requestBytes[req.URL.Path]
	}else{
		bytesLength=requestBytes[req.URL.Path]+resp.ContentLength
	}
	requestBytes[req.URL.Path] = bytesLength
	return bytesLength
}

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
					bytesLength:=updateStats(req,resp)
					resp.Header.Set("X-Bytes-Length",strconv.FormatInt(bytesLength,10))
					if err:=resp.Write(conn);err==nil{
						log.Printf("%s:%d",req.URL.Path,resp.StatusCode)
					}
					if resp.Close {
						return
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
	// this ticker will print to map[string]int64 obejct every 10 seconds.
	ticker := time.NewTicker(time.Second*10)
	go func() {
		for t := range ticker.C {
			fmt.Printf("[%s] Stats are %+v\n",t.String(),requestBytes)
		}
	}()

	for {
		if conn,err:=ln.Accept();err==nil{
			go handleConnection(conn)
		}
	}

}
