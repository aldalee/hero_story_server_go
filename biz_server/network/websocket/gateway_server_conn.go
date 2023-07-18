package websocket

import (
	"encoding/binary"
	"fmt"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"hero_story.go_server/biz_server/base"
	"hero_story.go_server/biz_server/cmd_handler"
	"hero_story.go_server/biz_server/msg"
	"hero_story.go_server/biz_server/network/broadcaster"
	"hero_story.go_server/comm/log"
	"hero_story.go_server/comm/main_thread"
)

type GatewayServerConn struct {
	GatewayServerId     int32
	WsConn              *websocket.Conn
	sendMsgQ            chan *msg.InternalServerMsg // BlockingQueue
	ctxMap              *sync.Map
	ctxMapLastClearTime int64
}

func (conn *GatewayServerConn) LoopSendMsg() {
	conn.sendMsgQ = make(chan *msg.InternalServerMsg, 64)

	go func() {
		for {
			msgObj := <-conn.sendMsgQ

			if nil == msgObj {
				continue
			}

			byteArray := msgObj.ToByteArray()

			if err := conn.WsConn.WriteMessage(websocket.BinaryMessage, byteArray); nil != err {
				log.Error("%+v", err)
			}
		}
	}()
}

func (conn *GatewayServerConn) LoopReadMsg() {
	if nil == conn.WsConn {
		return
	}

	conn.ctxMap = &sync.Map{}

	for {
		_, msgData, err := conn.WsConn.ReadMessage()

		if nil != err {
			log.Error("%+v", err)
			base.GetLoadStat().DeleteByGatewayServerId(conn.GatewayServerId)
			break
		}

		func() {
			defer func() {
				if err := recover(); nil != err {
					log.Error("发生异常, %+v", err)
				}
			}()

			// 网关服务器发来的消息， 可以看成是带包装的消息
			innerMsg := &msg.InternalServerMsg{}
			//innerMsg.UserId
			//innerMsg.GatewayServerId
			//innerMsg.SessionId

			innerMsg.FromByteArray(msgData)

			if 0 != innerMsg.Disconnect {
				cmd_handler.OnUserQuit(innerMsg.UserId)
				base.GetLoadStat().DeleteUserId(innerMsg.GatewayServerId, innerMsg.UserId)
				return
			}

			if conn.GatewayServerId <= 0 {
				conn.GatewayServerId = innerMsg.GatewayServerId
			}

			realMsgData := innerMsg.MsgData // 拆掉包装, 拿到真实消息 == 用户消息

			msgCode := binary.BigEndian.Uint16(realMsgData[2:4])
			newMsgX, err := msg.Decode(realMsgData[4:], int16(msgCode))

			if nil != err {
				log.Error(
					"消息解码错误, msgCode = %d, error = %+v",
					msgCode, err,
				)
				return
			}

			log.Info(
				"收到客户端消息, remoteSessionId = %d, userId = %d, msgCode = %d, msgName = %s",
				innerMsg.SessionId,
				innerMsg.UserId,
				msgCode,
				newMsgX.Descriptor().Name(),
			)

			// 创建指令处理器
			cmdHandler := cmd_handler.CreateCmdHandler(msgCode)

			if nil == cmdHandler {
				log.Error(
					"未找到指令处理器, msgCode = %d",
					msgCode,
				)
				return
			}

			// 获取唯一的会话 Id
			sessionUId := fmt.Sprintf("%d_%d", innerMsg.GatewayServerId, innerMsg.SessionId)

			mapVal, _ := conn.ctxMap.Load(sessionUId)

			if nil == mapVal {
				newCtx := &innerCmdContextImpl{
					gatewayServerId:   innerMsg.GatewayServerId,
					remoteSessionId:   innerMsg.SessionId,
					userId:            innerMsg.UserId,
					GatewayServerConn: conn,
				}

				log.Info("新建 Ctx, sessionUId = %s", sessionUId)
				mapVal, _ = conn.ctxMap.LoadOrStore(sessionUId, newCtx)
			}

			if nil == mapVal {
				log.Error("Ctx 依然为空")
				return
			}

			currCtx := mapVal.(*innerCmdContextImpl)
			currCtx.lastActiveTime = time.Now().UnixMilli()
			main_thread.Process(func() {
				cmdHandler(currCtx, newMsgX)
			})

			broadcaster.AddCmdCtx(sessionUId, currCtx)
			base.GetLoadStat().AddUserId(innerMsg.GatewayServerId, innerMsg.UserId)

			// 在这里判断 ctxMap 里有没有长时间没有发送消息的用户，
			// 如果有，就删除掉
			conn.clearCtxMap()
		}()
	}
}

func (conn *GatewayServerConn) clearCtxMap() {
	nowTime := time.Now().UnixMilli()

	if nowTime-conn.ctxMapLastClearTime < 120000 {
		// 如果上次清除时间距离当前时间还没有超过 2 分钟,
		// 不执行任何操作
		return
	}

	conn.ctxMapLastClearTime = nowTime

	deleteSessionUIdArray := make([]interface{}, 64)

	conn.ctxMap.Range(func(key interface{}, val interface{}) bool {
		if nil == key ||
			nil == val {
			return true
		}

		currCtx := val.(*innerCmdContextImpl)

		if nowTime-currCtx.lastActiveTime < 120000 {
			// 如果 currCtx 上次活动时间是在 2 分钟之内,
			// 保留, 不做处理
			return true
		}

		deleteSessionUIdArray = append(deleteSessionUIdArray, key)
		return true
	})

	for _, sessionUId := range deleteSessionUIdArray {
		if nil == sessionUId {
			continue
		}

		log.Info("删除 Ctx, sessionUId = %+v", sessionUId)
		conn.ctxMap.Delete(sessionUId)
		broadcaster.RemoveCmdCtxBySessionId(sessionUId.(string))
	}
}
