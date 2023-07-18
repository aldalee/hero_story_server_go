package websocket

import (
	"context"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	clientv3 "go.etcd.io/etcd/client/v3"
	"hero_story.go_server/biz_server/base"
	"hero_story.go_server/biz_server/msg"
	"hero_story.go_server/comm/cluster"
	"hero_story.go_server/comm/log"
	"hero_story.go_server/gateway_server/cluster/biz_server_finder"
)

const oneSecond = 1000
const readMsgCountPerSecond = 16

// CmdContextImpl 就是 MyCmdContext 的 WebSocket 实现
type CmdContextImpl struct {
	userId          int64
	clientIpAddr    string
	closeFlag       bool
	Conn            *websocket.Conn
	sendMsgQ        chan []byte // BlockingQueue
	GatewayServerId int32
	SessionId       int32
	GameServerId    int32
}

func (ctx *CmdContextImpl) GetUserId() int64 {
	return ctx.userId
}

func (ctx *CmdContextImpl) GetClientIpAddr() string {
	return ctx.clientIpAddr
}

func (ctx *CmdContextImpl) Write(byteArray []byte) {
	if nil == byteArray ||
		nil == ctx.Conn ||
		nil == ctx.sendMsgQ {
		return
	}

	ctx.sendMsgQ <- byteArray // queue.push
}

func (ctx *CmdContextImpl) SendError(errorCode int, errorInfo string) {
}

func (ctx *CmdContextImpl) Disconnect() {
	if nil != ctx.Conn {
		_ = ctx.Conn.Close()
	}
}

// LoopSendMsg 循环发送消息,
// 内部通过协程来实现
func (ctx *CmdContextImpl) LoopSendMsg() {
	// 首先构建发送队列
	ctx.sendMsgQ = make(chan []byte, 64)

	go func() { // new Thread().start(() -> { ... })
		for {
			byteArray := <-ctx.sendMsgQ // queue.pop

			if nil == byteArray {
				continue
			}

			func() {
				defer func() {
					if err := recover(); nil != err {
						log.Error("发生异常, %+v", err)
					}
				}()

				if err := ctx.Conn.WriteMessage(websocket.BinaryMessage, byteArray); nil != err {
					log.Error("%+v", err)
				}
			}()
		}
	}() // 相当于启动一个线程, 专门负责发送消息
}

// LoopReadMsg 循环读取消息
func (ctx *CmdContextImpl) LoopReadMsg() {
	if nil == ctx.Conn {
		return
	}

	// 设置读取字节数限制
	ctx.Conn.SetReadLimit(64 * 1024)

	t0 := int64(0)
	counter := 0

	//
	// 循环读取游戏客户端发来的消息,
	// 转发给游戏服
	//
	for {
		// 接收游戏客户端发来的消息
		msgType, msgData, err := ctx.Conn.ReadMessage()

		if nil != err {
			log.Error("%+v", err)
			break
		}

		t1 := time.Now().UnixMilli()

		if (t1 - t0) > oneSecond {
			t0 = t1
			counter = 0
		}

		if counter >= readMsgCountPerSecond {
			log.Error("消息过于频繁")
			continue
		}

		counter++

		msgCode := binary.BigEndian.Uint16(msgData[2:4]) // [2, 4) ==> 2, 3 ==> 3rd, 4th
		sjt := biz_server_finder.GetServerJobTypeByMsgCode(msgCode)

		// 创建到游戏服务器的连接
		bizServerConn, err := biz_server_finder.GetBizServerConn(
			sjt, getBizServerFindStrategy(ctx, sjt),
		)

		// 如果 null == bizServerConn,
		// 我们就需要调用第二个策略...

		if nil != err {
			log.Error("%+v", err)
			continue
		}

		func() {
			defer func() {
				if err := recover(); nil != err {
					log.Error("发生异常, %+v", err)
				}
			}()

			log.Info("收到客户端消息并转发")

			innerMsg := &msg.InternalServerMsg{
				GatewayServerId: ctx.GatewayServerId,
				SessionId:       ctx.SessionId,
				UserId:          ctx.GetUserId(),
				MsgData:         msgData,
			}

			innerMsgByteArray := innerMsg.ToByteArray()

			// 将客户端消息转发给游戏服
			if err = bizServerConn.WriteMessage(msgType, innerMsgByteArray); nil != err {
				log.Error("转发消息失败, %+v", err)
			}
		}()
	}

	ctx.closeFlag = true

	// innerMsg := &msg.InternalServerMsg{
	// 	GatewayServerId: 0,
	// 	SessionId:       ctx.SessionId,
	// 	UserId:          ctx.GetUserId(),
	// 	Disconnect:      1,
	// }

	// innerMsgByteArray := innerMsg.ToByteArray()

	// _ = bizServerConn.WriteMessage(websocket.BinaryMessage, innerMsgByteArray)

	if ctx.userId <= 0 {
		// 如果不知道用户是谁?
		// 也就是用户没有登录成功,
		// 直接退出!
		return
	}

	grantResp, _ := cluster.GetEtcdCli().Grant(context.TODO(), 5)
	offlineLockName := fmt.Sprintf(
		"user_offline_%d_%d",
		ctx.userId,
		ctx.GatewayServerId,
	)
	cluster.GetEtcdCli().Put(
		context.TODO(), offlineLockName, "1",
		clientv3.WithLease(grantResp.ID),
	)

	userOfflineEvent := &base.UserOfflineEvent{
		GatewayServerId: ctx.GatewayServerId,
		SessionId:       ctx.SessionId,
		UserId:          ctx.userId,
	}

	byteArray, _ := json.Marshal(userOfflineEvent)

	cluster.GetEtcdCli().Put(
		context.TODO(),
		"hero_story.go_server/publish/user_offline", // 频道、主题 Topic
		string(byteArray),
	)

	//cluster.GetEtcdCli().Revoke(context.TODO(), grantResp.ID)
	ctx.UnbindUserId()
}

// 根据服务器职责类型返回服务器的选择策略
func getBizServerFindStrategy(ctx *CmdContextImpl, sjt base.ServerJobType) biz_server_finder.FindStrategy {
	if nil == ctx {
		return nil
	}

	if base.ServerJobTypeLOGIN == sjt {
		return &biz_server_finder.RandomFindStrategy{}
	}

	if base.ServerJobTypeGAME == sjt {
		return &biz_server_finder.CompositeFindStrategy{
			Finder1: &biz_server_finder.IdFindStrategy{
				ServerId: ctx.GameServerId, // Read
			},
			Finder2: &biz_server_finder.EtcdFindStrategy{
				EtcdKey:       fmt.Sprintf("user_%d_at_game_server_id", ctx.userId),
				WriteServerId: &ctx.GameServerId,
			},
			Finder3: &biz_server_finder.LeastLoadFindStrategy{
				WriteServerId: &ctx.GameServerId,
			},
		}
	}

	return nil
}
