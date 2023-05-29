package lockhub

import (
	"smartlockservice/crv"
	"smartlockservice/common"
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
	lc.LockList=&LockList{
		CRVClient:lc.CRVClient,
	}
	lc.LockList.load()

	lc.HubList=&HubList{
		HubMap:make(map[string]HubItem),
		Port:lc.HubPort,
		CommandResultHandler:&CommandResultHandler{
			LockList:lc.LockList,
		},
	}
	lc.HubList.load(lc.CRVClient,"")
	lc.StartHubClient()

	go lc.StartMonitor()
}

func (lc *LockController)StartHubClient(){			
	//循环启动hubclient
	for _,hubItem:=range lc.HubList.HubMap {
		hubClient:=hubItem.HubClient
		hubClient.StartSendRecive()
	}
}

func (lc *LockController)StartMonitor(){			
	durationBatchInterval,_:=time.ParseDuration(lc.BatchInterval)

	for{
		//每轮开始前首先检查列表是否更新
		lc.LockList.SyncLockList()
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

func (lc *LockController)UpdateLockList(){		
	lc.LockList.UpdateLockList()
}

func (lc *LockController)OpenLocks(closeDelay string,lockIDs []string)(int){
	//循环处理每个锁
	for _,lockID:=range lockIDs {
		//根据锁号查找锁
		lockItem:=lc.LockList.FindLock(lockID)
		if lockItem==nil {
			log.Println("lockItem with No ",lockID," not found.")
			return common.ResultOpenLockError
		}

		//发送开锁命令
		cmdCloseDelay:=Command{
			CmdType:CMD_TYPE_DELAY,
			LockNo:lockItem.ID,
			Param:closeDelay,
		}
	
		cmdOpen:=Command{
			CmdType:CMD_TYPE_OPEN,
			LockNo:lockItem.ID,
			Param:"000",
		}

		//获取锁所在的主从集线器
		hubClient:=lc.HubList.getHubClient(lockItem.MasterHub)
		if hubClient!=nil {
			hubClient.SendCommand(cmdCloseDelay)
			ok:=hubClient.SendCommand(cmdOpen)
			if ok {
				continue
			}
		} else {
			log.Println("MasterHub not found:",lockItem.MasterHub)
			return common.ResultOpenLockError
		}
		//send by slaver hub
		hubClient=lc.HubList.getHubClient(lockItem.SlaverHub)
		if hubClient!=nil {
			hubClient.SendCommand(cmdCloseDelay)
			ok:=hubClient.SendCommand(cmdOpen)
			if !ok {
				return common.ResultOpenLockError
			}
		} else {
			log.Println("SlaverHub not found:",lockItem.SlaverHub)
			return common.ResultOpenLockError
		}
	}
	return common.ResultSuccess
}
