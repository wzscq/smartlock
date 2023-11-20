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
)

type FindWorkTicketRequestBody struct {
	PlanBeginTimeStart string `json:"planBeginTimeStart"`
	//PlanBeginTimeEnd string `json:"planBeginTimeEnd"`
	PageNum string `json:"pageNum"`
	PageSize string `json:"pageSize"`
	AllOrgIds []string `json:"allOrgIds"`
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
	NowHandleName string `json:"nowHandleName"`
	TaskOrTicketName string `json:"taskOrTicketName"`
	SafetyMeasures string `json:"safetyMeasures"`     
}

type WorkTicketResponseData struct {
	Total int `json:"total"`
	List *[]WorkTicketItem `json:"list"` 
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
	Success int `json:"success"`
	Code string `json:"code"`
	Msg string `json:"msg"`
	Data *WorkTicketResponseData `json:"data"`
}

type I6000Client struct {
	CRVClient *crv.CRVClient
	I6000Conf *common.I6000Conf
}

func (client *I6000Client) Init() {
	go client.StartQueryWorkTicket()
}

func (client *I6000Client) StartQueryWorkTicket() {
	durationInterval,_:=time.ParseDuration(client.I6000Conf.QueryInterval)
	for{
		client.syncWorkTicket("")
		time.Sleep(durationInterval)
	}
}

func (client *I6000Client)getExistWorkTickets(workTickets *[]WorkTicketItem)(*[]string){
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
	req.Header.Set("AccessToken",client.I6000Conf.AccessToken)
	req.Header.Set("signData",client.I6000Conf.SignData)
	
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
	req.Header.Set("AccessToken",client.I6000Conf.AccessToken)
	req.Header.Set("signData",client.I6000Conf.SignData)
	
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


func (client *I6000Client)saveNewWorkTicket(workTicket *WorkTicketItem){
	systemInfo:=client.getWorkTicketSystems(workTicket.ID)
	if systemInfo==nil {
		log.Println("I6000Client saveNewWorkTicket WorkTicketSystems is nil workTicket id is ",workTicket.ID)
		return
	}
	if len(*systemInfo)==0 {
		log.Println("I6000Client saveNewWorkTicket WorkTicketSystems is empty workTicket id is ",workTicket.ID)
		return
	}

	deviceInfo:=client.getWorkTicketDevices(workTicket.ID)
	if deviceInfo==nil {
		log.Println("I6000Client saveNewWorkTicket WorkTicketDevices is nil workTicket id is ",workTicket.ID)
		return
	}

	if len(*deviceInfo)==0 {
		log.Println("I6000Client saveNewWorkTicket WorkTicketDevices is empty workTicket id is ",workTicket.ID)
		return
	}

	systemInfoList:=[]map[string]interface{}{}
	for _,system:=range *systemInfo {
		systemInfoList=append(systemInfoList,map[string]interface{}{
			"work_ticket_id":workTicket.ID,
			"system_id":system.ID,
			"name":system.Name,
			"category":system.Category,
			"net":system.Net,
			"type":system.Type,
			"address":system.Address,
			"person":system.Person,
			"_save_type":"create",
		})
	}

	deviceInfoList:=[]map[string]interface{}{}
	for _,device:=range *deviceInfo {
		deviceInfoList=append(deviceInfoList,map[string]interface{}{
			"work_ticket_id":workTicket.ID,
			"device_id":device.ID,
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
			"involve_systems":map[string]interface{}{
				"fieldType":"one2many",
        "relatedModelID":"sl_involve_system",
        "relatedField":"work_ticket_id",
				"list":systemInfoList,
			},
			"involve_devices":map[string]interface{}{
				"fieldType":"one2many",
        "relatedModelID":"sl_involve_device",
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

func (client *I6000Client)saveNewWorkTickets(workTickets *[]WorkTicketItem){
	log.Println("I6000Client saveNewWorkTickets start")
	for _,workTicket:=range *workTickets {
		client.saveNewWorkTicket(&workTicket)
	}
	log.Println("I6000Client saveNewWorkTickets end")
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
	crvTickets:=client.getExistWorkTickets(workTickets)
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
	req.Header.Set("AccessToken",client.I6000Conf.AccessToken)
	req.Header.Set("signData",client.I6000Conf.SignData)
	
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

func (client *I6000Client)getWorkTicket()(*[]WorkTicketItem){
	workTickets:=[]WorkTicketItem{}
	//查询开工日期在当前时间之后的工单
	beginTime:=time.Now().Format("2006-01-02")+ " 00:00:00"
	requestBody:=FindWorkTicketRequestBody{
		PlanBeginTimeStart:beginTime,
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

		if rsp.Success!=1 {
			log.Println("I6000Client getWorkTicketByPage failed",rsp.Msg)
			break;
		}

		if rsp.Data == nil {
			log.Println("I6000Client getWorkTicketByPage failed,no data")
			break;
		}

		if rsp.Data.List == nil {
			log.Println("I6000Client getWorkTicketByPage failed,no data list")
			break;
		}

		if len(*rsp.Data.List)==0 {
			log.Println("I6000Client getWorkTicketByPage failed, data list is empty")
			break;
		}

		for _,item:=range *rsp.Data.List{
			workTickets=append(workTickets,item)
		}

		if len(*rsp.Data.List)<100 {
			break;
		}

		pageNum++
	}

	return &workTickets
}