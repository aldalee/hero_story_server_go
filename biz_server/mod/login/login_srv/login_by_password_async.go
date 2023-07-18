package login_srv

import (
	"context"
	"fmt"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
	"hero_story.go_server/biz_server/base"
	"hero_story.go_server/biz_server/mod/user/user_dao"
	"hero_story.go_server/biz_server/mod/user/user_data"
	"hero_story.go_server/biz_server/mod/user/user_lock"
	"hero_story.go_server/comm/async_op"
	"hero_story.go_server/comm/cluster"
	"hero_story.go_server/comm/log"
)

// LoginByPasswordAsync 根据用户名称和密码进行登录,
// 将返回一个异步的业务结果
func LoginByPasswordAsync(userName string, password string) *base.AsyncBizResult {
	// 要说下面这两种写法有什么不同么?
	// func LoginByPasswordAsync(userName string, password string, callback func(user *user_data.User)) { ... }
	// func LoginByPasswordAsync(userName string, password string) *base.AsyncBizResult { ... }
	// 第一个是回调方式的, 第二个是返回 Future 方式的,
	// 这两个有什么不一样么?
	// 可以参考其他语言中的 async / await 相关知识...
	//
	if len(userName) <= 0 ||
		len(password) <= 0 {
		return nil
	}

	bizResult := &base.AsyncBizResult{}

	async_op.Process(
		async_op.StrToBindId(userName),
		func() {
			// 通过 DAO 获得用户数据
			user := user_dao.GetUserByName(userName)

			nowTime := time.Now().UnixMilli()

			if nil == user {
				// 如果用户数据为空，
				// 则新建数据...
				user = &user_data.User{
					UserName:   userName,
					Password:   password,
					CreateTime: nowTime,
					HeroAvatar: "Hero_Hammer",
				}
			}

			var etcdCli *clientv3.Client = cluster.GetEtcdCli()

			offlineLockName := fmt.Sprintf("user_offline_%d", user.UserId)
			getResp, _ := etcdCli.Get(
				context.TODO(), offlineLockName,
				clientv3.WithPrefix(),
			)

			if getResp.Count > 0 {
				// 这说明有下辖逻辑没有处理完,
				// 直接忽略本次登录...
				log.Error("下线逻辑未处理完成, 忽略本次登录! userName = %s", userName)
				bizResult.SetReturnedObj(nil)
				return
			}

			//
			// 是否有登出锁,
			// 如果有锁,
			// 那就直接退出吧...
			//
			key := fmt.Sprintf("UserQuit_%d", user.UserId)
			if user_lock.HasLock(key) {
				bizResult.SetReturnedObj(nil)
				return
			}

			// 更新最后登录时间
			user.LastLoginTime = nowTime
			user_dao.SaveOrUpdate(user)

			// 将用户添加到字典
			user_data.GetUserGroup().Add(user)

			bizResult.SetReturnedObj(user)
		},
		nil,
	)

	return bizResult
}
