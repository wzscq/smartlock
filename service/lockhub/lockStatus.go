package lockhub

import (
	"sync"
	"smartlockservice/mqtt"
	"smartlockservice/common"
	"smartlockservice/crv"
	"smartlockservice/lock"
	"encoding/json"
	"log"
	"time"
	"net"
)

var lockFields=[]map[string]interface{}{
	{"field": "id"},
	{
		"field": "master_hub",
		"fieldType":"many2one",
		"relatedModelID":"sl_lock_hub",
		"fields": []map[string]interface{}{
			{"field": "id"},
			{"field": "ip"},
		},
	},
	{
		"field": "slaver_hub",
		"fieldType":"many2one",
		"relatedModelID":"sl_lock_hub",
		"fields": []map[string]interface{}{
			{"field": "id"},
			{"field": "ip"},
		},
	},
}

type lockStatus struct {
	LockID string
	LockStatus string
	MasterHub string
	SlaveHub string
}
//这里做两个锁的列表，一个用于实际监控锁的状态，一个用于更新需要监控的锁的列表
type lockStatusList struct {
	MonitorLockList []lockStatus
	UpdatedLockList []lockStatus
	MonitorListVersion int
	UpdatedLockVersion int
	LockListMutex sync.Mutex
}

type LockStatusMonitor struct {
	LockStatusList lockStatusList
	CRVClient *crv.CRVClient
	MQTTClient *mqtt.MQTTClient
	HubPort string
	Timeout string
	Interval string
	BatchInterval string
}

//检查锁更新列表是否更新过，如果更新过则将更新列表数据同步到监控列表
func (lsm *LockStatusMonitor)syncLockList(){
	if lsm.LockStatusList.MonitorListVersion != lsm.LockStatusList.UpdatedLockVersion {
		lsm.LockStatusList.LockListMutex.Lock()
		//将锁的历史状态先更新到新的锁列表中
		for index,updatedItem:=range lsm.LockStatusList.UpdatedLockList {
			for _,monitorItem:=range lsm.LockStatusList.MonitorLockList {
				if monitorItem.LockID == updatedItem.LockID {
					lsm.LockStatusList.UpdatedLockList[index].LockStatus=monitorItem.LockStatus
				}
			}
		}
		lsm.LockStatusList.MonitorLockList=lsm.LockStatusList.UpdatedLockList
		lsm.LockStatusList.MonitorListVersion=lsm.LockStatusList.UpdatedLockVersion
		lsm.LockStatusList.LockListMutex.Unlock()
	}
}

func (lsm *LockStatusMonitor)getLockStatus(lockStatus *lockStatus,timeoutDuration time.Duration)(string){
	log.Println("getLockStatus lockID:",lockStatus.LockID,lockStatus.MasterHub,lsm.HubPort)
	cmd:=Command{
		CmdType:CMD_TYPE_STATUS,
		LockNo:lockStatus.LockID,
		Param:"000",
	}
	//先链接主网关，如果主网关失败则链接备份网关
	if len(lockStatus.MasterHub)>0 {
		server:=lockStatus.MasterHub+":"+lsm.HubPort
		var err error
		cmd.Return,err=SendCommand(server,cmd.GetCommandStr(),timeoutDuration)
		if err == nil {
			log.Println(cmd.Return)
			newStatus:=cmd.GetRetVal()
			if len(newStatus)>0 {
				return newStatus
			}
			return lockStatus.LockStatus
		} else {
			log.Println("SendCommand error:",err)
		}
	}
	
	if len(lockStatus.SlaveHub)>0 {
		server:=lockStatus.SlaveHub+":"+lsm.HubPort
		var err error
		cmd.Return,err=SendCommand(server,cmd.GetCommandStr(),timeoutDuration)
		if err == nil {
			log.Println(cmd.Return)
			newStatus:=cmd.GetRetVal()
			if len(newStatus)>0 {
				return newStatus
			}
			return lockStatus.LockStatus
		} else {
			log.Println("SendCommand error:",err)
		}
	}

	return lockStatus.LockStatus
}

func (lsm *LockStatusMonitor)SendUpdateItemToMqtt(monitorItem lockStatus){
	opParam:=lock.OperParam{
		OperType:lock.OPER_ZTSS,
		Data:[]interface{}{
			map[string]interface{}{
				"lock_number":monitorItem.LockID,
				"lock_status":monitorItem.LockStatus,
			},
		},
	}

	bytes, err := json.Marshal(opParam)
	if err!=nil {
		log.Println("SendUpdateItemToMqtt error:",err)
		return
	}
  	// Convert bytes to string.
  	jsonStr := string(bytes)
	lsm.MQTTClient.Publish(lsm.MQTTClient.SendTopic,jsonStr)
}

func (lsm *LockStatusMonitor)StartMonitor() {
	log.Println("StartMonitor Lock Count：",len(lsm.LockStatusList.MonitorLockList))
	durationInterval, _ := time.ParseDuration(lsm.Interval)			
	durationBatchInterval,_:=time.ParseDuration(lsm.BatchInterval)
	timeoutDuration,_:=time.ParseDuration(lsm.Timeout)
	for{
		//每轮开始前首先检查列表是否更新
		lsm.syncLockList()
		//逐个对锁的状态进行查询
		for index,monitorItem:=range lsm.LockStatusList.MonitorLockList {
			status:=lsm.getLockStatus(&monitorItem,timeoutDuration)
			if status != monitorItem.LockStatus {
				log.Printf("lockID:%s,oldStatus:%s,newStatus:%s",monitorItem.LockID,monitorItem.LockStatus,status)
				lsm.LockStatusList.MonitorLockList[index].LockStatus=status
				lsm.SendUpdateItemToMqtt(lsm.LockStatusList.MonitorLockList[index])
			}
			time.Sleep(durationInterval)
		}
		//等待对应间隔后再开始
		time.Sleep(durationBatchInterval)
	}
}

func (lsm *LockStatusMonitor)queryLockList(token string,filter *map[string]interface{})([]interface{},int){
	//获取锁列表
	commonRep:=crv.CommonReq{
		ModelID:"sl_lock",
		Fields:&lockFields,
		Filter:filter,
		Sorter:&[]crv.Sorter{
			crv.Sorter{
				Field:"master_hub",
				Order:"asc",
			},
		},
	}

	rsp,commonErr:=lsm.CRVClient.Query(&commonRep,token)
	if commonErr!=common.ResultSuccess {
		return nil,commonErr
	}

	if rsp.Error == true {
		return nil,rsp.ErrorCode
	}

	resLst,ok:=rsp.Result["list"].([]interface{})
	if !ok {
		return nil,common.ResultNoParams
	}

	return resLst,common.ResultSuccess
}

func (lsm *LockStatusMonitor)dbItemToStatusItem(dbLockItem map[string]interface{})(lockStatus){
	//将数据转换为lockStatus
	lockStatus:=lockStatus{}
	lockStatus.LockID=dbLockItem["id"].(string)
	masterHub,ok:=dbLockItem["master_hub"].(map[string]interface{})
	if ok {
		masterHubList,ok:=masterHub["list"].([]interface{})
		if ok && len(masterHubList)>0 {
			masterHubRow,ok:=masterHubList[0].(map[string]interface{})
			if ok {
				lockStatus.MasterHub=masterHubRow["ip"].(string)
			}
		}
	}

	slaveHub,ok:=dbLockItem["slaver_hub"].(map[string]interface{})
	if ok {
		slaveHubList,ok:=slaveHub["list"].([]interface{})
		if ok && len(slaveHubList)>0 {
			slaveHubRow,ok:=slaveHubList[0].(map[string]interface{})
			if ok {
				lockStatus.SlaveHub=slaveHubRow["ip"].(string)
			}
		}
	}

	return lockStatus
}

func (lsm *LockStatusMonitor)dbListToStatusList(dbLockList []interface{})([]lockStatus){
	//将数据转换为lockStatus数组
	var data []lockStatus
	for _,item:=range dbLockList {
		lockItem:=item.(map[string]interface{})
		lockStatus:=lsm.dbItemToStatusItem(lockItem)
		data=append(data,lockStatus)
	}
	return data
}

func (lsm *LockStatusMonitor)UpdateLockList(token string)(int){
	//获取锁列表
	dbLockList,err:=lsm.queryLockList(token,nil)
	if err != common.ResultSuccess {
		return err
	}

	lockStatusList:=lsm.dbListToStatusList(dbLockList)

	//获取更新锁
	lsm.LockStatusList.LockListMutex.Lock()
	lsm.LockStatusList.UpdatedLockVersion+=1
	lsm.LockStatusList.UpdatedLockList=lockStatusList
	lsm.LockStatusList.LockListMutex.Unlock()
	return common.ResultSuccess
}

func (lsm *LockStatusMonitor)OpenBatch(token,closeDelay string,lockIDs []string)(int){
	//获取锁列表
	filter:=&map[string]interface{}{
		"id":map[string]interface{}{
			"Op.in":lockIDs,
		},
	}

	dbLockList,err:=lsm.queryLockList(token,filter)
	if err != common.ResultSuccess {
		return err
	}
	lockStatusList:=lsm.dbListToStatusList(dbLockList)
	if len(lockStatusList)<1 {
		return common.ResultQueryLockError
	}

	timeoutDuration,_:=time.ParseDuration(lsm.Timeout)

	log.Println("OpenBatch Lock Count：",len(lockStatusList))
	log.Println("OpenBatch Lock List：",lockStatusList)

	//获取锁
	GetLock()
	defer ReleaseLock()

	lastLockHub:=""
	var conn net.Conn
	for _,lockStatus:=range lockStatusList {
		//如果锁的主从hub不一致，则需要先关闭上一个hub的连接
		if lockStatus.MasterHub != lastLockHub {
			if conn != nil { 
				CloseConnect(conn)
				conn=nil
			}

			lastLockHub=lockStatus.MasterHub
			server:=lockStatus.MasterHub+":"+lsm.HubPort
			var err error
			conn,err=GetConnect(server,timeoutDuration)
			if err != nil {
				//主HUB链接失败，则尝试链接从HUB
				server:=lockStatus.SlaveHub+":"+lsm.HubPort
				conn,err=GetConnect(server,timeoutDuration)
				if err != nil {
					return common.ResultOpenLockError				
				}
			}
		}

		cmdCloseDelay:=Command{
			CmdType:CMD_TYPE_DELAY,
			LockNo:lockStatus.LockID,
			Param:closeDelay,
		}

		cmdOpen:=Command{
			CmdType:CMD_TYPE_OPEN,
			LockNo:lockStatus.LockID,
			Param:"000",
		}

		var err error
		//set close delay
		cmdCloseDelay.Return,err=SendCommandWithConnection(conn,timeoutDuration,cmdCloseDelay.GetCommandStr())
		if err!=nil {
			log.Println("SendCommand error:",err)
		}

		//open
		cmdOpen.Return,err=SendCommandWithConnection(conn,timeoutDuration,cmdOpen.GetCommandStr())
		if err!=nil {
			log.Println("SendCommand error:",err)
			return common.ResultOpenLockError
		}
	}

	//关闭链接
	if conn != nil {
		CloseConnect(conn)
	}

	return common.ResultSuccess
}

func (lsm *LockStatusMonitor)Open(token,lockID,closeDelay string)(int){
	//获取锁列表
	filter:=&map[string]interface{}{
		"id":lockID,
	}
	dbLockList,err:=lsm.queryLockList(token,filter)
	if err != common.ResultSuccess {
		return err
	}
	lockStatusList:=lsm.dbListToStatusList(dbLockList)
	if len(lockStatusList)<1 {
		return common.ResultQueryLockError
	}

	lockStatus:=lockStatusList[0]
	cmdCloseDelay:=Command{
		CmdType:CMD_TYPE_DELAY,
		LockNo:lockStatus.LockID,
		Param:closeDelay,
	}

	cmdOpen:=Command{
		CmdType:CMD_TYPE_OPEN,
		LockNo:lockStatus.LockID,
		Param:"000",
	}
	timeoutDuration,_:=time.ParseDuration(lsm.Timeout)
	//先链接主网关，如果主网关失败则链接备份网关
	log.Println(lockStatus.MasterHub)
	if len(lockStatus.MasterHub)>0 {
		server:=lockStatus.MasterHub+":"+lsm.HubPort
		var err error
		
		//set close delay
		cmdCloseDelay.Return,err=SendCommand(server,cmdCloseDelay.GetCommandStr(),timeoutDuration)
		if err!=nil {
			log.Println("SendCommand error:",err)
		}
		log.Println(cmdCloseDelay.Return)

		//open
		cmdOpen.Return,err=SendCommand(server,cmdOpen.GetCommandStr(),timeoutDuration)
		if err == nil {
			log.Println(cmdOpen.Return)
			retVal:=cmdOpen.GetRetVal()
			log.Println(retVal)
			return common.ResultSuccess
		} else {
			log.Println("SendCommand error:",err)
		}
	}
	
	log.Println(lockStatus.SlaveHub)
	if len(lockStatus.SlaveHub)>0 {
		server:=lockStatus.SlaveHub+":"+lsm.HubPort
		var err error
		//set close delay
		cmdCloseDelay.Return,err=SendCommand(server,cmdCloseDelay.GetCommandStr(),timeoutDuration)
		if err!=nil {
			log.Println("SendCommand error:",err)
		}
		log.Println(cmdCloseDelay.Return)
		
		cmdOpen.Return,err=SendCommand(server,cmdOpen.GetCommandStr(),timeoutDuration)
		if err == nil {
			log.Println(cmdOpen.Return)
			retVal:=cmdOpen.GetRetVal()
			log.Println(retVal)
			return common.ResultSuccess
		} else {
			log.Println("SendCommand error:",err)
		}
	}		
	return common.ResultOpenLockError
}