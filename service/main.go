package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gin-contrib/cors"
	"smartlockservice/common"
	"smartlockservice/mqtt"
	"smartlockservice/crv"
	"smartlockservice/lock"
	"smartlockservice/lockservice"
	"smartlockservice/lockhub"
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

	//mqttclient
	mqttClient:=&mqtt.MQTTClient{
		Broker:conf.MQTT.Broker,
		User:conf.MQTT.User,
		Password:conf.MQTT.Password,
		SendTopic:conf.MQTT.SendTopic,
		//Handler:&busiModule,
		ClientID:conf.MQTT.ClientID,
	}

	lockStatusMonitor:=&lockhub.LockStatusMonitor{
		CRVClient:crvClinet,
		MQTTClient:mqttClient,
		Interval:conf.Monitor.Interval,
		BatchInterval:conf.Monitor.BatchInterval,
		HubPort:conf.Monitor.HubPort,
		Timeout:conf.Monitor.Timeout,
	}
	
	lockStatusMonitor.UpdateLockList("")
	lockOperator:=&lock.LockOperator{
		MQTTClient:mqttClient,
		AcceptTopic:conf.MQTT.AcceptTopic,
		CRVClient:crvClinet,
		KeyControlSendTopic:conf.MQTT.KeyControlSendTopic,
	}

	mqttClient.Handler=lockOperator

	slController:=&lockservice.SmartLockController{
		LockOperator:lockOperator,
		LockStatusMonitor:lockStatusMonitor,
	}

	mqttClient.Init()
	slController.Bind(router)
	
	go lockStatusMonitor.StartMonitor()
		
	router.Run(conf.Service.Port) // listen and serve on 0.0.0.0:8080 (for windows "localhost:8080")
}