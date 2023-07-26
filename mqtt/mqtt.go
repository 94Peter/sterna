package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"
	"sync"
	"time"

	"github.com/94peter/sterna/log"
	"github.com/94peter/sterna/util"

	"github.com/google/uuid"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type MqttDI interface {
	NewMqttServ(serviceID string, l log.Logger) (MqttServ, error)
	NewMqttServWithFileStore(serviceID, clientID string, l log.Logger) (MqttServ, error)
}

type MqttServ interface {
	PublishByte(topics []string, data []byte, retain bool) error
	Publish(topics []string, message map[string]interface{}, retain bool) error
	Subcribe(mssm MqttSubSerInter) error
	SubscribeMultiple(mssm MqttSubSerMap) error
	Disconnect()

	onConnect(client mqtt.Client)
	onConnectLost(client mqtt.Client, err error)
}

type MqttConf struct {
	TCP           string `yaml:"tcp"`
	CaFile        string `yaml:"ca"`
	ClientCrt     string `yaml:"crt"`
	ClientKey     string `yaml:"key"`
	UserName      string `yaml:"user"`
	Password      string `yaml:"password"`
	FileStorePath string `yaml:"fileStorePath"`
}

func (mm *MqttConf) NewMqttServ(serviceID string, l log.Logger) (MqttServ, error) {
	return newBasicMqtt(mm, serviceID, l)
}

func (mm *MqttConf) NewMqttServWithFileStore(serviceID, clientID string, l log.Logger) (MqttServ, error) {
	return newDataResumeMqtt(mm, serviceID+clientID, clientID, l), nil
}

func (mm *MqttConf) newTlsConfig() *tls.Config {
	// Import trusted certificates from CAfile.pem.
	// Alternatively, manually add CA certificates to
	// default openssl CA bundle.
	certpool := x509.NewCertPool()
	pemCerts, err := ioutil.ReadFile(mm.CaFile)
	if err == nil {
		certpool.AppendCertsFromPEM(pemCerts)
	} else {
		panic(err)
	}

	if mm.ClientCrt != "" && mm.ClientKey != "" {
		// Import client certificate/key pair
		cert, err := tls.LoadX509KeyPair(mm.ClientCrt, mm.ClientKey)
		if err != nil {
			panic(err)
		}

		// // Just to print out the client certificate..
		cert.Leaf, err = x509.ParseCertificate(cert.Certificate[0])
		if err != nil {
			panic(err)
		}

		return &tls.Config{
			// RootCAs = certs used to verify server cert.
			RootCAs: certpool,
			// ClientAuth = whether to request cert from server.
			// Since the server is set up for SSL, this happens
			// anyways.
			ClientAuth: tls.NoClientCert,
			// ClientCAs = certs used to validate client cert.
			ClientCAs: nil,
			// InsecureSkipVerify = verify that cert contents
			// match server. IP matches what is in cert etc.
			InsecureSkipVerify: true,
			// Certificates = list of certs client sends to server.
			Certificates: []tls.Certificate{cert},
		}
	}
	// Create tls.Config with desired tls properties
	return &tls.Config{
		// RootCAs = certs used to verify server cert.
		RootCAs: certpool,
		// ClientAuth = whether to request cert from server.
		// Since the server is set up for SSL, this happens
		// anyways.
		ClientAuth: tls.NoClientCert,
		// ClientCAs = certs used to validate client cert.
		ClientCAs: nil,
		// InsecureSkipVerify = verify that cert contents
		// match server. IP matches what is in cert etc.
		InsecureSkipVerify: true,
		// Certificates = list of certs client sends to server.
		//	Certificates: []tls.Certificate{cert},
	}
}

type MqttSubSerInter interface {
	GetTopic() string
	GetQos() byte
	Handler(client mqtt.Client, msg mqtt.Message)
}

type MqttSubSerMap struct {
	subscribeMap map[string]byte
	handlerMap   map[string]func(client mqtt.Client, message mqtt.Message)
}

func (mssl *MqttSubSerMap) Add(mss MqttSubSerInter) {
	topic := mss.GetTopic()
	handlerTopic := topic
	if strings.Contains(topic, "$share") {
		splitStrAry := strings.SplitN(topic, "/", 3)
		if len(splitStrAry) != 3 {
			return
		}
		handlerTopic = splitStrAry[2]
	}
	if mssl.subscribeMap == nil {
		mssl.subscribeMap = make(map[string]byte)
	}
	if _, ok := mssl.subscribeMap[topic]; ok {
		return
	}

	mssl.subscribeMap[topic] = mss.GetQos()

	if mssl.handlerMap == nil {
		mssl.handlerMap = make(map[string]func(client mqtt.Client, message mqtt.Message))
	}
	mssl.handlerMap[handlerTopic] = mss.Handler
}

func (mssl MqttSubSerMap) GetSubscribeMap() map[string]byte {
	return mssl.subscribeMap
}

func (mssl MqttSubSerMap) Handler(client mqtt.Client, message mqtt.Message) {
	ser, ok := mssl.handlerMap[message.Topic()]
	if !ok {
		return
	}
	go ser(client, message)
}

func newBasicMqtt(mm *MqttConf, serviceID string, l log.Logger) (MqttServ, error) {
	basic := &basicMqttServImpl{
		MqttConf:  mm,
		serviceID: serviceID,
		log:       l,
	}
	for err := basic.connect(); err != nil; err = basic.connect() {
		l.Warn(fmt.Sprintln("connect err: ", serviceID, err.Error()))
	}
	return basic, nil
}

func newDataResumeMqtt(mm *MqttConf, serviceID, clientID string, l log.Logger) MqttServ {
	dataResumeServ := &dataResumeMqttServImpl{
		clientID: clientID,
	}
	basicServ := &basicMqttServImpl{
		MqttConf:    mm,
		serviceID:   serviceID,
		log:         l,
		connectFunc: dataResumeServ.connect,
	}
	dataResumeServ.basicMqttServImpl = basicServ
	for err := dataResumeServ.connect(); err != nil; err = dataResumeServ.connect() {
		l.Warn(fmt.Sprintln("connect err: ", serviceID, clientID, err.Error()))
	}
	return dataResumeServ
}

type basicMqttServImpl struct {
	wg sync.WaitGroup
	mu sync.Mutex

	serviceID string
	*MqttConf
	client      mqtt.Client
	connectFunc func() error
	log         log.Logger
}

func (mm *basicMqttServImpl) onConnect(client mqtt.Client) {
	if !client.IsConnectionOpen() {
		return
	}
	mm.client = client
	mm.log.Info(fmt.Sprintf("mqtt service [%s] connect", mm.serviceID))
}

func (mm *basicMqttServImpl) onConnectLost(client mqtt.Client, err error) {
	mm.log.Warn(fmt.Sprintf("mqtt service [%s] onconnectLost: %s", mm.serviceID, err.Error()))
	mm.wg.Done()
}

func (mm *basicMqttServImpl) connect() error {
	if mm.connectFunc != nil {
		return mm.connectFunc()
	}
	uid := uuid.New()
	clientID := fmt.Sprintf("%s-%s", mm.serviceID, uid.String())
	opts := mqtt.NewClientOptions().AddBroker(mm.TCP).SetClientID(clientID)

	opts.SetProtocolVersion(4)
	if mm.UserName != "" {
		opts = opts.SetUsername(mm.UserName)
	}
	if mm.Password != "" {
		opts = opts.SetPassword(mm.Password)
	}

	if mm.CaFile != "" {
		tls := mm.newTlsConfig()
		opts = opts.SetTLSConfig(tls)
	}
	opts.OnConnect = mm.onConnect
	opts.OnConnectionLost = mm.onConnectLost
	opts.SetCleanSession(true)
	opts.SetResumeSubs(false)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(time.Second * 10)
	opts.SetConnectRetry(true)
	opts.SetOrderMatters(false)
	mqc := mqtt.NewClient(opts)
	if token := mqc.Connect(); token.Wait() && token.Error() != nil {
		mm.log.Warn(fmt.Sprintf("mqtt service [%s] reconnect failed: %s", mm.serviceID, token.Error().Error()))
		return token.Error()
	}
	return nil
}

func (mm *basicMqttServImpl) PublishByte(topics []string, data []byte, retain bool) (err error) {
	for mm.client == nil {
		time.Sleep(time.Second * 10)
		return mm.PublishByte(topics, data, retain)
	}
	for _, topic := range topics {
		if token := mm.client.Publish(topic, 1, retain, data); token.WaitTimeout(time.Second*2) && token.Error() != nil {
			err = token.Error()
		}
	}
	return
}

func (mm *basicMqttServImpl) Publish(topics []string, message map[string]interface{}, retain bool) error {
	jsonData, _ := json.Marshal(message)
	return mm.PublishByte(topics, jsonData, retain)
}

func (mm *basicMqttServImpl) SubscribeMultiple(mssm MqttSubSerMap) error {
	if mm.client == nil {
		return errors.New("mqtt client is nil")
	}
	subMap := mssm.GetSubscribeMap()
	if len(subMap) == 0 {
		return errors.New("submap is empty")
	}
	token := mm.client.SubscribeMultiple(subMap, mssm.Handler)
	if token.Error() != nil {
		return token.Error()
	}
	mm.wg.Add(1)
	mm.wg.Wait()
	return nil
}

func (mm *basicMqttServImpl) Subcribe(mssi MqttSubSerInter) error {
	if mm.client == nil {
		return errors.New("mqtt client is nil")
	}
	if token := mm.client.Subscribe(mssi.GetTopic(), mssi.GetQos(), mssi.Handler); token.Wait() && token.Error() != nil {
		return token.Error()
	}
	mm.wg.Add(1)
	mm.wg.Wait()
	return nil
}

func (mm *basicMqttServImpl) Disconnect() {
	if mm.client != nil && mm.client.IsConnected() {
		mm.client.Disconnect(2)
	}
}

type dataResumeMqttServImpl struct {
	clientID string
	*basicMqttServImpl
	fileStore *mqtt.FileStore
}

func (mm *dataResumeMqttServImpl) connect() error {
	clientID := mm.clientID
	opts := mqtt.NewClientOptions().AddBroker(mm.TCP).SetClientID(clientID)

	opts.SetProtocolVersion(4)
	if mm.UserName != "" {
		opts = opts.SetUsername(mm.UserName)
	}
	if mm.Password != "" {
		opts = opts.SetPassword(mm.Password)
	}

	if mm.CaFile != "" {
		tls := mm.newTlsConfig()
		opts = opts.SetTLSConfig(tls)
	}
	if mm.FileStorePath == "" {
		return errors.New("missing fileStorePath")
	}
	if mm.fileStore == nil {
		mm.fileStore = mqtt.NewFileStore(util.StrAppend(mm.FileStorePath, "/", clientID))
	}
	opts = opts.SetStore(mm.fileStore)
	opts.OnConnect = mm.onConnect
	opts.OnConnectionLost = mm.onConnectLost
	opts.SetCleanSession(false)
	opts.SetResumeSubs(false)
	opts.SetAutoReconnect(true)
	opts.SetMaxReconnectInterval(time.Second * 10)
	opts.SetConnectRetry(true)
	mqc := mqtt.NewClient(opts)

	if token := mqc.Connect(); token.Wait() && token.Error() != nil {
		mm.log.Warn(fmt.Sprintf("mqtt service [%s] reconnect failed: %s", mm.serviceID, token.Error().Error()))
		return token.Error()
	}
	return nil
}
