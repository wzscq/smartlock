package lock

import (
	"smartlockservice/mqtt"
	"smartlockservice/common"
	"smartlockservice/crv"
	"log"	
	"encoding/json"
)

const (
	OPER_OPEN="KS"  //开锁
	OPER_INIT="CSH"  //初始化锁具
	OPER_DELAY="BSYS"  //闭锁延时
	OPER_STATUS="MSZT"  //门锁状态
	OPER_GETID ="SHSC"  //锁号上传
	OPER_ZZSS = "ZTSS"  //状态上送
)

type OperParam struct {
	LockID string `json:"lock_number"`
	OperType string `json:"command_type"`
	TimeLapse *string `json:"time_lapse"`
	Data []map[string]interface{} `json:"data"`
}

type LockOperator struct {
	CRVClient *crv.CRVClient
	MQTTClient *mqtt.MQTTClient
	AcceptTopic string
}

func (lockOperator *LockOperator)GetOperParamStr(param *OperParam)(string,int){
	bytes, err := json.Marshal(param)
	if err!=nil {
		log.Println("GetOperParamStr error:",err.Error())
		return "",common.ResultJonsMarshalError
	}
  // Convert bytes to string.
  jsonStr := string(bytes)
	return jsonStr,common.ResultSuccess
}

func (lockOperator *LockOperator)Open(lockID string)(int){
	param:=&OperParam{
		LockID:lockID,
		OperType:OPER_OPEN,
	}

	paramStr,err:=lockOperator.GetOperParamStr(param)
	if err!=common.ResultSuccess {
		return err
	}

	return lockOperator.MQTTClient.Publish(lockOperator.AcceptTopic,paramStr)
}

func (lockOperator *LockOperator)DealLockOperation(opMsg []byte){
	var op OperParam
	if err := json.Unmarshal(opMsg, &op); err != nil {
		log.Println(err)
		return
	}

	if op.OperType==OPER_ZZSS {
		lockOperator.DealZZSS(&op)
	}
}

func (lockOperator *LockOperator)DealZZSS(op *OperParam){
	if op.Data!=nil && len(op.Data)>0 {
		lockList:=make([]map[string]interface{},len(op.Data))
		for index,lockItem:=range op.Data {
			lockList[index]=map[string]interface{}{
				"lock_id":lockItem["lock_number"],
				"status":lockItem["lock_status"],
				"_save_type":"create",
			}
		}
		saveReq:=&crv.CommonReq{
			ModelID:"sl_lock_status_record",
			List:&lockList,
		}
		lockOperator.CRVClient.Save(saveReq,"")
	}
}