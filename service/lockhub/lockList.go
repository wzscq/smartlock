package lockhub

import (
	"smartlockservice/crv"
	"smartlockservice/common"
	"log"
	"sync"
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
	Status string
}

type LockList struct {
	Locks []LockItem
	UpdateList []LockItem
	CRVClient *crv.CRVClient
	VersionMonitor int
	VersionUpdate int
	LocksMutex sync.Mutex
}

func (ll *LockList)load(){
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

	rsp,commonErr:=ll.CRVClient.Query(&commonRep,"")
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

func (ll *LockList)UpdateLockStatus(result *CommandResult){
	ll.LocksMutex.Lock()
	defer ll.LocksMutex.Unlock()
	//根据锁号查找锁
	lockItem:=ll.FindLock(result.LockNo)
	if lockItem==nil {
		log.Println("lockItem with No ",result.LockNo," not found.")
		return
	}

	log.Println("lockItem.ID:",lockItem.ID,"LockNo ",result.LockNo," lockItem.Status:",lockItem.Status,"result.Result:",result.Result)
	//查看锁的状态和返回结果的状态是否一致，如果一致则不需要更新
	if lockItem.Status==result.Result {
		return
	}

	log.Println("save sl_lock_status_record ...")
	lockItem.Status=result.Result
	//更新状态到数据库
	saveReq:=&crv.CommonReq{
		ModelID:"sl_lock_status_record",
		List:&[]map[string]interface{}{
			map[string]interface{}{
				"lock_id":lockItem.ID,
				"status":lockItem.Status,
				"_save_type":"create",
			},
		},
	}
	ll.CRVClient.Save(saveReq,"")
}

func (ll *LockList)FindLock(lockNo string) *LockItem {
	for index,lockItem:=range ll.Locks {
		if lockItem.ID == lockNo {
			return &ll.Locks[index]
		}
	}

	return nil
}

func (ll *LockList)SyncLockList(){
	//根据版本号判断是否需要更新
	if ll.VersionMonitor == ll.VersionUpdate {
		return
	}

	ll.LocksMutex.Lock()
	defer ll.LocksMutex.Unlock()

	//更新锁列表
	//将locks中的状态更新到updateList中
	for i:=0;i<len(ll.Locks);i++ {
		for j:=0;j<len(ll.UpdateList);j++ {
			if ll.Locks[i].ID == ll.UpdateList[j].ID {
				ll.UpdateList[j].Status=ll.Locks[i].Status
				break
			}
		}
	}

	//locks改为updateList
	ll.Locks=ll.UpdateList
	ll.VersionMonitor=ll.VersionUpdate
}

func (ll *LockList)UpdateLockList(){
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

	rsp,commonErr:=ll.CRVClient.Query(&commonRep,"")
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

	updateList:=[]LockItem{}
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

		updateList=append(updateList,lockItem)
	}

	ll.LocksMutex.Lock()
	defer ll.LocksMutex.Unlock()

	ll.UpdateList=updateList
	ll.VersionUpdate++
}