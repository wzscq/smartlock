package lockhub

import (
	"testing"
	"time"
	"log"
)

func TestSendCommand(t *testing.T){
	server:="192.168.0.241:18100"
	duration:=time.Duration(5)*time.Second

	cmdCloseDelay:=Command{
		CmdType:CMD_TYPE_DELAY,
		LockNo:"",
		Param:"110",
	}

	cmdOpen:=Command{
		CmdType:CMD_TYPE_OPEN,
		LockNo:"",
		Param:"000",
	}

	rsp,err:=SendCommand(server,cmdCloseDelay.GetCommandStr(),duration)
	log.Println("rsp:",rsp,"err:",err)
	if err!=nil {
		t.Fatalf(err.Error())
	}

	rsp,err=SendCommand(server,cmdOpen.GetCommandStr(),duration)
	log.Println("rsp:",rsp,"err:",err)
	if err!=nil {
		t.Fatalf(err.Error())
	}
}