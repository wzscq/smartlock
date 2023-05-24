package lockhub

import (
	"smartlockservice/crv"
	"time"
	"log"
)

type LockController struct {
	LockList *LockList
	HubList *HubList
	CRVClient *crv.CRVClient
	HubPort string
	Timeout string
	Interval string
	BatchInterval string
}

func (lc *LockController)Init(){
	lc.LockList=new(LockList)
	lc.LockList.load(lc.CRVClient,"")

	lc.HubList=&HubList{
		HubMap:make(map[string]HubItem),
		Port:lc.HubPort,
		CommandResultHandler:&CommandResultHandler{
			CRVClient:lc.CRVClient,
		},
	}
	lc.HubList.load(lc.CRVClient,"")

	go lc.StartMonitor()
}

func (lc *LockController)StartMonitor(){			
	durationBatchInterval,_:=time.ParseDuration(lc.BatchInterval)

	for{
		//每轮开始前首先检查列表是否更新
		//lc.LockList.syncLockList()
		//逐个对锁的状态进行查询
		for _,lockItem:=range lc.LockList.Locks {
			cmd:=Command{
				CmdType:CMD_TYPE_STATUS,
				LockNo:lockItem.ID,
				Param:"000",
			}
			//send by master hub
			hubClient:=lc.HubList.getHubClient(lockItem.MasterHub)
			if hubClient!=nil {
				ok:=hubClient.SendCommand(cmd)
				if ok {
					continue
				}
			} else {
				log.Println("MasterHub not found:",lockItem.MasterHub)
			}
			//send by slaver hub
			hubClient=lc.HubList.getHubClient(lockItem.SlaverHub)
			if hubClient!=nil {
				hubClient.SendCommand(cmd)
			} else {
				log.Println("SlaverHub not found:",lockItem.SlaverHub)
			}
		}
		//等待对应间隔后再开始
		time.Sleep(durationBatchInterval)
	}
}
