package cmd_handler

import (
	"fmt"
	"hero_story.go_server/biz_server/mod/user/user_data"
	"hero_story.go_server/biz_server/mod/user/user_lock"
	"hero_story.go_server/biz_server/mod/user/user_lso"
	"hero_story.go_server/biz_server/msg"
	"hero_story.go_server/biz_server/network/broadcaster"
	"hero_story.go_server/comm/lazy_save"
	"hero_story.go_server/comm/log"
)

//
// OnUserQuit 当用户退出游戏时执行
func OnUserQuit(userId int64) {
	if userId <= 0 {
		return
	}

	log.Info("用户离线, userId = %d", userId)

	//
	// 加锁
	//
	key := fmt.Sprintf("UserQuit_%d", userId)
	user_lock.TryLock(key)

	broadcaster.Broadcast(&msg.UserQuitResult{
		QuitUserId: uint32(userId),
	})

	user := user_data.GetUserGroup().GetByUserId(userId)

	if nil == user {
		return
	}

	userLso := user_lso.GetUserLso(user)
	lazy_save.Discard(userLso)

	log.Info("用户离线, 立即保存数据! userId = %d", userId)

	userLso.SaveOrUpdate(func() {
		//
		// 解锁
		//
		user_lock.UnLock(key)
	})
}
