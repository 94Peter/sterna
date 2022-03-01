package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/94peter/sterna/api/mid"
	"github.com/94peter/sterna/auth"

	"github.com/gorilla/mux"
)

type ApiError interface {
	GetStatus() int
	GetErrorKey() string
	GetErrorMsg() string
	error
}

type myApiError struct {
	statusCode int
	message    string
	key        string
}

func (e myApiError) GetStatus() int {
	return e.statusCode
}

func (e myApiError) GetErrorKey() string {
	return e.key
}

func (e myApiError) GetErrorMsg() string {
	return e.message
}

func (e myApiError) Error() string {
	return fmt.Sprintf("%v: %v", e.statusCode, e.message)
}

func NewApiError(status int, msg string) ApiError {
	return myApiError{statusCode: status, message: msg}
}

func NewApiErrorWithKey(status int, msg string, key string) ApiError {
	return myApiError{statusCode: status, message: msg, key: key}
}

func OutputErr(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}
	if apiErr, ok := err.(ApiError); ok {
		w.WriteHeader(apiErr.GetStatus())
		OutputJson(w, map[string]interface{}{
			"status":   apiErr.GetStatus(),
			"title":    apiErr.GetErrorMsg(),
			"errorKey": apiErr.GetErrorKey(),
		})
	} else {
		w.WriteHeader(http.StatusInternalServerError)
		OutputJson(w, map[string]interface{}{
			"status":   apiErr.GetStatus(),
			"title":    apiErr.GetErrorMsg(),
			"errorKey": "",
		})
	}
	return
}

func OutputJson(w http.ResponseWriter, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		OutputErr(w, err)
	}
}

type APIHandler struct {
	Path   string
	Next   func(http.ResponseWriter, *http.Request)
	Method string
	Auth   bool
	Group  []auth.UserPerm
}

type API interface {
	GetAPIs() []*APIHandler
	GetName() string
}

type ApiDI interface {
	InitAPI(
		r *mux.Router,
		middles []mid.Middle,
		authMiddle mid.AuthMidInter,
		apis ...API,
	)
	GetPort() string
	EnableCORS() bool
}

type APIConf struct {
	Port   string          `yaml:"port,omitempty"`
	Cors   bool            `yaml:"cors"`
	Middle map[string]bool `yaml:"middle,omitempty"`
	Apis   map[string]bool `yaml:"api,omitempty"`
}

func (ac *APIConf) apiEnable(name string) bool {
	if v, ok := ac.Apis[name]; ok {
		return v
	}
	return true
}

func (ac *APIConf) middleEnable(name string) bool {
	if v, ok := ac.Middle[name]; ok {
		return v
	}
	return true
}

func (ac *APIConf) getMiddleList(ml []mid.Middle) []mid.Middleware {
	var middlewares []mid.Middleware
	for _, m := range ml {
		if !ac.middleEnable(m.GetName()) {
			continue
		}
		middlewares = append(middlewares, m.GetMiddleWare())
	}
	return middlewares
}

func (ac *APIConf) EnableCORS() bool {
	return ac.Cors
}
func (ac *APIConf) GetPort() string {
	return ac.Port
}

func (ac *APIConf) InitAPI(
	r *mux.Router,
	middles []mid.Middle,
	authMiddle mid.AuthMidInter,
	apis ...API,
) {
	if ac == nil {
		panic("api not set")
	}
	ml := ac.getMiddleList(middles)
	for _, myapi := range apis {
		if !ac.apiEnable(myapi.GetName()) {
			continue
		}
		for _, handler := range myapi.GetAPIs() {
			if authMiddle != nil {
				authMiddle.AddAuthPath(handler.Path, handler.Method, handler.Auth, handler.Group)
			}
			r.HandleFunc(handler.Path, mid.BuildChain(handler.Next, ml...)).Methods(handler.Method)
		}
	}
}
