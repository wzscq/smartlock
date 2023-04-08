package lockhub

import (
	"log"
    "net"
	"sync"
	"time"
    "strings"
)

var connect_mutex sync.Mutex

func SendCommand(server,cmd string,timeoutDuration time.Duration)(string,error){
	connect_mutex.Lock()
	log.Println("connect to lockhub：", server)
	// 建立客户端连接
    conn, err := net.DialTimeout("tcp", server,timeoutDuration)
    if err != nil {
        connect_mutex.Unlock()
		return "",err
    }

    err = conn.SetDeadline(time.Now().Add(timeoutDuration))
    if err != nil {
        connect_mutex.Unlock()
        return "",err
    }

	log.Println("send command：", cmd)
    // 发送数据
    conn.Write([]byte(cmd))

    // 接收数据
    buffer := make([]byte, 1024)
    length, err := conn.Read(buffer)
    if err != nil {
		conn.Close()
		connect_mutex.Unlock()
        return "",err
    }
    response := string(buffer[:length])
    log.Println("receive buffer", response)
    resList:=strings.SplitAfter(response,"END")
    log.Println("receive resList", resList,len(resList))
    response=resList[len(resList)-2]
    log.Println("receive response：", response)
	conn.Close()
	connect_mutex.Unlock()
	return response,nil
}
