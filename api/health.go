package api

import (
	"net/http"
	"time"

	"github.com/94peter/sterna/db"
	"github.com/gin-gonic/gin"
)

func NewHealthAPI(service string) GinAPI {
	return &healthAPI{
		ErrorOutputAPI: NewErrorOutputAPI(service),
	}
}

type healthAPI struct {
	ErrorOutputAPI
}

func (a *healthAPI) GetName() string {
	return "health"
}

func (a *healthAPI) GetAPIs() []*GinApiHandler {
	return []*GinApiHandler{
		// health check
		{Method: "GET", Path: "__health", Handler: a.healthHandler, Auth: false},
	}
}

func (a *healthAPI) healthHandler(c *gin.Context) {
	healthResp := healthResponse{
		Status: "ok",
	}
	dbclt := db.GetMgoDBClientByGin(c)
	err := dbclt.Ping()
	if err != nil {
		healthResp.Connection.Mongo.Status = "red"
		healthResp.Connection.Mongo.Msg = err.Error()
	} else {
		healthResp.Connection.Mongo.Status = "green"
		healthResp.Connection.Mongo.Msg = "ok"
	}
	healthResp.Now = time.Now()
	c.JSON(http.StatusOK, healthResp)
}

type healthResponse struct {
	Now        time.Time       `json:"now"`
	Status     string          `json:"status"`
	Connection connectionState `json:"connection"`
}

type connectionState struct {
	Mongo struct {
		Status string `json:"status"`
		Msg    string `json:"msg"`
	} `json:"mongo"`
}
