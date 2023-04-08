package lockservice

import (
	"smartlockservice/common"
	"smartlockservice/lock"
	"smartlockservice/lockhub"
	"smartlockservice/crv"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
)

type SmartLockController struct {
	LockOperator *lock.LockOperator
	LockStatusMonitor *lockhub.LockStatusMonitor
}

func (controller *SmartLockController)open(c *gin.Context){
	log.Println("SmartLockController start open")
	
	var header crv.CommonHeader
	if err := c.ShouldBindHeader(&header); err != nil {
		log.Println(err)
		rsp:=common.CreateResponse(common.CreateError(common.ResultWrongRequest,nil),nil)
		c.IndentedJSON(http.StatusOK, rsp)
		log.Println("end redirect with error")
		return
	}	
	
	var rep crv.CommonReq
	if err := c.BindJSON(&rep); err != nil {
		log.Println(err)
		rsp:=common.CreateResponse(common.CreateError(common.ResultWrongRequest,nil),nil)
		c.IndentedJSON(http.StatusOK, rsp)
		return
  	}	

	if rep.List==nil || len(*rep.List)==0 {
		rsp:=common.CreateResponse(common.CreateError(common.ResultWrongRequest,nil),nil)
		c.IndentedJSON(http.StatusOK, rsp)
		log.Println("error：request list is empty")
		return
	}

	//发送消息
	for _,row:=range *rep.List {
		closeDelay,_:=row["close_delay"].(string)
		locksField,_:=row["locks"].(map[string]interface{})
		locksList,_:=locksField["list"].([]interface{})
		for _,lockItem:=range locksList {
			//log.Println(lockItem)
			lockID:=lockItem.(map[string]interface{})["id"].(string)
			//err:=controller.LockOperator.Open(lockID)
			err:=controller.LockStatusMonitor.Open(header.Token,lockID,closeDelay)
			if err!=common.ResultSuccess {
				rsp:=common.CreateResponse(common.CreateError(err,nil),nil)
				c.IndentedJSON(http.StatusOK, rsp)
				return
			}			
		}
	}

	//保存数据
	controller.LockOperator.CRVClient.Save(&rep,header.Token)

	rsp:=common.CreateResponse(nil,nil)
	c.IndentedJSON(http.StatusOK, rsp)
	log.Println("SmartLockController end open")
}

func (controller *SmartLockController)writekey(c *gin.Context){
	log.Println("SmartLockController start writekey")
	
	var header crv.CommonHeader
	if err := c.ShouldBindHeader(&header); err != nil {
		log.Println(err)
		rsp:=common.CreateResponse(common.CreateError(common.ResultWrongRequest,nil),nil)
		c.IndentedJSON(http.StatusOK, rsp)
		log.Println("end writekey with error")
		return
	}	
	
	var rep crv.CommonReq
	if err := c.BindJSON(&rep); err != nil {
		log.Println(err)
		rsp:=common.CreateResponse(common.CreateError(common.ResultWrongRequest,nil),nil)
		c.IndentedJSON(http.StatusOK, rsp)
		return
  }	

	if rep.SelectedRowKeys ==nil || len(*rep.SelectedRowKeys)==0 {
		rsp:=common.CreateResponse(common.CreateError(common.ResultWrongRequest,nil),nil)
		c.IndentedJSON(http.StatusOK, rsp)
		log.Println("SmartLockController end writekey with error：request list is empty")
		return
	}

	//发送消息
	applicationID:=(*rep.SelectedRowKeys)[0]
	err:=controller.LockOperator.WriteKey(applicationID,header.Token)
	if err!=common.ResultSuccess {
		rsp:=common.CreateResponse(common.CreateError(err,nil),nil)
		c.IndentedJSON(http.StatusOK, rsp)
		return
	}
	
	rsp:=common.CreateResponse(nil,nil)
	c.IndentedJSON(http.StatusOK, rsp)
	log.Println("SmartLockController end writekey")
}

/*func (controller *SmartLockController)syncLockList(c *gin.Context){
	log.Println("SmartLockController start syncLockList")

	var header crv.CommonHeader
	if err := c.ShouldBindHeader(&header); err != nil {
		log.Println(err)
		rsp:=common.CreateResponse(common.CreateError(common.ResultWrongRequest,nil),nil)
		c.IndentedJSON(http.StatusOK, rsp)
		log.Println("end writekey with error")
		return
	}	

	err:=controller.LockOperator.SyncLockList(header.Token)
	if err!=common.ResultSuccess {
		rsp:=common.CreateResponse(common.CreateError(err,nil),nil)
		c.IndentedJSON(http.StatusOK, rsp)
		return
	}

	rsp:=common.CreateResponse(nil,nil)
	c.IndentedJSON(http.StatusOK, rsp)
	log.Println("SmartLockController end syncLockList")
}*/

func (controller *SmartLockController)syncLockList(c *gin.Context){
	log.Println("SmartLockController start syncLockList")

	var header crv.CommonHeader
	if err := c.ShouldBindHeader(&header); err != nil {
		log.Println(err)
		rsp:=common.CreateResponse(common.CreateError(common.ResultWrongRequest,nil),nil)
		c.IndentedJSON(http.StatusOK, rsp)
		log.Println("end writekey with error")
		return
	}	

	err:=controller.LockStatusMonitor.UpdateLockList(header.Token)
	if err!=common.ResultSuccess {
		rsp:=common.CreateResponse(common.CreateError(err,nil),nil)
		c.IndentedJSON(http.StatusOK, rsp)
		return
	}

	rsp:=common.CreateResponse(nil,nil)
	c.IndentedJSON(http.StatusOK, rsp)
	log.Println("SmartLockController end syncLockList")
}

//Bind bind the controller function to url
func (controller *SmartLockController) Bind(router *gin.Engine) {
	log.Println("Bind SmartLockController")
	router.POST("/open", controller.open)
	router.POST("/writekey", controller.writekey)
	router.POST("/syncLockList", controller.syncLockList)
}