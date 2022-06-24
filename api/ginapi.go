package api

import (
	"net/http"

	"github.com/94peter/sterna/api/mid"
	"github.com/94peter/sterna/auth"
	"github.com/gin-gonic/gin"
)

type GinAPI interface {
	GetAPIs() []*GinApiHandler
	GetName() string
}

func NewGinApiServer(mode string) GinApiServer {
	gin.SetMode(mode)
	return &apiService{
		//Engine: gin.Default(),
		Engine: gin.New(),
	}
}

type GinApiHandler struct {
	Method  string
	Path    string
	Handler func(c *gin.Context)
	Auth    bool
	Group   []auth.UserPerm
}

type GinApiServer interface {
	AddAPIs(handlers ...GinAPI) GinApiServer
	Middles(mids ...mid.GinMiddle) GinApiServer
	SetAuth(authmid mid.AuthGinMidInter) GinApiServer
	Run(port string) error
}

type apiService struct {
	*gin.Engine
	authMid mid.AuthGinMidInter
}

func (serv *apiService) SetAuth(authMid mid.AuthGinMidInter) GinApiServer {
	serv.authMid = authMid
	return serv
}

func (serv *apiService) Middles(mids ...mid.GinMiddle) GinApiServer {
	for _, m := range mids {
		serv.Engine.Use(m.Handler())
	}

	return serv
}

func (serv *apiService) AddAPIs(apis ...GinAPI) GinApiServer {
	for _, api := range apis {
		for _, h := range api.GetAPIs() {
			switch h.Method {
			case "GET":
				serv.Engine.GET(h.Path, h.Handler)
			case "POST":
				serv.Engine.POST(h.Path, h.Handler)
			}
		}
	}
	// Listen and serve on 0.0.0.0:8080
	return serv
}

func (serv *apiService) Run(port string) error {
	// if serv.authMid != nil {
	// 	serv.Engine.Use(serv.authMid.Handler())
	// }
	return serv.Engine.Run(":" + port)
}

func GinOutputErr(c *gin.Context, err error) {
	if err == nil {
		return
	}
	if apiErr, ok := err.(ApiError); ok {
		GinOutputJson(c, apiErr.GetStatus(),
			map[string]interface{}{
				"status":   apiErr.GetStatus(),
				"title":    apiErr.GetErrorMsg(),
				"errorKey": apiErr.GetErrorKey(),
			})
	} else {
		GinOutputJson(c, http.StatusInternalServerError,
			map[string]interface{}{
				"status":   http.StatusInternalServerError,
				"title":    err.Error(),
				"errorKey": "",
			})
	}
	return
}

func GinOutputJson(c *gin.Context, statusCode int, data interface{}) {
	c.Header("Content-Type", "application/json")
	c.SecureJSON(statusCode, data)
}
