package lockhub

import (
	"smartlockservice/crv"
)

type CommandResult struct {
	CmdType string
	LockNo string
	Result string
}

type CommandResultHandler struct {
	CRVClient *crv.CRVClient
}

func (crh *CommandResultHandler)HandleCommandResult(cmdResult string){
	
}

