package user_data

type userGroup struct {
	innerMap map[int64]*User // Map<long, User>
}

var userGroupInstance = &userGroup{
	innerMap: make(map[int64]*User), // Map<Long, User>
}

// GetUserGroup 获取用户组
func GetUserGroup() *userGroup {
	return userGroupInstance
}

// Add 添加用户到字典
func (group *userGroup) Add(user *User) {
	if nil == user {
		return
	}

	group.innerMap[user.UserId] = user
}

// RemoveByUserId 删除用户
func (group *userGroup) RemoveByUserId(userId int64) {
	if userId <= 0 {
		return
	}

	delete(group.innerMap, userId)
}

// GetByUserId 根据用户 Id 获取用户数据
func (group *userGroup) GetByUserId(userId int64) *User {
	if userId <= 0 {
		return nil
	}

	return group.innerMap[userId]
}

// GetUserALL 获得所有用户
func (group *userGroup) GetUserALL() map[int64]*User {
	return group.innerMap
}
