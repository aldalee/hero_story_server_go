package user_lso

import "hero_story.go_server/biz_server/mod/user/user_data"

func GetUserLso(user *user_data.User) *UserLso {
	if nil == user {
		return nil
	}

	existComp, _ := user.GetComponentMap().Load("UserLso") // map.get("UserLso")

	if nil != existComp {
		return existComp.(*UserLso)
	}

	existComp = &UserLso{
		User: user,
	}

	existComp, _ = user.GetComponentMap().LoadOrStore("UserLso", existComp)

	return existComp.(*UserLso)
}
