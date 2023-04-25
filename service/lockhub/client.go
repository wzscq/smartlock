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
    defer connect_mutex.Unlock()
	log.Println("connect to lockhub：", server)
	// 建立客户端连接
    conn, err := net.DialTimeout("tcp", server,timeoutDuration)
    if err != nil {
        //connect_mutex.Unlock()
        log.Println("connect to lockhub error:", err)
		return "",err
    }
    defer conn.Close()

    //先接收数据，把缓存中的数据清空
    log.Println("receive old message before send") 
    err = conn.SetDeadline(time.Now().Add(timeoutDuration))
    if err != nil {
        log.Println("receive old message set timeout error:", err) 
        //connect_mutex.Unlock()
        return "",err
    }
    // 接收数据
    buffer := make([]byte, 1024)
    for {
        length, err := conn.Read(buffer)
        if err != nil {
           log.Println("receive before send commnd", err) 
           break
        }
        if length == 0 {
            break
        } 
        response := string(buffer[:length])
        log.Println("receive before send commnd", response)
    }

	log.Println("send command：", cmd)
    // 发送数据
    conn.SetDeadline(time.Now().Add(timeoutDuration))
    if err != nil {
        //connect_mutex.Unlock()
        log.Println("send command set timeout error:", err) 
        return "",err
    }
    _, err=conn.Write([]byte(cmd))
    if err != nil {
        //connect_mutex.Unlock()
        log.Println("send command error:", err) 
        return "",err
    }

    // 接收数据
    //buffer := make([]byte, 1024)
    log.Println("receive response message ...") 
    conn.SetDeadline(time.Now().Add(timeoutDuration))
    if err != nil {
        //connect_mutex.Unlock()
        log.Println("receive response set timeout error:", err) 
        return "",err
    }
    receiveLength:=0
    for {
        length, err := conn.Read(buffer[receiveLength:])
        log.Println("receive response message :", length, err)
        if err != nil {
            if receiveLength<=0 && length<=0 {
                //conn.Close()
                //connect_mutex.Unlock()
                return "",err
            }
            receiveLength+=length
            break
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
	return response,nil
}
