package mqtt

import (
	"log"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"smartlockservice/common"
)

const (
	MSG_TYPE_DIAG="Diag"
	MSG_TYPE_EVENT="Event"
	MSG_TYPE_SIGNAL="SignalFilter"
)

type eventHandler interface {
	DealLockOperation(opMsg []byte)
	DealKeyControllerOperation(opMsg []byte)
}

type MQTTClient struct {
	Broker string
	User string
	Password string
	SendTopic string
	KeyControlAcceptTopic string
	ClientID string
	Handler eventHandler
	Client mqtt.Client
}

func (mqc *MQTTClient) getClient()(mqtt.Client){
	opts := mqtt.NewClientOptions()
	opts.AddBroker(mqc.Broker)
	opts.SetClientID(mqc.ClientID)
	opts.SetUsername(mqc.User)
	opts.SetPassword(mqc.Password)
	opts.SetAutoReconnect(true)

	opts.SetDefaultPublishHandler(mqc.messagePublishHandler)
	opts.OnConnect = mqc.connectHandler
	opts.OnConnectionLost = mqc.connectLostHandler
	opts.OnReconnecting = mqc.reconnectingHandler

	client:=mqtt.NewClient(opts)
	if token:=client.Connect(); token.Wait() && token.Error() != nil {
		log.Println(token.Error)
		return nil
	}
	return client
}

func (mqc *MQTTClient) connectHandler(client mqtt.Client){
	log.Println("MQTTClient connectHandler connect status: ",client.IsConnected())
}

func (mqc *MQTTClient) connectLostHandler(client mqtt.Client, err error){
	log.Println("MQTTClient connectLostHandler connect status: ",client.IsConnected(),err)
}

func (mqc *MQTTClient) messagePublishHandler(client mqtt.Client, msg mqtt.Message){
	log.Println("MQTTClient messagePublishHandler topic: ",msg.Topic())
}

func (mqc *MQTTClient) reconnectingHandler(Client mqtt.Client,opts *mqtt.ClientOptions){
	log.Println("MQTTClient reconnectingHandler ")
}

func (mqc *MQTTClient)onResponse(Client mqtt.Client, msg mqtt.Message){
	log.Println("MQTTClient onResponse ",msg.Topic())
	josnStr:=string(msg.Payload())
	log.Println("MQTTClient onResponse ",josnStr)
	mqc.Handler.DealLockOperation(msg.Payload())
}

func (mqc *MQTTClient)onKeyControlResponse(Client mqtt.Client, msg mqtt.Message){
	log.Println("MQTTClient onKeyControlResponse ",msg.Topic())
	josnStr:=string(msg.Payload())
	log.Println("MQTTClient onKeyControlResponse ",josnStr)
	mqc.Handler.DealKeyControllerOperation(msg.Payload())
}

func (mqc *MQTTClient)Publish(topic,content string)(int){
	if mqc.Client == nil {
		return common.ResultMqttClientError
	}
	log.Println("MQTTClient Publish topic:"+topic+" content:"+content)
	token:=mqc.Client.Publish(topic,0,false,content)
	token.Wait()
	return common.ResultSuccess
}

func (mqc *MQTTClient) Init(){
	mqc.Client=mqc.getClient()
	mqc.Client.Subscribe(mqc.SendTopic,0,mqc.onResponse)
	mqc.Client.Subscribe(mqc.KeyControlAcceptTopic,0,mqc.onKeyControlResponse)
}