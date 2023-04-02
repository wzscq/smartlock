package lockhub

import (
	"fmt"
)

const (
	CMD_START="AT+"
	CMD_END="END"
	CMD_JOIN="="
	CMD_RET="R"
)

const (
	CMD_TYPE_INIT="K"
	CMD_TYPE_DELAY="C"
	CMD_TYPE_OPEN="O"
	CMD_TYPE_STATUS="S"
	CMD_TYPE_QUERYNO="L"
)

const (
	CMD_SUCCESS="001"
	CMD_FAILURE="002"
)

const (
	STATUS_DOOROPEN_LOCKOPEN="000"
	STATUS_DOOROPEN_LOCKCLOSE="001"
	STATUS_DOORCLOSE_LOCKOPEN="010"
	STATUS_DOORCLOSE_LOCLCLOSE="011"	
)

type Command struct {
	CmdType string
	LockNo string
	Param string
	Return string 
}

func (cmd *Command)GetCommandStr()(string){
	return fmt.Sprintf("AT+%s=%s%sEND",cmd.CmdType,cmd.LockNo,cmd.Param)
}

func  (cmd *Command)getCommandRetrunPre()(string){
	return fmt.Sprintf("AT+%sR=%s",cmd.CmdType,cmd.LockNo)
}

func (cmd *Command)GetRetVal()(string){
	preLen:=len(cmd.getCommandRetrunPre())
	if len(cmd.Return)<=preLen {
		return ""
	}
	
	val:=cmd.Return[preLen:]
	if len(val)<3 {
		return val
	}

	return val[:len(val)-3]
}