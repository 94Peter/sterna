package sterna

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"

	"github.com/94peter/sterna/util"

	yaml "gopkg.in/yaml.v3"
)

const (
	CtxServDiKey = util.CtxKey("ServiceDI")
)

func InitConfByFile(f string, di interface{}) {
	yamlFile, err := ioutil.ReadFile(f)
	if err != nil {
		fmt.Println("load conf fail: " + f)
		panic(err)
	}
	InitConfByByte(yamlFile, di)
}

func InitConfByByte(b []byte, di interface{}) {
	err := yaml.Unmarshal(b, di)
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

func InitConfByUri(uri string, di interface{}) error {
	resp, err := http.Get(uri)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	err = yaml.Unmarshal(body, di)
	if err != nil {
		return err
	}
	util.InitValidator()
	return nil
}

func GetDiByCtx(ctx context.Context) interface{} {
	diInter := ctx.Value(CtxServDiKey)
	return diInter
}
