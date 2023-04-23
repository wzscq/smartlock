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

    //先接收数据，把缓存中的数据清空
    // 接收数据
    buffer := make([]byte, 1024)
    for {
        length, err := conn.Read(buffer)
        if err != nil {
            conn.Close()
            connect_mutex.Unlock()
            return "",err
        }
        if length == 0 {
            break
        } 
        response := string(buffer[:length])
        log.Println("receive before send commnd", response)
    }

	log.Println("send command：", cmd)
    // 发送数据
    conn.Write([]byte(cmd))

    // 接收数据
    //buffer := make([]byte, 1024)
    receiveLength:=0
    for {
        length, err := conn.Read(buffer[receiveLength:])
        if err != nil {
            conn.Close()
            connect_mutex.Unlock()
            return "",err
        }
        if length == 0 {
            break
        }
        receiveLength+=length
    }
    response := string(buffer[:receiveLength])
    log.Println("receive buffer", response)
    resList:=strings.SplitAfter(response,"END")
    log.Println("receive resList", resList,len(resList))
    if len(resList)>=2 {
        response=resList[len(resList)-2]
    } else {
        response=resList[0]
    }
    log.Println("receive response：", response)
	conn.Close()
	connect_mutex.Unlock()
	return response,nil
}
