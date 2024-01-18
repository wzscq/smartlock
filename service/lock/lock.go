package lock

import (
	"smartlockservice/mqtt"
	"smartlockservice/common"
	"smartlockservice/crv"
	"log"	
	"encoding/json"
	"time"
)

const (
	OPER_OPEN="KS"  //开锁
	OPER_INIT="CSH"  //初始化锁具
	OPER_DELAY="BSYS"  //闭锁延时
	OPER_STATUS="MSZT"  //门锁状态
	OPER_GETID ="SHSC"  //锁号上传
	OPER_ZTSS = "ZTSS"  //状态上送
	OPER_SHSS = "SHSS"  //锁号上送
)

const (
	KC_OPER_AUTHREC="APPAuthRec"
)

type KCOperParm struct {
	OperType string `json:"commandType"`
	ApplicationID string `json:"applicationID"`
	KeyControllerID string `json:"keyControllerID"`
	KeyID string `json:"keyID"`
	Status string `json:"status"`
	Message string `json:"message"`
}

type OperParam struct {
	LockID string `json:"lock_number"`
	OperType string `json:"command_type"`
	TimeLapse *string `json:"time_lapse"`
	Time *string `json:"time"`
	Data []interface{} `json:"data"`
}

type LockOperator struct {
	CRVClient *crv.CRVClient
	MQTTClient *mqtt.MQTTClient
	AcceptTopic string
	KeyControlSendTopic string
	LockConf *common.LockConf
}

var appAuhorRecFields=[]map[string]interface{}{
	{"field": "id"},
	{"field": "approver"},
}

var applicatonFields=[]map[string]interface{}{
	{"field": "id"},
	{
		"field":"locks",
		"fieldType":"many2many",
		"relatedModelID":"sl_lock",
		"fields": []map[string]interface{}{
			{"field": "id"},
			{"field": "name"},
		},
	},
	{
		"field":"operators",
		"fieldType":"many2many",
		"relatedModelID":"sl_person",
		"fields": []map[string]interface{}{
			{"field": "id"},
			{"field": "name"},
		},
	},
	{
		"field":"approver",
		"fieldType":"many2one",
		"relatedModelID":"sl_person",
		"fields": []map[string]interface{}{
			{"field": "id"},
			{"field": "name"},
		},
	},
	{"field":"start_date"},
	{"field":"end_date"},
	//{"field":"start_time"},
	//{"field":"end_time"},
	{"field":"description"},
	{"field":"status"},
	{"field":"approval_comments"},
	{"field":"create_time"},
	{"field":"create_user"},
	{"field":"update_time"},
	{"field":"update_user"},
	{"field":"version"},
}

var lockFields=[]map[string]interface{}{
	{"field": "id"},
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

	if op.OperType==OPER_ZTSS {
		lockOperator.DealZTSS(&op)
	}
}

func (lockOperator *LockOperator)DealKeyControllerOperation(opMsg []byte){
	var op KCOperParm
	if err := json.Unmarshal(opMsg, &op); err != nil {
		log.Println(err)
		return
	}

	if op.OperType==KC_OPER_AUTHREC {
		lockOperator.DealAuthRec(&op)
	}
}

func (lockOperator *LockOperator)getAppAuthor(appID string)(string){
	commonRep:=crv.CommonReq{
		ModelID:"sl_application",
		Filter:&map[string]interface{}{
			"id":appID,
		},
		Fields:&appAuhorRecFields,
	}

	req,_:=lockOperator.CRVClient.Query(&commonRep,"")
	if req.Error == true {
		return ""
	}

	resLst,ok:=req.Result["list"].([]interface{})
	if !ok {
		return ""
	}

	if len(resLst)>0 {
		row,ok:=resLst[0].(map[string]interface{})
		if ok {
			author,ok:=row["approver"]
			if ok {
				return author.(string)
			}
		}
	}

	return ""
}

func (lockOperator *LockOperator)DealAuthRec(op *KCOperParm){
	author:=lockOperator.getAppAuthor(op.ApplicationID)

	recList:=[]map[string]interface{}{
		map[string]interface{}{
			"key_controller_id":op.KeyControllerID,
			"key_id":op.KeyID,
			"application_id":op.ApplicationID,
			"status":op.Status,
			"message":op.Message,
			"author":author,
			"_save_type":"create",
		},
	}
	
	saveReq:=&crv.CommonReq{
		ModelID:"sl_key_authorization",
		List:&recList,
	}
	lockOperator.CRVClient.Save(saveReq,"")
}

func (lockOperator *LockOperator)DealZTSS(op *OperParam){
	if op.Data!=nil && len(op.Data)>0 {
		lockList:=make([]map[string]interface{},len(op.Data))
		for index,lockItem:=range op.Data {
			lockItemMap:=lockItem.(map[string]interface{})
			lockList[index]=map[string]interface{}{
				"lock_id":lockItemMap["lock_number"],
				"status":lockItemMap["lock_status"],
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

func (lockOperator *LockOperator)WriteKey(keyControllerID,appID,token string)(int){
	//查询数据
	commonRep:=crv.CommonReq{
		ModelID:"sl_application",
		Filter:&map[string]interface{}{
			"id":appID,
		},
		Fields:&applicatonFields,
	}

	req,commonErr:=lockOperator.CRVClient.Query(&commonRep,token)
	if commonErr!=common.ResultSuccess {
		return commonErr
	}

	//构造发送结构
	bytes, err := json.Marshal(req)
	if err!=nil {
		log.Println("WriteKey convert query result to json error:",err.Error())
		return common.ResultJonsMarshalError
	}
  // Convert bytes to string.
  jsonStr := string(bytes)

	//发送mq消息	
	commonErr=lockOperator.MQTTClient.Publish(lockOperator.KeyControlSendTopic+"/"+keyControllerID,jsonStr)	
	if commonErr!=common.ResultSuccess {
		return commonErr
	}

	//更新申请状态
	if lockOperator.LockConf.UpdateAppStatus==true {
		list,ok:=req.Result["list"].([]interface{})
		if ok && len(list)>0 {
			row,ok:=list[0].(map[string]interface{})
			if ok {
				version,ok:=row["version"]
				if ok {
					lockOperator.UpdateAppStatus(appID,token,version)
				}
			}
		}
	}

	return common.ResultSuccess
}

func (lockOperator *LockOperator)SyncLockList(token string)(int){
	//查询数据
	commonRep:=crv.CommonReq{
		ModelID:"sl_lock",
		Fields:&lockFields,
	}

	rsp,commonErr:=lockOperator.CRVClient.Query(&commonRep,token)
	if commonErr!=common.ResultSuccess {
		return commonErr
	}

	if rsp.Error == true {
		return rsp.ErrorCode
	}

	resLst,ok:=rsp.Result["list"].([]interface{})
	if !ok {
		return common.ResultNoParams
	}

	var data=[]interface{}{}
	for _,item:=range resLst {
		lockItem:=item.(map[string]interface{})
		data=append(data,lockItem["id"])
	}

	//{'time': '2023-01-12 17:07:20', 'command_type': 'SHSS'#锁号上送, 'data': ['000000ab', '000000ac']}
	timeStr:=time.Now().Format("2006-01-02 15:04:05")
	param:=&OperParam{
		Time:&timeStr,
		OperType:OPER_SHSS,
		Data:data,
	}

	paramStr,err:=lockOperator.GetOperParamStr(param)
	if err!=common.ResultSuccess {
		return err
	}

	return lockOperator.MQTTClient.Publish(lockOperator.AcceptTopic,paramStr)
}

func (lockOperator *LockOperator)UpdateAppStatus(applicationID,token string,version interface{})(int){
	recList:=[]map[string]interface{}{
		map[string]interface{}{
			"id":applicationID,
			"status":"2",
			"_save_type":"update",
			"version":version,
		},
	}
	
	saveReq:=&crv.CommonReq{
		ModelID:"sl_application",
		List:&recList,
	}
	_,errorCode:=lockOperator.CRVClient.Save(saveReq,token)
	return errorCode
}