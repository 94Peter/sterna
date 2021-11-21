package sterna

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/94peter/sterna/util"

	yaml "gopkg.in/yaml.v2"
)

func InitConfByFile(f string, di interface{}) {
	yamlFile, err := ioutil.ReadFile(f)
	if err != nil {
		fmt.Println("load conf fail: " + f)
		panic(err)
	}

	err = yaml.Unmarshal(yamlFile, di)
	if err != nil {
		panic(err)
	}
	util.InitValidator()
}

// 初始化設定檔，讀YAML檔
func IniConfByEnv(path, env string, di interface{}) {
	const confFileTpl = "%s/%s/config.yml"
	InitConfByFile(fmt.Sprintf(confFileTpl, path, env), di)
}

func InitDefaultConf(path string, di interface{}) {
	InitConfByFile(util.StrAppend(path, "/conf/config.yml"), di)
}

// type DI interface {
// 	db.MongoDI
// 	log.LoggerDI
// 	api.ApiDI
// 	auth.TransmitSecurity
// 	auth.JwtDI
// 	gcp.Storage
// 	gcp.PubSubConf
// 	mail.MailConf
// 	export.ExportConf
// 	LocationDI
// }

// type di struct {
// 	*db.MongoConf              `yaml:"mongo,omitempty"`
// 	*log.LoggerConf            `yaml:"log,omitempty"`
// 	*api.APIConf               `yaml:"api,omitempty"`
// 	*auth.TransmitSecurityConf `yaml:"transmitSecurity"`
// 	*auth.JwtConf              `yaml:"jwtConf"`
// 	*gcp.GcpConf               `yaml:"gcp"`
// 	*mail.SendGridConf         `yaml:"mail"`
// 	*export.MyExportConf       `yaml:"export"`
// 	location                   *time.Location
// }

func GetENV() string {
	return os.Getenv("env")
}

// func SystemPwd() string {
// 	md5pwd := os.Getenv("ENCR_WORD")
// 	return md5pwd
// }

// func (d *di) Close(key string) {

// }

// func (d *di) SetLocation(timezone string) {
// 	loc, err := time.LoadLocation(timezone)
// 	if err != nil {
// 		panic(err)
// 	}
// 	d.location = loc
// }

// func (d *di) Location() *time.Location {
// 	if d.location != nil {
// 		return d.location
// 	}
// 	return time.Now().Location()
// }

// type modbusConf struct {
// 	ReadDuration  string `yaml:"rd"`
// 	WriteDuration string `yaml:"wd"`
// }

// func (mc modbusConf) GetReadDuration() time.Duration {
// 	d, err := time.ParseDuration(mc.ReadDuration)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return d
// }

// func (mc modbusConf) GetWriteDuration() time.Duration {
// 	d, err := time.ParseDuration(mc.WriteDuration)
// 	if err != nil {
// 		panic(err)
// 	}
// 	return d
// }

// type LocationDI interface {
// 	SetLocation(loc string)
// 	Location() *time.Location
// }
