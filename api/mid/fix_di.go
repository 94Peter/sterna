package mid

import (
	"github.com/94peter/sterna"
	apiErr "github.com/94peter/sterna/api/err"
	"github.com/gin-gonic/gin"
)

type DevDIMiddle string

func NewGinFixDiMid(di interface{}, service string) GinMiddle {
	return &devDiMiddle{
		service: service,
		di:      di,
	}
}

type devDiMiddle struct {
	service string
	di      interface{}
}

func (lm *devDiMiddle) outputErr(c *gin.Context, err error) {
	apiErr.GinOutputErr(c, lm.service, err)
}

func (lm *devDiMiddle) GetName() string {
	return "di"
}

func (am *devDiMiddle) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set(string(sterna.CtxServDiKey), am.di)
		c.Next()
	}
}
