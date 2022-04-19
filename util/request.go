package util

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"reflect"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/mitchellh/mapstructure"
)

func GetClientKey(req *http.Request) string {
	deviceID := req.Header.Get("X-Device-ID")
	if deviceID != "" {
		return MD5(deviceID)
	}

	real := req.Header.Get("X-Real-IP")
	if real == "" {
		real = req.RemoteAddr
	}
	return MD5(real + req.UserAgent())
}

func IsLogin(req *http.Request) bool {
	isLoginStr := req.Header.Get("isLogin")
	isLogin, err := strconv.ParseBool(isLoginStr)
	if err != nil {
		return false
	}
	return isLogin
}

func DecodeTokenByKey(req *http.Request, key string) map[string]interface{} {
	token := req.Header.Get(key)
	if token == "" {
		return nil
	}
	return DecodeToken(token)
}

func DecodeToken(token string) map[string]interface{} {
	mapSerialize, err := DecodeMap(token)
	if err != nil {
		return nil
	}
	return *mapSerialize
}

func ParserDataRequest(req *http.Request, data interface{}) error {
	kindOfJ := reflect.ValueOf(data).Kind()
	if kindOfJ != reflect.Ptr {
		return errors.New("data is not pointer")
	}
	switch req.Header.Get("Content-Type") {
	case "application/json":
		err := json.NewDecoder(req.Body).Decode(data)
		if err != nil {
			return err
		}
	case "application/x-www-form-urlencoded":
		vars, err := GetPostValue(req, true, []string{"pwd"})
		if err != nil {
			return err
		}
		err = mapstructure.Decode(vars, data)
		if err != nil {
			return err
		}
	}
	return nil
}

func GetQueryValue(req *http.Request, keys []string, defaultEmpty bool) map[string]interface{} {
	queries := req.URL.Query()
	result := make(map[string]interface{})

	for _, key := range keys {
		value, ok := queries[key]
		if !ok {
			// if key not exist. use empty string
			if defaultEmpty {
				result[key] = ""
			}
			continue
		}
		if len(value) == 1 {
			result[key] = value[0]
		} else {
			result[key] = value
		}
	}
	return result
}

func GetPostValue(req *http.Request, defaultEmpty bool, keys []string) (map[string]interface{}, error) {
	err := req.ParseForm()
	if err != nil {
		return nil, err
	}
	result := make(map[string]interface{})
	for _, key := range keys {
		if vs := req.PostForm[key]; len(vs) > 0 {
			result[key] = vs[0]
		} else if defaultEmpty {
			result[key] = ""
		}
	}
	return result, nil
}

type RequestFile struct {
	ReqFile   multipart.File
	ReqHeader *multipart.FileHeader
}

func GetMutiFormPostValue(req *http.Request, fileKeys []string, valueKeys []string) (map[string]RequestFile, map[string]interface{}, error) {
	req.ParseMultipartForm(32 << 20)

	fileMap := make(map[string]RequestFile)
	for _, fk := range fileKeys {
		file, handler, err := req.FormFile(fk)
		if err != nil {
			for _, value := range fileMap {
				defer value.ReqFile.Close()
			}
			return nil, nil, err
		}
		fileMap[fk] = RequestFile{file, handler}
	}

	valueMap := make(map[string]interface{})
	for _, vk := range valueKeys {
		valueMap[vk] = req.FormValue(vk)
	}
	return fileMap, valueMap, nil
}

func GetPathVars(req *http.Request, keys []string) map[string]interface{} {
	vars := mux.Vars(req)
	if len(vars) == 0 {
		return nil
	}
	valueMap := make(map[string]interface{})
	for _, key := range keys {
		if v, ok := vars[key]; ok {
			valueMap[key] = v
		} else {
			valueMap[key] = ""
		}
	}
	if len(valueMap) == 0 {
		return nil
	}
	return valueMap
}

type CtxKey string

func GetCtxVal(req *http.Request, ck CtxKey) interface{} {
	ctx := req.Context()
	return ctx.Value(ck)
}

func SetCtxKeyVal(r *http.Request, ck CtxKey, val interface{}) *http.Request {
	ctx := context.WithValue(r.Context(), ck, val)
	return r.WithContext(ctx)
}

func GetFullUrlStr(req *http.Request) string {
	prot := "http"
	if req.TLS != nil {
		prot = "https"
	}
	return fmt.Sprintf("%s://%s%s", prot, req.Host, req.RequestURI)
}

func GetClientInfo(req *http.Request) map[string]interface{} {
	result := map[string]interface{}{
		"UserAgent": req.UserAgent(),
	}
	return result
}

//ReturnExist will return new if new is not empty else return original
func ReturnExist(ori interface{}, new interface{}) interface{} {

	switch dtype := reflect.TypeOf(ori).Kind(); dtype {
	case reflect.Int:
		if new.(int) == 0 {
			return ori
		}
		return new

	case reflect.Int64:
		if new.(int64) == 0 {
			return ori
		}
		return new

	case reflect.String:
		if new.(string) == "" {
			return ori
		}
		return new

	case reflect.Array:
		if reflect.ValueOf(new).Len() == 0 {
			return ori
		}
		return new

	case reflect.Slice:
		if reflect.ValueOf(new).Len() == 0 {
			return ori
		}
		return new

	case reflect.Struct:
		f := reflect.New(reflect.TypeOf(ori)).Elem().Interface()
		if f == new {
			return ori
		}
		return new

	}
	return nil
}
