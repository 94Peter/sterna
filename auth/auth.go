package auth

import (
	"net/http"

	"github.com/94peter/sterna/dao"
	"github.com/94peter/sterna/util"
)

type UserPerm string

func (up UserPerm) Validate() bool {
	switch up {
	case PermAdmin, PermOwner, PermEditor, PermViewer, PermGuest:
		return true
	default:
		return false
	}
}

const (
	// 管理者
	PermAdmin = UserPerm("admin")
	// 會員
	PermMember = UserPerm("member")
	// 擁有
	PermOwner = UserPerm("owner")
	// 編輯
	PermEditor = UserPerm("editor")
	// 檢視
	PermViewer = UserPerm("viewer")
	// 訪客
	PermGuest = UserPerm("guest")
)

const (
	CtxUserInfoKey = util.CtxKey("userInfo")
)

type ReqUser interface {
	dao.LogUser
	Host() string
	GetId() string
	GetPerm() string
	GetDB() string
}

type reqUserImpl struct {
	host string
	id   string
	acc  string
	name string
	perm string
}

func (ru *reqUserImpl) Host() string {
	return ru.host
}

func (ru reqUserImpl) GetId() string {
	return ru.id
}

func (ru reqUserImpl) GetDB() string {
	// ReqUser無userDB
	return ""
}

func (ru reqUserImpl) GetName() string {
	return ru.name
}

func (ru reqUserImpl) GetAccount() string {
	return ru.acc
}

func (ru reqUserImpl) GetPerm() string {
	return ru.perm
}

func NewReqUser(host, uid, acc, name, perm string) ReqUser {
	return &reqUserImpl{
		host: host,
		acc:  acc,
		id:   uid,
		name: name,
		perm: perm,
	}
}

type AccessGuest interface {
	ReqUser
	GetSource() string
	GetSourceID() string
}

type accessGuestImpl struct {
	host     string
	source   string
	sourceID string
	dB       string
	account  string
	name     string
	perm     string
}

func (ru *accessGuestImpl) Host() string {
	return ru.host
}

func (ru *accessGuestImpl) GetId() string {
	return ""
}

func (ru *accessGuestImpl) GetDB() string {
	return ru.dB
}

func (ru *accessGuestImpl) GetName() string {
	return ru.name
}

func (ru *accessGuestImpl) GetAccount() string {
	return ru.account
}

func (ru *accessGuestImpl) GetSource() string {
	return ru.source
}

func (ru *accessGuestImpl) GetSourceID() string {
	return ru.sourceID
}

func (ru *accessGuestImpl) GetPerm() string {
	return ru.perm
}

func NewAccessGuest(host, source, sid, acc, name, db, perm string) AccessGuest {
	return &accessGuestImpl{
		host:     host,
		source:   source,
		sourceID: sid,
		dB:       db,
		account:  acc,
		name:     name,
		perm:     perm,
	}
}

type CompanyUser interface {
	ReqUser
	GetCompID() string
	GetComp() string
}

type compUserImpl struct {
	*reqUserImpl
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

func NewCompUser(host, uid, acc, name, compID, comp, perm string) CompanyUser {
	return compUserImpl{
		reqUserImpl: &reqUserImpl{
			host: host,
			acc:  acc,
			id:   uid,
			name: name,
			perm: perm,
		},
		CompID: compID,
		Comp:   comp,
	}
}

type guestUser struct {
	host string
	ip   string
}

func NewGuestUser(host, ip string) ReqUser {
	return &guestUser{
		host: host,
		ip:   ip,
	}
}

func (ru *guestUser) Host() string {
	return ru.host
}

func (ru *guestUser) GetId() string {
	return ru.ip
}

func (ru *guestUser) GetDB() string {
	// ReqUser無userDB
	return ""
}

func (ru *guestUser) GetName() string {
	return ru.ip
}

func (ru *guestUser) GetAccount() string {
	return ru.ip
}

func (ru *guestUser) GetPerm() string {
	return string(PermGuest)
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
