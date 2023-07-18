package biz_server_finder

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	bizsrvbase "hero_story.go_server/biz_server/base"
	"hero_story.go_server/biz_server/msg"
	"hero_story.go_server/comm/cluster"
	"hero_story.go_server/comm/log"
	"hero_story.go_server/gateway_server/base"
)

var connectedBizServerMap = &sync.Map{}

// 连接到业务服务器
func connToBizServer(bizServerData *bizsrvbase.BizServerData) {
	if nil == bizServerData ||
		bizServerData.ServerId <= 0 ||
		len(bizServerData.ServerAddr) <= 0 ||
		nil == bizServerData.SjtArray ||
		len(bizServerData.SjtArray) <= 0 {
		return
	}

	bizServerId := bizServerData.ServerId
	_, ok := connectedBizServerMap.Load(bizServerId)

	if ok {
		return
	}

	// 创建到游戏服务器的连接
	newConn, _, err := websocket.DefaultDialer.Dial(fmt.Sprintf("ws://%s/websocket", bizServerData.ServerAddr), nil)

	if nil != err {
		log.Error("%+v", err)
		return
	}

	log.Info("已经连接到业务服务器, %s", bizServerData.ServerAddr)
	connectedBizServerMap.Store(bizServerId, 1)

	//
	// 循环读取游戏服发来的消息,
	// 转发给客户端
	//
	go func() {
		newInstance := &BizServerInstance{
			bizServerData,
			newConn,
		}

		addBizServerInstance(newInstance)
		defer deleteBizServerInstance(newInstance)
		defer connectedBizServerMap.Delete(bizServerId)

		for {
			// 读取从游戏服返回来的消息
			_, msgData, err := newConn.ReadMessage()

			if nil != err {
				log.Error("%+v", err)
			}

			innerMsg := &msg.InternalServerMsg{}
			innerMsg.FromByteArray(msgData)

			log.Info("从业务服务器返回结果, sessionId = %d, userId = %d", innerMsg.SessionId, innerMsg.UserId)

			// 这个是客户端到网关服务器的上下文对象,
			// 通过它来发送消息给客户端
			cmdCtx := base.GetCmdContextImplGroup().GetBySessionId(innerMsg.SessionId)

			if nil == cmdCtx {
				log.Error("未找到指令上下文")
				continue
			}

			if 0 != innerMsg.Disconnect {
				log.Info(
					"服务器强制断开玩家连接, sessionId = %d, userId = %d",
					innerMsg.SessionId,
					innerMsg.UserId,
				)
				cmdCtx.Disconnect()
				return
			}

			// 拦截一下消息, 做点手脚
			msgCode := binary.BigEndian.Uint16(innerMsg.MsgData[2:4]) // [2, 4) ==> [2, 3] ==> 3rd, 4th

			if uint16(msg.MsgCode_USER_LOGIN_RESULT) == msgCode &&
				innerMsg.UserId > 0 {
				// 如果是登录结果消息, msgCode == 14,
				// 并且有用户 Id,
				// 则说明登录成功...
				go handleLoginResult(cmdCtx, innerMsg)
			} else {
				// 不是登录消息,
				// 或则是登录失败了
				cmdCtx.Write(innerMsg.MsgData)
			}

		}
	}()
}

func handleLoginResult(cmdCtx base.ClientCmdContext, innerMsg *msg.InternalServerMsg) {
	// 如果写入 Etcd 成功, 才执行绑定用户 Id 的操作
	if nil == cmdCtx ||
		nil == innerMsg {
		return
	}

	etcdKey := fmt.Sprintf("user_%d_at_gateway_server_id", innerMsg.UserId)

	getResp, err := cluster.GetEtcdCli().Get(
		context.TODO(),
		etcdKey,
	)

	if nil != err {
		cmdCtx.Disconnect()
		return
	}

	if len(getResp.Kvs) <= 0 {
		// 如果没有其他网关服务器拿着该用户,
		// 直接登录成功
		onLoginSuccess(cmdCtx, innerMsg)
		return
	}

	// 拿到原来的网关服务器 Id
	oldGatewayServerId, _ := strconv.Atoi(string(getResp.Kvs[0].Value))

	userConnTransferCmd := &base.UserConnTransferCmd{
		UserId:          innerMsg.UserId,
		GatewayServerId: int32(oldGatewayServerId),
	}

	byteArray, _ := json.Marshal(userConnTransferCmd)
	jsonStr := string(byteArray)

	cluster.GetEtcdCli().Put(
		context.TODO(),
		"hero_story.go_server/publish/user_conn_transfer",
		jsonStr,
	)

	// 5 秒钟超时
	timeoutCtx, cancelFunc := context.WithTimeout(context.TODO(), time.Second*5)
	defer cancelFunc()

	watchChan := cluster.GetEtcdCli().Watch(timeoutCtx, etcdKey)
	// watchChan ==> BlockingQueue

	select {
	case resp := <-watchChan: // BlockingQueue.poll
		for _, event := range resp.Events {
			switch event.Type {
			case 1: // DELETE
				onLoginSuccess(cmdCtx, innerMsg)
			}
		}

	case <-timeoutCtx.Done(): // BlockingQueue.poll
		log.Error("等待连接转移答复超时! 关闭玩家连接. userId = %d", innerMsg.UserId)
		cmdCtx.Disconnect()
	}
}

func onLoginSuccess(cmdCtx base.ClientCmdContext, innerMsg *msg.InternalServerMsg) {
	if nil == cmdCtx ||
		nil == innerMsg {
		return
	}

	if cmdCtx.GetUserId() <= 0 &&
		innerMsg.UserId > 0 {
		cmdCtx.BindUserId(innerMsg.UserId)
	}

	cmdCtx.Write(innerMsg.MsgData)
}
