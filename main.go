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


type Backend struct {
	net.Conn
	Reader *bufio.Reader
	Writer *bufio.Writer
}

var backendQueue chan *Backend
var requestBytes map[string]int64
var requestLock sync.Mutex

func init(){
	requestBytes=make(map[string]int64)
	backendQueue =make(chan *Backend,10)
}

func getBackend()(*Backend,error){
	select {
	case be:=<-backendQueue:
		return be,nil
	case <-time.After(100*time.Millisecond):
		be,err:=net.Dial("tcp","127.0.0.1:8081")
		if err!=nil{
			return nil,err
		}
		return &Backend{
			Conn:be,
			Reader:bufio.NewReader(be),
			Writer:bufio.NewWriter(be),
		},nil
	}
}

func queueBackend(be *Backend){
	select {
	case backendQueue<-be:
	case <-time.After(1*time.Second):
		be.Close()
	}
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

		be,err:=getBackend()
		if err!=nil{
			return
		}
		if err:=req.Write(be);err==nil{
				//read the resonse from the backend
				be.Writer.Flush()
				if resp,err:=http.ReadResponse(be.Reader,req);err==nil{
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
		go queueBackend(be)
		}
}
func main() {
	ln,err:=net.Listen("tcp",":8080")

	if err!=nil{
		fmt.Println(err.Error())
	}
	// this ticker will print to map[string]int64 obejct every 10 seconds.
	ticker := time.NewTicker(time.Second*20)
	go func() {
		for t := range ticker.C {
			fmt.Printf("\n[%s] Stats are %+v\n",t.String(),requestBytes)
		}
	}()




	for {
		if conn,err:=ln.Accept();err==nil{
			go handleConnection(conn)
		}
	}

}
