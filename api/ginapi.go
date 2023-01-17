package api

import (
	apiErr "github.com/94peter/sterna/api/err"
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
			if serv.authMid != nil {
				serv.authMid.AddAuthPath(h.Path, h.Method, h.Auth, h.Group)
			}
			switch h.Method {
			case "GET":
				serv.Engine.GET(h.Path, h.Handler)
			case "POST":
				serv.Engine.POST(h.Path, h.Handler)
			case "PUT":
				serv.Engine.PUT(h.Path, h.Handler)
			case "DELETE":
				serv.Engine.DELETE(h.Path, h.Handler)
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

type ErrorOutputAPI interface {
	GinOutputErr(c *gin.Context, err error)
}

func NewErrorOutputAPI(service string) ErrorOutputAPI {
	return &commonAPI{
		service: service,
	}
}

type commonAPI struct {
	service string
}

func (capi *commonAPI) GinOutputErr(c *gin.Context, err error) {
	apiErr.GinOutputErr(c, capi.service, err)
}
