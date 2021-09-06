package auth

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

type User struct {
	Email string
	Pwd   string
	Perm  UserPerm
}
