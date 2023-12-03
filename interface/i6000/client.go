package i6000

import (
	"smartlockservice/common"
	"smartlockservice/crv"
	"log"
	"time"
	"bytes"
	"net/http"
	"encoding/json"
	"strconv"
	"regexp"
)

type FindWorkTicketRequestBody struct {
	PlanBeginTimeStart string `json:"planBeginTimeStart"`
	//PlanBeginTimeEnd string `json:"planBeginTimeEnd"`
	PageNum string `json:"pageNum"`
	PageSize string `json:"pageSize"`
	AllOrgIds []string `json:"allOrgIds"`
	FlowProcessStepId string `json:"flowProcessStepId"`
}

type RackItem struct {
	Room string
	Rack string
}

type WorkTicketItem struct {
	ID string `json:"id"`
	Code string `json:"code"`
	WorkPersionLiable string `json:"workPersionLiable"`
	TeamName string `json:"teamName"`
	TeamMemberCount string `json:"teamMemberCount"`
	WorkSceneName string `json:"workSceneName"`
	TeamMember string `json:"teamMember"`
	PlanBeginTime string `json:"planBeginTime"`
	PlanEndTime string `json:"planEndTime"`
	FlowProcessStepName string `json:"flowProcessStepName"`
	WorkTask string `json:"workTask"`
	WorkContent string `json:"workContent"`
	NowHandleName string `json:"nowHandleName"`
	TaskOrTicketName string `json:"taskOrTicketName"`
	SafetyMeasures string `json:"safetyMeasures"`     
}

type SystemInfoResponseData struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Category string `json:"category"`
	Net string `json:"net"`
	Type string `json:"type"`
	Address string `json:"address"`
	Person string `json:"person"`
}

type SystemInfoResponse struct {
	Success int `json:"success"`
	Code string `json:"code"`
	Msg string `json:"msg"`
	Data *[]SystemInfoResponseData `json:"data"`
}

type FindWorkTicketResponse struct {
	Success bool `json:"success"`
	Code string `json:"code"`
	Message string `json:"message"`
	Data *[]WorkTicketItem `json:"data"`
	Time string `json:"time"`
	Total int `json:"total"`
}

type SignData struct {
	SignData string `json:"signData"`
	AccessToken string `json:"accessToken"`
	Timestamp string `json:"timestamp"`
	Status int `json:"status"`
}

type I6000Client struct {
	CRVClient *crv.CRVClient
	I6000Conf *common.I6000Conf
}

func (client *I6000Client) Init() {
	log.Println("I6000Client Init")
	go client.StartQueryWorkTicket()
}

func (client *I6000Client) StartQueryWorkTicket() {
	durationInterval,_:=time.ParseDuration(client.I6000Conf.QueryInterval)
	log.Println("StartQueryWorkTicket with interval ",client.I6000Conf.QueryInterval)
	for{
		log.Println("StartQueryWorkTicket with interval ",client.I6000Conf.QueryInterval)
		client.syncWorkTicket("")
		client.updateWorkTickey("")
		time.Sleep(durationInterval)
	}
}

func GetSignData(url string)(*SignData){
	req,err:=http.NewRequest("GET",url,nil)
	if err != nil {
		log.Println("GetSignData error",err)
		return nil
	}	
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		log.Println("GetSignData Do request error",err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 { 
		log.Println("GetSignData StatusCode error",resp)
		return nil
	}

	decoder := json.NewDecoder(resp.Body)
	signData:=&SignData{}
	err = decoder.Decode(signData)
	if err != nil {
		log.Println("GetSignData result decode failed [Err:%s]", err.Error())
		return nil
	}
	return signData
}

func (client *I6000Client)GetExistWorkTickets(workTickets *[]WorkTicketItem)(*[]string){
	ids:=[]string{}
	for _,workTicket:=range *workTickets {
		ids=append(ids,workTicket.ID)
	}

	commonRep:=crv.CommonReq{
		ModelID:"sl_work_ticket",
		Fields:&[]map[string]interface{}{
			{"field":"id"},
		},
		Filter:&map[string]interface{}{
			"id":map[string]interface{}{
				"Op.in":ids,
			},
		},
		Pagination:&crv.Pagination{
			Current:1,
			PageSize:len(ids),
		},
	}

	rsp,commonErr:=client.CRVClient.Query(&commonRep,"")
	if commonErr!=common.ResultSuccess {
		return &ids
	}

	if rsp.Error == true {
		log.Println("Query work ticket list error:",rsp.ErrorCode,rsp.Message)
		return &ids
	}

	ids=[]string{}

	resLst,ok:=rsp.Result["list"].([]interface{})
	if !ok {
		log.Println("Query work ticket list error: no list in rsp.")
		return &ids
	}

	for _,res:=range resLst {
		resMap,ok:=res.(map[string]interface{})
		if !ok {
			log.Println("Query work ticket list error: no map in list.")
			return &ids
		}
		ids=append(ids,resMap["id"].(string))
	}
	return &ids
}

func (client *I6000Client)removeExistWorkTicket(workTickets *[]WorkTicketItem,exist *[]string)(*[]WorkTicketItem){
	newWorkTickets:=[]WorkTicketItem{}

	for _,workTicket:=range *workTickets {
		existFlag:=false
		for _,id:=range *exist {
			if workTicket.ID==id {
				existFlag=true
				break
			}
		}
		if !existFlag {
			newWorkTickets=append(newWorkTickets,workTicket)
		}
	}

	return &newWorkTickets
}

func (client *I6000Client)getWorkTicketSystems(id string)(*[]SystemInfoResponseData){
	req,err:=http.NewRequest("POST",client.I6000Conf.SelectInvolveSystemInfo+"/"+id,nil)
	if err != nil {
		log.Println("I6000Client getWorkTicketSystems NewRequest error",err)
		return nil
	}
	
	req.Header.Set("Content-Type","application/json")
	signData:=GetSignData(client.I6000Conf.GetSignDataUrl)
	if signData==nil {
		log.Println("I6000Client getWorkTicketSystems GetSignData error")
		return nil
	}
	req.Header.Set("AccessToken",signData.AccessToken)
	req.Header.Set("signData",signData.SignData)
	
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		log.Println("I6000Client getWorkTicketSystems Do request error",err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 { 
		log.Println("I6000Client getWorkTicketSystems StatusCode error",resp)
		return nil
	}

	decoder := json.NewDecoder(resp.Body)
	rsp:=SystemInfoResponse{}
	err = decoder.Decode(&rsp)
	if err != nil {
		log.Println("I6000Client getWorkTicketSystems result decode failed [Err:%s]", err.Error())
		return nil
	}
	return rsp.Data
}

func (client *I6000Client)getWorkTicketDevices(id string)(*[]SystemInfoResponseData){
	req,err:=http.NewRequest("POST",client.I6000Conf.SelectInvolveDeviceInfo+"/"+id,nil)
	if err != nil {
		log.Println("I6000Client getWorkTicketDevices NewRequest error",err)
		return nil
	}
	
	req.Header.Set("Content-Type","application/json")
	signData:=GetSignData(client.I6000Conf.GetSignDataUrl)
	if signData==nil {
		log.Println("I6000Client getWorkTicketDevices GetSignData error")
		return nil
	}
	req.Header.Set("AccessToken",signData.AccessToken)
	req.Header.Set("signData",signData.SignData)
	
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		log.Println("I6000Client getWorkTicketDevices Do request error",err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 { 
		log.Println("I6000Client getWorkTicketDevices StatusCode error",resp)
		return nil
	}

	decoder := json.NewDecoder(resp.Body)
	rsp:=SystemInfoResponse{}
	err = decoder.Decode(&rsp)
	if err != nil {
		log.Println("I6000Client getWorkTicketDevices result decode failed [Err:%s]", err.Error())
		return nil
	}
	return rsp.Data
}

func (client *I6000Client)saveNewWorkTicket(workTicket *WorkTicketItem,deviceInfo *[]SystemInfoResponseData){
	deviceInfoList:=[]map[string]interface{}{}
	for _,device:=range *deviceInfo {
		deviceInfoList=append(deviceInfoList,map[string]interface{}{
			"work_ticket_id":workTicket.ID,
			"system_id":device.ID,
			"name":device.Name,
			"category":device.Category,
			"net":device.Net,
			"type":device.Type,
			"address":device.Address,
			"person":device.Person,
			"_save_type":"create",
		})
	}

	//保存工单
	recList:=[]map[string]interface{}{
		map[string]interface{}{
			"id":workTicket.ID,
			"code":workTicket.Code,
			"work_persion_liable":workTicket.WorkPersionLiable,
			"team_name":workTicket.TeamName,
			"team_member_count":workTicket.TeamMemberCount,
			"work_scene_name":workTicket.WorkSceneName,
			"team_member":workTicket.TeamMember,
			"plan_begin_time":workTicket.PlanBeginTime,
			"plan_end_time":workTicket.PlanEndTime,
			"flow_process_step_name":workTicket.FlowProcessStepName,
			"now_handle_name":workTicket.NowHandleName,
			"task_or_ticket_name":workTicket.TaskOrTicketName,
			"safety_measures":workTicket.SafetyMeasures,
			"work_task":workTicket.WorkTask,
			"involve_devices":map[string]interface{}{
				"fieldType":"one2many",
        "modelID":"sl_involve_device",
        "relatedField":"work_ticket_id",
				"list":deviceInfoList,
			},
			"_save_type":"create",
		},
	}

	saveReq:=&crv.CommonReq{
		ModelID:"sl_work_ticket",
		List:&recList,
	}
	client.CRVClient.Save(saveReq,"")
}

func (client *I6000Client)getDeviceRoomRack(deviceName string)(*RackItem){
	re := regexp.MustCompile(`(.*)机房(.*)机柜`)
	replaceItems:=re.FindAllStringSubmatch(deviceName,-1)
	log.Println("I6000Client getDeviceRoomRack deviceName",replaceItems)
	if len(replaceItems)!=1 {
		log.Println("I6000Client getDeviceRoomRack deviceName error",deviceName)
		return nil
	}

	rackItem:=RackItem{
		Room:replaceItems[0][1],
		Rack:replaceItems[0][2],
	}

	return &rackItem
}

func (client *I6000Client)getDeviceRackLockIds(deviceInfo *[]SystemInfoResponseData)(*[]string){
	if deviceInfo==nil {
		log.Println("I6000Client getDeviceRackLockIds deviceInfo is nil")
		return nil
	}

	if len(*deviceInfo)==0 {
		log.Println("I6000Client getDeviceRackLockIds deviceInfo is empty")
		return nil
	}

	rackList:=[]RackItem{}
	for _,device:=range *deviceInfo {
		reckItem:=client.getDeviceRoomRack(device.Name)
		if reckItem!=nil {
			rackList=append(rackList,*reckItem)
		}
	}

	if len(rackList)==0 {
		log.Println("I6000Client getDeviceRackLockIds rackList is empty")
		return nil
	}

	return client.getLockIds(&rackList)
}

func (client *I6000Client)getLockIds(rackList *[]RackItem)(*[]string){
	conditions:=[]map[string]interface{}{}
	for _,rackItem:=range *rackList {
		conditions=append(conditions,map[string]interface{}{
			"room":rackItem.Room,
			"rack":rackItem.Rack,
		})
	}

	commonRep:=crv.CommonReq{
		ModelID:"sl_lock",
		Fields:&[]map[string]interface{}{
			{"field":"id"},
		},
		Filter:&map[string]interface{}{
			"Op.or":conditions,
		},
		Pagination:&crv.Pagination{
			Current:1,
			PageSize:len(conditions)*2,
		},
	}

	rsp,commonErr:=client.CRVClient.Query(&commonRep,"")
	if commonErr!=common.ResultSuccess {
		return nil
	}

	if rsp.Error == true {
		log.Println("Query rack list error:",rsp.ErrorCode,rsp.Message)
		return nil
	}

	ids:=[]string{}
	resLst,ok:=rsp.Result["list"].([]interface{})
	if !ok {
		log.Println("Query rack list error: no list in rsp.")
		return nil
	}

	for _,res:=range resLst {
		resMap,ok:=res.(map[string]interface{})
		if !ok {
			log.Println("Query rack list error: no map in list.")
			return nil
		}
		ids=append(ids,resMap["id"].(string))
	}
	return &ids
}

func (client *I6000Client)createOpenLockApp(workTicket *WorkTicketItem,lockList *[]string){
	log.Println("createOpenLockApp start ");
	crvLockList:=[]map[string]interface{}{}
	for _,lockID:=range *lockList {
		crvLockList=append(crvLockList,map[string]interface{}{
			"id":lockID,
			"_save_type":"create",
		})
	}

	//保存工单
	des:="工单："+workTicket.ID+"，工单编号："+workTicket.Code+"，工作负责人："+workTicket.WorkPersionLiable+"，工作班成员："+workTicket.TeamMember+"，工作内容："+workTicket.WorkContent
	recList:=[]map[string]interface{}{
		map[string]interface{}{
			"start_date":workTicket.PlanBeginTime,
			"end_date":workTicket.PlanEndTime,
      "description":des,
			"locks":map[string]interface{}{
				"fieldType":"many2many",
        "modelID":"sl_lock",
				"list":crvLockList,
			},
			"status":"2",
			"_save_type":"create",
		},
	}

	saveReq:=&crv.CommonReq{
		ModelID:"sl_application",
		List:&recList,
	}
	client.CRVClient.Save(saveReq,"")
	log.Println("createOpenLockApp end ");
}

func (client *I6000Client)saveNewWorkTickets(workTickets *[]WorkTicketItem){
	log.Println("I6000Client saveNewWorkTickets start")
	for _,workTicket:=range *workTickets {
		deviceInfo:=client.getWorkTicketDevices(workTicket.ID)
		lockList:=client.getDeviceRackLockIds(deviceInfo)
		if lockList!=nil {
			client.saveNewWorkTicket(&workTicket,deviceInfo)
			client.createOpenLockApp(&workTicket,lockList)
		}
	}
	log.Println("I6000Client saveNewWorkTickets end")
}

func (client *I6000Client)queryWorkTicket(id string)(map[string]interface{}){
	commonRep:=crv.CommonReq{
		ModelID:"sl_work_ticket",
		Fields:&[]map[string]interface{}{
			{"field":"id"},
			{"field":"version"},
			{
				"field":"involve_devices",
				"fieldType":"one2many",
				"relatedModelID":"sl_involve_device",
				"relatedField":"work_ticket_id",
				"fields":&[]map[string]interface{}{
					{"field":"id"},
					{"field":"version"},
					{"field":"work_ticket_id"},
				},
			},
		},
		Filter:&map[string]interface{}{
			"id":id,
		},
		Pagination:&crv.Pagination{
			Current:1,
			PageSize:1,
		},
	}

	rsp,commonErr:=client.CRVClient.Query(&commonRep,"")
	if commonErr!=common.ResultSuccess {
		return nil
	}

	if rsp.Error == true {
		log.Println("Query work ticket list error:",rsp.ErrorCode,rsp.Message)
		return nil
	}

	return rsp.Result
}

func (client *I6000Client)getDeleteDeviceList(involveDevices map[string]interface{})([]map[string]interface{}){
	involveDevicesList,ok:=involveDevices["list"].([]interface{})
	if !ok {
		log.Println("getDeleteDeviceList get involve device list error: no list in rsp.")
		return nil
	}
	deviceList:=[]map[string]interface{}{}
	for _,device:=range involveDevicesList {
		deviceMap,ok:=device.(map[string]interface{})
		if !ok {
			log.Println("getDeleteDeviceList get device list error: no map in list.")
			return nil
		}
		deviceMap["_save_type"]="delete"
		deviceList=append(deviceList,deviceMap)
	}
	return deviceList
}

func (client *I6000Client)deleteWorkTicketByRes(res map[string]interface{}){
	delTickets:=[]map[string]interface{}{}
	tickeList,ok:=res["list"].([]interface{})
	if !ok {
		log.Println("deleteWorkTicketByRes get work ticket list error: no list in rsp.")
		return
	}

	for _,ticket:=range tickeList {
		ticketMap,ok:=ticket.(map[string]interface{})
		if !ok {
			log.Println("deleteWorkTicketByRes get work ticket list error: no map in list.")
			return
		}
		ticketMap["_save_type"]="delete"
		involveDevices,ok:=ticketMap["involve_devices"].(map[string]interface{})
		if ok {
			deviceList:=client.getDeleteDeviceList(involveDevices)
			log.Println("deleteWorkTicketByRes delete ticket:",deviceList)
			if deviceList != nil {
				saveReq:=&crv.CommonReq{
					ModelID:"sl_involve_device",
					List:&deviceList,
				}
				client.CRVClient.Save(saveReq,"")
			}
		}
		delete(ticketMap,"involve_devices")
		log.Println("deleteWorkTicketByRes delete ticket:",ticketMap)
		delTickets=append(delTickets,ticketMap)
	}

	saveReq:=&crv.CommonReq{
		ModelID:"sl_work_ticket",
		List:&delTickets,
	}
	client.CRVClient.Save(saveReq,"")
}

func (client *I6000Client)queryApplication(ticketID string)(map[string]interface{}){
	commonRep:=crv.CommonReq{
		ModelID:"sl_application",
		Fields:&[]map[string]interface{}{
			{"field":"id"},
			{"field":"version"},
		},
		Filter:&map[string]interface{}{
			"description":map[string]interface{}{
				"Op.like":"工单："+ticketID+"%",
			},
			"status":"2",
		},
	}

	rsp,commonErr:=client.CRVClient.Query(&commonRep,"")
	if commonErr!=common.ResultSuccess {
		return nil
	}

	if rsp.Error == true {
		log.Println("Query work ticket list error:",rsp.ErrorCode,rsp.Message)
		return nil
	}

	return rsp.Result
}

func (client *I6000Client)updateApplicationStatus(res map[string]interface{}){
	applicationList,ok:=res["list"].([]interface{})
	if !ok {
		log.Println("updateApplicationStatus get application list error: no list in rsp.")
		return
	}
	for _,application:=range applicationList {
		applicationMap,ok:=application.(map[string]interface{})
		if !ok {
			log.Println("updateApplicationStatus get application list error: no map in list.")
			return
		}
		applicationMap["status"]="4"
		applicationMap["_save_type"]="update"
		saveReq:=&crv.CommonReq{
			ModelID:"sl_application",
			List:&[]map[string]interface{}{applicationMap},
		}
		client.CRVClient.Save(saveReq,"")
	}
}

func (client *I6000Client)deleteWorkTicket(workTicket *WorkTicketItem){
	//查询出工单信息
	res:=client.queryWorkTicket(workTicket.ID)
	if res==nil {
		return
	}
	//删除工单
	client.deleteWorkTicketByRes(res)

	//查询出对应的申请单
	res=client.queryApplication(workTicket.ID)
	if res == nil {
		return
	}
	//更新申请单状态
	client.updateApplicationStatus(res)
}

func (client *I6000Client)deleteWorkTickets(workTickets *[]WorkTicketItem){
	log.Println("I6000Client deleteWorkTickets start")
	for _,workTicket:=range *workTickets {
		client.deleteWorkTicket(&workTicket)
	}
	log.Println("I6000Client deleteWorkTickets end")
}

func (client *I6000Client)updateWorkTickey(token string) (int) {
	log.Println("I6000Client updateWorkTickey start")
	//获取工单列表
	workTickets:=client.getUpdatedWorkTicket()
	if len(*workTickets)==0 {
		log.Println("I6000Client updateWorkTickey end with no work ticket")
		return common.ResultSuccess
	}

	//删除原有工单，将对应开锁申请的状态设置为4
	client.deleteWorkTickets(workTickets)

	//保存新增工单
	client.saveNewWorkTickets(workTickets)

	log.Println("I6000Client updateWorkTickey end")
	return common.ResultSuccess
}

func (client *I6000Client) syncWorkTicket(token string) (int) {
	log.Println("I6000Client syncWorkTicket start")
	//获取工单列表
	workTickets:=client.getWorkTicket()
	if len(*workTickets)==0 {
		log.Println("I6000Client syncWorkTicket end with no work ticket")
		return common.ResultSuccess
	}
	//检查工单是否已经存在
	crvTickets:=client.GetExistWorkTickets(workTickets)
	if len(*crvTickets)==len(*workTickets) {
		log.Println("I6000Client syncWorkTicket end with all work ticket exist")
		return common.ResultSuccess
	}

	if len(*crvTickets)>0 && len(*crvTickets)<len(*workTickets) {
		workTickets=client.removeExistWorkTicket(workTickets,crvTickets)
	}

	//保存新增工单
	client.saveNewWorkTickets(workTickets)

	log.Println("I6000Client syncWorkTicket end")
	return common.ResultSuccess
}

func (client *I6000Client)getWorkTicketByPage(requestBody *FindWorkTicketRequestBody)(*FindWorkTicketResponse){
	postJson,_:=json.Marshal(requestBody)
	postBody:=bytes.NewBuffer(postJson)
	req,err:=http.NewRequest("POST",client.I6000Conf.FindWorkTicket,postBody)
	if err != nil {
		log.Println("I6000Client getWorkTicketByPage NewRequest error",err)
		return nil
	}
	
	req.Header.Set("Content-Type","application/json")
	signData:=GetSignData(client.I6000Conf.GetSignDataUrl)
	if signData==nil {
		log.Println("I6000Client getWorkTicketByPage GetSingData error")
		return nil
	}
	req.Header.Set("AccessToken",signData.AccessToken)
	req.Header.Set("signData",signData.SignData)
	
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		log.Println("I6000Client getWorkTicketByPage Do request error",err)
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 { 
		log.Println("I6000Client getWorkTicketByPage StatusCode error",resp)
		return nil
	}

	decoder := json.NewDecoder(resp.Body)
	rsp:=FindWorkTicketResponse{}
	err = decoder.Decode(&rsp)
	if err != nil {
		log.Println("I6000Client getWorkTicketByPage result decode failed [Err:%s]", err.Error())
		return nil
	}
	return &rsp
}

func (client *I6000Client)getUpdatedWorkTicket()(*[]WorkTicketItem){
	workTickets:=[]WorkTicketItem{}
	//查询开工日期在当前时间之后的工单
	now := time.Now()
	// 计算两天之前的时间
	twoDaysAgo := now.AddDate(0, 0, client.I6000Conf.BeginDateDiff)
	beginTime:=twoDaysAgo.Format("2006-01-02")+ " 00:00:00"
	requestBody:=FindWorkTicketRequestBody{
		PlanBeginTimeStart:beginTime,
		AllOrgIds:client.I6000Conf.AllOrgIds,
		FlowProcessStepId:client.I6000Conf.UpdateWorkTicketStatusValue,
		PageNum:"1",
		PageSize:"100",
	}

	pageNum:=1

	for{
		requestBody.PageNum=strconv.Itoa(pageNum)
		rsp:=client.getWorkTicketByPage(&requestBody)
		if rsp==nil {
			break;
		}	

		if rsp.Success!=true {
			log.Println("I6000Client getWorkTicketByPage failed",rsp.Message)
			break;
		}

		if rsp.Data == nil {
			log.Println("I6000Client getWorkTicketByPage failed,no data")
			break;
		}

		if len(*rsp.Data)==0 {
			log.Println("I6000Client getWorkTicketByPage failed, data list is empty")
			break;
		}

		log.Println("I6000Client getWorkTicketByPage success,pageNum:",pageNum,"data count:",len(*rsp.Data))
		for _,item:=range *rsp.Data{
			workTickets=append(workTickets,item)
		}

		if len(*rsp.Data)<100 {
			break;
		}

		pageNum++
	}

	return &workTickets
}

func (client *I6000Client)getWorkTicket()(*[]WorkTicketItem){
	workTickets:=[]WorkTicketItem{}
	//查询开工日期在当前时间之后的工单
	now := time.Now()
	// 计算两天之前的时间
	twoDaysAgo := now.AddDate(0, 0, client.I6000Conf.BeginDateDiff)
	beginTime:=twoDaysAgo.Format("2006-01-02")+ " 00:00:00"
	requestBody:=FindWorkTicketRequestBody{
		PlanBeginTimeStart:beginTime,
		AllOrgIds:client.I6000Conf.AllOrgIds,
		FlowProcessStepId:client.I6000Conf.NewWorkTicketStatusLabel,
		PageNum:"1",
		PageSize:"100",
	}

	pageNum:=1

	for{
		requestBody.PageNum=strconv.Itoa(pageNum)
		rsp:=client.getWorkTicketByPage(&requestBody)
		if rsp==nil {
			break;
		}	

		if rsp.Success!=true {
			log.Println("I6000Client getWorkTicketByPage failed",rsp.Message)
			break;
		}

		if rsp.Data == nil {
			log.Println("I6000Client getWorkTicketByPage failed,no data")
			break;
		}

		if len(*rsp.Data)==0 {
			log.Println("I6000Client getWorkTicketByPage failed, data list is empty")
			break;
		}

		log.Println("I6000Client getWorkTicketByPage success,pageNum:",pageNum,"data count:",len(*rsp.Data))
		for _,item:=range *rsp.Data{
			workTickets=append(workTickets,item)
		}

		if len(*rsp.Data)<100 {
			break;
		}

		pageNum++
	}

	return &workTickets
}

func (client *I6000Client)Test(){
	//client.TestSignData()
	//client.TestGetExistWorkTickets()
	//client.TestGetDeviceRoomRack()
	//client.TestGetLockIds()
	//client.TestGetDeviceRackLockIds()
	//client.TestSaveNewWorkTicket()
	//client.TestCreateOpenLockApp()
	//client.TestDate()
	//client.TestQueryWorkTicket()
	//client.TestQueryApplication()
	//client.TestDeleteWorkTicket()
}

func (client *I6000Client)TestQueryApplication(){
	res:=client.queryApplication("1234")
	client.updateApplicationStatus(res)
}

func (client *I6000Client)TestDeleteWorkTicket(){
	workTicket:=WorkTicketItem{
		ID:"1234",
	}
	client.deleteWorkTicket(&workTicket)
}

func (client *I6000Client)TestQueryWorkTicket(){
	client.queryWorkTicket("1234")
}

func (client *I6000Client)TestCreateOpenLockApp(){
	workTicket:=WorkTicketItem{
		ID:"1234",
		Code:"1234",
		WorkPersionLiable:"1234",
		TeamName:"1234",
		TeamMemberCount:"1234",
		WorkSceneName:"1234",
		TeamMember:"1234",
		PlanBeginTime:"1234",
		PlanEndTime:"1234",
		FlowProcessStepName:"1234",
		WorkTask:"1234",
		NowHandleName:"1234",
		TaskOrTicketName:"1234",
		SafetyMeasures:"1234",  
	}
	lockList:=[]string{
		"00002300",
	}
	client.createOpenLockApp(&workTicket,&lockList)
}



func (client *I6000Client)TestSaveNewWorkTicket(){
	workTicket:=WorkTicketItem{
		ID:"1234",
		Code:"1234",
		WorkPersionLiable:"1234",
		TeamName:"1234",
		TeamMemberCount:"1234",
		WorkSceneName:"1234",
		TeamMember:"1234",
		PlanBeginTime:"1234",
		PlanEndTime:"1234",
		FlowProcessStepName:"1234",
		WorkTask:"1234",
		NowHandleName:"1234",
		TaskOrTicketName:"1234",
		SafetyMeasures:"1234",  
	}
	deviceInfo:=[]SystemInfoResponseData{
		SystemInfoResponseData{
			Name:"jf001机房C02机柜zzzzzzzz",
			ID:"1234",
			Category:"1234",
			Net:"1234",
			Type:"1234",
			Address:"1234",
			Person:"1234",
		},
		SystemInfoResponseData{
			Name:"jf001机房C02机柜zzzzzzzz",
			ID:"1235",
			Category:"1234",
			Net:"1234",
			Type:"1234",
			Address:"1234",
			Person:"1234",
		},
	}
	client.saveNewWorkTicket(&workTicket,&deviceInfo)		
}

func (client *I6000Client)TestGetDeviceRackLockIds(){
	deviceInfo:=[]SystemInfoResponseData{
		SystemInfoResponseData{
			Name:"jf001机房C01机柜zzzzzzzz",
		},
		SystemInfoResponseData{
			Name:"jf001机房C02机柜zzzzzzzz",
		},
		SystemInfoResponseData{
			Name:"jf001机房C03机柜zzzzzzzz",
		},
	}
	lockList:=client.getDeviceRackLockIds(&deviceInfo)
	log.Println("TestGetDeviceRackLockIds lockList:",lockList)
}

func (client *I6000Client)TestGetLockIds(){
	rackList:=[]RackItem{
		RackItem{
			Room:"jf001",
			Rack:"C02",
		},
	}
	res:=client.getLockIds(&rackList)
	log.Println("TestGetLockIds res:",res)
}

func (client *I6000Client)TestGetDeviceRoomRack(){
	reckItem:=client.getDeviceRoomRack("xxx机房yyyy机柜zzzzzzzz")
	log.Println("TestGetDeviceRoomRack reckItem:",reckItem)
}

func (client *I6000Client)TestGetExistWorkTickets(){
	workTickets:=&[]WorkTicketItem{
		WorkTicketItem{
			ID:"1",
		},
		WorkTicketItem{
			ID:"3",
		},
	}
	crvTickets:=client.GetExistWorkTickets(workTickets)
	log.Println("TestGetExistWorkTickets crvTickets:",*crvTickets)

	if len(*crvTickets)==len(*workTickets) {
		log.Println("I6000Client syncWorkTicket end with all work ticket exist")
		return
	}

	if len(*crvTickets)>0 && len(*crvTickets)<len(*workTickets) {
		workTickets=client.removeExistWorkTicket(workTickets,crvTickets)
	}

	log.Println("TestGetExistWorkTickets workTickets:",*workTickets)
}

func (client *I6000Client)TestSignData(){
	signData:=GetSignData(client.I6000Conf.GetSignDataUrl)
	log.Println("TestSignData singData:",signData)
}

func (client *I6000Client)TestDate(){
	now := time.Now()
	// 计算两天之前的时间
	twoDaysAgo := now.AddDate(0, 0, client.I6000Conf.BeginDateDiff)
	beginTime:=twoDaysAgo.Format("2006-01-02")+ " 00:00:00"
	log.Println("beginTime:",beginTime)
}