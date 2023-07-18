package user_data

import "sync"

type User struct {
	UserId        int64  `db:"user_id"`
	UserName      string `db:"user_name"`
	Password      string `db:"password"`
	HeroAvatar    string `db:"hero_avatar"`
	CurrHp        int32  `db:"curr_hp"`
	CreateTime    int64  `db:"create_time"`
	LastLoginTime int64  `db:"last_login_time"`
	MoveState     *MoveState

	componentMap *sync.Map // ConcurrentHashMap
	tempLocker   sync.Mutex
}

func (u *User) GetComponentMap() *sync.Map {
	if nil != u.componentMap {
		return u.componentMap
	}

	u.tempLocker.Lock()
	defer u.tempLocker.Unlock()

	if nil != u.componentMap {
		return u.componentMap
	}

	u.componentMap = &sync.Map{}

	return u.componentMap
}
