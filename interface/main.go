package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"smartlockservice/common"
	"smartlockservice/crv"
	"smartlockservice/i6000"
	"log"
	"os"
)

func main() {
	//设置log打印文件名和行号
  log.SetFlags(log.Lshortfile | log.LstdFlags)

	confFile:="conf/conf.json"
	if len(os.Args)>1 {
			confFile=os.Args[1]
			log.Println(confFile)
	}
	
	conf:=common.InitConfig(confFile)

	router := gin.Default()
	router.Use(cors.New(cors.Config{
		AllowAllOrigins:true,
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
  	}))

   	//crvClinet 用于到crvframeserver的请求
	crvClinet:=&crv.CRVClient{
		Server:conf.CRV.Server,
		Token:conf.CRV.Token,
		AppID:conf.CRV.AppID,
	}

	i6000Client:=&i6000.I6000Client{
		CRVClient:crvClinet,
		I6000Conf:&conf.I6000Conf,
	}
	//i6000Client.Test()
	i6000Client.Init()
	i6000Controller:=&i6000.I6000Controller{
		I6000Client:i6000Client,
	}
	i6000Controller.Bind(router)

	router.Run(conf.Service.Port) // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}