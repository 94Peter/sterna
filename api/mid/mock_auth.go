package mid

import (
	"strings"

	"github.com/94peter/sterna/auth"
	"github.com/94peter/sterna/util"
	"github.com/gin-gonic/gin"
)

func NewMockAuthMid() AuthGinMidInter {
	return &mockAuthMiddle{}
}

type mockAuthMiddle struct {
}

func (lm *mockAuthMiddle) GetName() string {
	return "mockAuth"
}

func (am *mockAuthMiddle) AddAuthPath(path string, method string, isAuth bool, group []auth.UserPerm) {
}

func (am *mockAuthMiddle) Handler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Request = util.SetCtxKeyVal(
			c.Request,
			auth.CtxUserInfoKey,
			auth.NewReqUser(
				util.GetHost(c.Request),
				c.GetHeader("Mock_User_UID"),
				c.GetHeader("Mock_User_ACC"),
				c.GetHeader("Mock_User_NAM"),
				strings.Split(c.GetHeader("Mock_User_Roles"), ","),
			))
		c.Next()
	}
}
