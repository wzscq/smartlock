package lockhub

import (
	"smartlockservice/crv"
	"smartlockservice/common"
	"log"
)


var	hubFields=[]map[string]interface{}{
							{"field":"id"},
							{"field":"ip"},
						}


type HubItem struct {
	ID string
	IP string
	HubClient *HubClient
}

type HubList struct {
	HubMap map[string]HubItem
	Port string
	CommandResultHandler *CommandResultHandler
}

func (hl *HubList)load(crvClient * crv.CRVClient,token string){
	//获取锁列表
	commonRep:=crv.CommonReq{
		ModelID:"sl_hub",
		Fields:&hubFields,
	}

	rsp,commonErr:=crvClient.Query(&commonRep,token)
	if commonErr!=common.ResultSuccess {
		return
	}

	if rsp.Error == true {
		log.Println("Query hub list error:",rsp.ErrorCode,rsp.Message)
		return
	}

	resLst,ok:=rsp.Result["list"].([]interface{})
	if !ok {
		log.Println("Query hub list error: no list in rsp.")
		return
	}

	for _,res:=range resLst {
		resMap,ok:=res.(map[string]interface{})
		if !ok {
			log.Println("Query hub list error: no map in list.")
			return
		}
		hub:=HubItem{
			ID:resMap["id"].(string),
			IP:resMap["ip"].(string),
		}
		hubClient:=&HubClient{
			HubItem:hub,
			Port:hl.Port,
			Connected:false,
			CommandResultHandler:hl.CommandResultHandler,
		}
		hub.HubClient=hubClient
		hl.HubMap[hub.ID]=hub
	}
}

func (hl *HubList)getHubClient(id string)(*HubClient){
	hub,ok:=hl.HubMap[id]
	if !ok {
		return nil
	}

	return hub.HubClient
}