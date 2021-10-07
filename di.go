package sterna

import (
	"fmt"
	"io/ioutil"
	"os"
	"time"

	"github.com/94peter/sterna/db"
	"github.com/94peter/sterna/export"
	"github.com/94peter/sterna/gcp"
	"github.com/94peter/sterna/mail"

	"github.com/94peter/sterna/api"
	"github.com/94peter/sterna/auth"
	"github.com/94peter/sterna/log"
	"github.com/94peter/sterna/util"

	yaml "gopkg.in/yaml.v2"
)

type DI interface {
	db.MongoDI
	log.LoggerDI
	api.ApiDI
	auth.TransmitSecurity
	auth.JwtDI
	gcp.Storage
	gcp.PubSubConf
	mail.MailConf
	export.ExportConf
	LocationDI
}

type di struct {
	db.MongoConf               `yaml:"mongo,omitempty"`
	*log.LoggerConf            `yaml:"log,omitempty"`
	*api.APIConf               `yaml:"api,omitempty"`
	*auth.TransmitSecurityConf `yaml:"transmitSecurity"`
	*auth.JwtConf              `yaml:"jwtConf"`
	*gcp.GcpConf               `yaml:"gcp"`
	*mail.SendGridConf         `yaml:"mail"`
	*export.MyExportConf       `yaml:"export"`
	location                   *time.Location
}

var mydi *di

func GetDI() DI {
	if mydi == nil {
		panic("not init di")
	}
	return mydi
}

func InitConfByFile(f string) {

	yamlFile, err := ioutil.ReadFile(f)
	if err != nil {
		fmt.Println("load conf fail: " + f)
		panic(err)
	}
	mydi = &di{}
	err = yaml.Unmarshal(yamlFile, mydi)
	if err != nil {
		panic(err)
	}
	util.InitValidator()
}

var _env = ""

// 初始化設定檔，讀YAML檔
func IniConfByEnv(path, env string) {
	const confFileTpl = "%s/%s/config.yml"
	_env = env
	InitConfByFile(fmt.Sprintf(confFileTpl, path, env))
}

func InitDefaultConf(path string) {
	InitConfByFile(util.StrAppend(path, "/conf/config.yml"))
}

func IsTest() bool {
	return _env == "test"
}

func SystemPwd() string {
	md5pwd := os.Getenv("ENCR_WORD")
	return md5pwd
}

func (d *di) Close(key string) {

}

func (d *di) SetLocation(timezone string) {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		panic(err)
	}
	d.location = loc
}

func (d *di) Location() *time.Location {
	if d.location != nil {
		return d.location
	}
	return time.Now().Location()
}

type modbusConf struct {
	ReadDuration  string `yaml:"rd"`
	WriteDuration string `yaml:"wd"`
}

func (mc modbusConf) GetReadDuration() time.Duration {
	d, err := time.ParseDuration(mc.ReadDuration)
	if err != nil {
		panic(err)
	}
	return d
}

func (mc modbusConf) GetWriteDuration() time.Duration {
	d, err := time.ParseDuration(mc.WriteDuration)
	if err != nil {
		panic(err)
	}
	return d
}

type LocationDI interface {
	SetLocation(loc string)
	Location() *time.Location
}
