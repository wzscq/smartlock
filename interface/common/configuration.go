package common

import (
	"log"
	"os"
	"encoding/json"
)

type serviceConf struct {
	Port string `json:"port"`
}

type mqttConf struct {
	Broker string `json:"broker"`
	User string `json:"user"`
	Password string `json:"password"`
	SendTopic string `json:"sendTopic"`
	AcceptTopic string `json:"acceptTopic"`
	ClientID string `json:"clientID"`
	KeyControlSendTopic string `json:"keyControlSendTopic"`
	KeyControlAcceptTopic string `json:"keyControlAcceptTopic"`
}

type crvConf struct {
	Server string `json:"server"`
    User string `json:"user"`
    Password string `json:"password"`
    AppID string `json:"appID"`
	Token string `json:"token"`
}

type I6000Conf struct {
	FindWorkTicket string `json:"findWorkTicket"`
	SelectInvolveSystemInfo string `json:"selectInvolveSystemInfo"`
	SelectInvolveDeviceInfo string `json:"selectInvolveDeviceInfo"`
	GetSignDataUrl string `json:"getSignDataUrl"`
	AllOrgIds []string `json:"allOrgIds"`
	QueryInterval string `json:"queryInterval"`
	BeginDateDiff int `json:"beginDateDiff"`
	NewWorkTicketStatusLabel string `json:"newWorkTicketStatusLabel"`
	UpdateWorkTicketStatusValue string `json:"updateWorkTicketStatusValue"`
}

type lockMonitorConf struct {
	HubPort string `json:"hubPort"`
	Interval string `json:"interval"`
	BatchInterval string `json:"batchInterval"`
	Timeout string `json:"timeout"`
}

type Config struct {
	Service serviceConf `json:"service"`
	CRV crvConf `json:"crv"`
	I6000Conf I6000Conf `json:"i6000"`
}

var gConfig Config

func InitConfig(confFile string)(*Config){
	log.Println("init configuation start ...")
	//获取用户账号
	//获取用户角色信息
	//根据角色过滤出功能列表
	fileName := confFile
	filePtr, err := os.Open(fileName)
	if err != nil {
        log.Fatal("Open file failed [Err:%s]", err.Error())
    }
    defer filePtr.Close()

	// 创建json解码器
    decoder := json.NewDecoder(filePtr)
    err = decoder.Decode(&gConfig)
	if err != nil {
		log.Println("json file decode failed [Err:%s]", err.Error())
	}
	log.Println("init configuation end")
	return &gConfig
}

func GetConfig()(*Config){
	return &gConfig
}