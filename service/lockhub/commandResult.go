package lockhub

import (
	"log"
	"strings"
)

type CommandResult struct {
	CmdType string
	LockNo string
	Result string
}

type CommandResultHandler struct {
	LockList *LockList
}

func (crh *CommandResultHandler)HandleCommandResult(resultStr string){
	log.Println("HandleCommandResult:",resultStr)
	//目前仅处理状态反馈消息，其它消息均丢弃
	cmdResult:=crh.GetCommandResult(resultStr)
	if cmdResult==nil {
		log.Println("cmdResult is nil.")
		return
	}

	crh.LockList.UpdateLockStatus(cmdResult)
}

func (crh *CommandResultHandler)GetCommandResult(resultStr string)(* CommandResult){
	//命令中间是用=号分割的，先用=将字符串拆分成两个部分
	resultParts:=strings.Split(resultStr,"=")
	if len(resultParts)!=2 {
		return nil
	}

	//第一部分是命令类型，先根据命令类型判断是否是状态反馈消息
	cmdType:=resultParts[0]
	cmdReturnStatus:=GetCommandReturnType(CMD_TYPE_STATUS)
	if cmdType!=cmdReturnStatus {
		log.Println("cmdType is not status return type, cmdType:",cmdType)
		return nil
	}
	cmdType=CMD_TYPE_STATUS
	//第二部分是锁号和状态，最后三位是状态，其它位是锁号
	lockNoHexStr:=resultParts[1][0:len(resultParts[1])-6]
	lockNoDecStr:=GetLockNoDecStr(lockNoHexStr)
	status:=resultParts[1][len(resultParts[1])-6:len(resultParts[1])-3]

	log.Println("resultParts[1]",resultParts[1],"cmdType:",cmdType," lockNoHexStr:",lockNoHexStr," lockNoDecStr:",lockNoDecStr," status:",status)

	//将锁号和状态组装成CommandResult
	cmdResult:=&CommandResult{
		CmdType:cmdType,
		LockNo:lockNoDecStr,
		Result:status,
	}
	return cmdResult
}

