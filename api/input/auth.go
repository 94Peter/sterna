package input

import (
	"net/http"

	"github.com/94peter/sterna/dao"
	"github.com/94peter/sterna/util"
)

const (
	CtxUserInfoKey = util.CtxKey("userInfo")
)

type ReqUser interface {
	dao.LogUser
	GetId() string
	GetPerm() string
	GetDB() string
}

type reqUserImpl struct {
	Id      string
	Account string
	Name    string
	Perm    string
}

func (ru reqUserImpl) GetId() string {
	return ru.Id
}

func (ru reqUserImpl) GetDB() string {
	// ReqUser無userDB
	return ""
}

func (ru reqUserImpl) GetName() string {
	return ru.Name
}

func (ru reqUserImpl) GetAccount() string {
	return ru.Account
}

func (ru reqUserImpl) GetPerm() string {
	return ru.Perm
}

func NewReqUser(uid, acc, name, perm string) ReqUser {
	return reqUserImpl{
		Account: acc,
		Id:      uid,
		Name:    name,
		Perm:    perm,
	}
}

type AccessGuest interface {
	ReqUser
	GetSource() string
	GetSourceID() string
}

type accessGuestImpl struct {
	Source   string
	SourceID string
	DB       string
	Account  string
	Name     string
	Perm     string
}

func (ru accessGuestImpl) GetId() string {
	return ""
}

func (ru accessGuestImpl) GetDB() string {
	return ru.DB
}

func (ru accessGuestImpl) GetName() string {
	return ru.Name
}

func (ru accessGuestImpl) GetAccount() string {
	return ru.Account
}

func (ru accessGuestImpl) GetSource() string {
	return ru.Source
}

func (ru accessGuestImpl) GetSourceID() string {
	return ru.SourceID
}

func (ru accessGuestImpl) GetPerm() string {
	return ru.Perm
}

func NewAccessGuest(source, sid, acc, name, db, perm string) AccessGuest {
	return accessGuestImpl{
		Source:   source,
		SourceID: sid,
		DB:       db,
		Account:  acc,
		Name:     name,
		Perm:     perm,
	}
}

type CompanyUser interface {
	ReqUser
	GetCompID() string
	GetComp() string
}

type compUserImpl struct {
	reqUserImpl
	CompID string
	Comp   string
}

func (c compUserImpl) GetDB() string {
	return c.CompID
}

func (c compUserImpl) GetCompID() string {
	return c.CompID
}

func (c compUserImpl) GetComp() string {
	return c.Comp
}

func NewCompUser(uid, acc, name, compID, comp, perm string) CompanyUser {
	return compUserImpl{
		reqUserImpl: reqUserImpl{
			Account: acc,
			Id:      uid,
			Name:    name,
			Perm:    perm,
		},
		CompID: compID,
		Comp:   comp,
	}

}

func GetUserInfo(req *http.Request) ReqUser {
	ctx := req.Context()
	reqID := ctx.Value(CtxUserInfoKey)
	if ret, ok := reqID.(ReqUser); ok {
		return ret
	}
	if ret, ok := reqID.(AccessGuest); ok {
		return ret
	}
	if ret, ok := reqID.(CompanyUser); ok {
		return ret
	}
	return nil
}

func GetCompUserInfo(req *http.Request) CompanyUser {
	ctx := req.Context()
	reqID := ctx.Value(CtxUserInfoKey)
	if ret, ok := reqID.(CompanyUser); ok {
		return ret
	}
	return nil
}
