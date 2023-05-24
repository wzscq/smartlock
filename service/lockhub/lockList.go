package lockhub

import (
	"smartlockservice/crv"
	"smartlockservice/common"
	"log"
)

var	queryLockFields=[]map[string]interface{}{
							{"field": "id"},
							{"field": "master_hub"},
							{"field": "slaver_hub"},
						}

type LockItem struct {
	ID string
	MasterHub string
	SlaverHub string
}

type LockList struct {
	Locks []LockItem
}

func (ll *LockList)load(crvClient * crv.CRVClient,token string){
	//获取锁列表
	commonRep:=crv.CommonReq{
		ModelID:"sl_lock",
		Fields:&queryLockFields,
		Sorter:&[]crv.Sorter{
			crv.Sorter{
				Field:"master_hub",
				Order:"asc",
			},
		},
	}

	rsp,commonErr:=crvClient.Query(&commonRep,token)
	if commonErr!=common.ResultSuccess {
		return
	}

	if rsp.Error == true {
		log.Println("Query lock list error:",rsp.ErrorCode,rsp.Message)
		return 
	}

	resLst,ok:=rsp.Result["list"].([]interface{})
	if !ok {
		log.Println("Query lock list error: no list in rsp.")
		return
	}

	for _,res:=range resLst {
		resMap,ok:=res.(map[string]interface{})
		if !ok {
			log.Println("Query lock list error: no map in list.")
			return
		}

		lockItem:=LockItem{
			ID:resMap["id"].(string),
			MasterHub:resMap["master_hub"].(string),
			SlaverHub:resMap["slaver_hub"].(string),
		}

		ll.Locks=append(ll.Locks,lockItem)
	}
}

func (ll *LockList)findLock(lockNo string) *LockItem {
	for _,lockItem:=range ll.Locks {
		if lockItem.ID == lockNo {
			return &lockItem
		}
	}

	return nil
}