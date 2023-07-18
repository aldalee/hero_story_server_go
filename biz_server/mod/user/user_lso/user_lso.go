package user_lso

import (
	"fmt"
	"hero_story.go_server/biz_server/mod/user/user_dao"
	"hero_story.go_server/biz_server/mod/user/user_data"
	"hero_story.go_server/comm/async_op"
)

type UserLso struct {
	*user_data.User
}

func (lso *UserLso) GetLsoId() string {
	return fmt.Sprintf("UserLso_%d", lso.UserId)
}

func (lso *UserLso) SaveOrUpdate(callback func()) {
	async_op.Process(
		int(lso.UserId),
		func() {
			user_dao.SaveOrUpdate(lso.User)

			if nil != callback {
				callback()
			}
		},
		nil,
	)
}
