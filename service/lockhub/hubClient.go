package lockhub

import (
	"log"
  "net"
	"sync"
	"time"
  "strings"
	"errors"

)

type SendCmdListItem struct {
	CMD Command
	Next *SendCmdListItem
}

type HubClient struct {
	HubItem HubItem
	Port string
	Timeout time.Duration
	Connected bool
	Conn net.Conn
	SendCmdListItem *SendCmdListItem
	CMDMutex sync.Mutex
	CommandResultHandler *CommandResultHandler
}

func (hc *HubClient)StartSendRecive(){
	go hc.SendReceive()
}

func (hc *HubClient)SendCommands(cmds []Command,first bool)(bool){
	hc.CMDMutex.Lock()
  defer hc.CMDMutex.Unlock()
	//如果没有连接，则先连接
	if hc.Connected==false {
		return false
	}

	var SendItems *SendCmdListItem
	var lastItem *SendCmdListItem
	//逐个添加到SendItems
	for _,cmd:=range cmds {
		if lastItem==nil {
			SendItems=&SendCmdListItem{
				CMD:cmd,
			}
			lastItem=SendItems
		} else {
			lastItem.Next=&SendCmdListItem{
				CMD:cmd,
			}
			lastItem=lastItem.Next
		}
	}

	//将发送命令放入队列
	if hc.SendCmdListItem==nil {
		hc.SendCmdListItem=SendItems
	} else {
		if first==true {
			lastItem.Next=hc.SendCmdListItem
			hc.SendCmdListItem=SendItems
		} else {
			item:=hc.SendCmdListItem
			for item.Next!=nil {
				item=item.Next
			}
			item.Next=SendItems
		}
	}
	
	return true
}

func (hc *HubClient)SendCommand(cmd Command,first bool)(bool){
	hc.CMDMutex.Lock()
  defer hc.CMDMutex.Unlock()
	//如果没有连接，则先连接
	if hc.Connected==false {
		return false
	}
	//讲发送命令放入队列
	if hc.SendCmdListItem==nil {
		hc.SendCmdListItem=&SendCmdListItem{
			CMD:cmd,
		}
	} else {
		if first==true {
			item:=&SendCmdListItem{
				CMD:cmd,
			}
			item.Next=hc.SendCmdListItem
			hc.SendCmdListItem=item
		} else {
			item:=hc.SendCmdListItem
			for item.Next!=nil {
				item=item.Next
			}
			item.Next=&SendCmdListItem{
				CMD:cmd,
			}
		}
	}

	return true
}

func (hc *HubClient)Connect()(bool){
	if hc.Connected==true {
		return true
	}

	var err error
	server:=hc.HubItem.IP+":"+hc.Port
	hc.Conn, err= net.DialTimeout("tcp", server,hc.Timeout)
	if err != nil {
		log.Println("connect to lockhub ",server, " error:", err)
		return false
	}
	log.Println("connect to lockhub ",server," success.")
	hc.Connected=true
	return true
}

func (hc *HubClient)Disconnect(){
	if hc.Connected==false {
		return
	}
	hc.Connected=false
	log.Println("lockhub connection closed. ip:",hc.HubItem.IP)
	hc.Conn.Close()
}

func (hc *HubClient)Send()(bool){
	hc.CMDMutex.Lock()
	defer hc.CMDMutex.Unlock()
	
	err:=hc.Conn.SetDeadline(time.Now().Add(hc.Timeout))
	if err != nil {
		log.Println("send command set timeout error:", err) 
		return false
	}

	if hc.SendCmdListItem!=nil {
		cmd:=hc.SendCmdListItem.CMD
		hc.SendCmdListItem=hc.SendCmdListItem.Next
		cmdStr:=cmd.GetCommandStr()
		log.Println("send command：", cmdStr, " to lockhub ",hc.HubItem.IP)
		hc.Conn.Write([]byte(cmdStr))
		if err != nil {
			log.Println("send command error:", err)
			return false
		}
	}

	return true
}

func (hc *HubClient)Receive()(bool){
	// 接收数据
	log.Println("receive response message ...", hc.HubItem.IP) 
	err:=hc.Conn.SetDeadline(time.Now().Add(hc.Timeout))
	if err != nil {
		log.Println("receive response set timeout error:", err) 
		return false
	}

	buffer := make([]byte, 1024)
	receiveLength:=0
	for {
			length, err := hc.Conn.Read(buffer[receiveLength:])
			log.Println("receive response message :", length, err)
			if err != nil {
				//如果不是超时错误，则报错返回
				if !hc.isTimeoutError(err) {
					return false
				}
				receiveLength+=length
				break
			}

			if length == 0 {
				break
			}

			receiveLength+=length
	}
	
	log.Println("receive response message length:",receiveLength)
	response := string(buffer[:receiveLength])
	log.Println("receive buffer", response)
	resList:=strings.SplitAfter(response,"END")
	
	//回调返回信息处理函数
	for _,res:=range resList {
		if res=="" {
			continue
		}
		if hc.CommandResultHandler!=nil {
			hc.CommandResultHandler.HandleCommandResult(res)
		}
	}
	log.Println("receive response：", response)
	return true
}

func (hc *HubClient)SendReceive(){
	for {
		if hc.Connected==false {
			//连接
			ok:=hc.Connect()
			if !ok {
				time.Sleep(10*time.Second)
				continue
			}
		}
		//接收数据
		ok:=hc.Receive()
		if !ok {
			hc.Disconnect()
			continue
		}
		//发送命令
		ok=hc.Send()
		if !ok {
			hc.Disconnect()
			continue
		}
		time.Sleep(2*time.Second)
	}
}

func (hc *HubClient)isTimeoutError(err error) bool {
	var netErr net.Error
	if errors.As(err, &netErr) {
			return netErr.Timeout()
	}
	return false
}

