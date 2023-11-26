package i6000

import (
	"smartlockservice/common"
	"smartlockservice/crv"
	"github.com/gin-gonic/gin"
	"net/http"
	"log"
)

type I6000Controller struct {
	I6000Client *I6000Client
}

func (controller *I6000Controller) syncWorkTicket(c *gin.Context) {
	var header crv.CommonHeader
	if err := c.ShouldBindHeader(&header); err != nil {
		log.Println(err)
		rsp:=common.CreateResponse(common.CreateError(common.ResultWrongRequest,nil),nil)
		c.IndentedJSON(http.StatusOK, rsp)
		log.Println("end syncWorkTicket with error")
		return
	}	

	err:=controller.I6000Client.syncWorkTicket(header.Token)
	if err!=common.ResultSuccess {
		rsp:=common.CreateResponse(common.CreateError(err,nil),nil)
		c.IndentedJSON(http.StatusOK, rsp)
		return
	}

	rsp:=common.CreateResponse(nil,nil)
	c.IndentedJSON(http.StatusOK, rsp)
	log.Println("I6000Controller end syncWorkTicket")
}

//Bind bind the controller function to url
func (controller *I6000Controller) Bind(router *gin.Engine) {
	log.Println("Bind I6000Controller")
	router.POST("/syncWorkTicket", controller.syncWorkTicket)
}