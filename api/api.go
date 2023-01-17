package api

import (
	"encoding/json"
	"net/http"

	apiErr "github.com/94peter/sterna/api/err"
	"github.com/94peter/sterna/api/mid"
	"github.com/94peter/sterna/auth"

	"github.com/gorilla/mux"
)

func OutputJson(w http.ResponseWriter, statusCode int, data interface{}) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(data)
	if err != nil {
		apiErr.OutputErr(w, err)
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

func NewApiConf(port string, cors bool, ignoreMids []string, ignoreApis []string) *APIConf {
	conf := &APIConf{
		Port: port,
		Cors: cors,
	}
	if len(ignoreMids) > 0 {
		conf.Middle = make(map[string]bool)
	}
	for _, k := range ignoreMids {
		conf.Middle[k] = false
	}
	if len(ignoreApis) > 0 {
		conf.Apis = make(map[string]bool)
	}
	for _, k := range ignoreApis {
		conf.Apis[k] = false
	}

	return conf
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
