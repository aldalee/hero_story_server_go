package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path"
	"strings"

	"github.com/gorilla/websocket"
	"hero_story.go_server/comm/cluster"
	"hero_story.go_server/comm/log"
	"hero_story.go_server/gateway_server/base"
	"hero_story.go_server/gateway_server/cluster/biz_server_finder"
	mywebsocket "hero_story.go_server/gateway_server/network/websocket"
)

var pServerId *int
var pBindHost *string
var pBindPort *int
var pEtcdEndpointArray *string

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var sessionId int32 = 0

// 启动网关服务器
func main() {
	fmt.Println("start gatewayServer")

	ex, err := os.Executable()

	if nil != err {
		panic(err)
	}

	log.Config(path.Dir(ex) + "/log/gateway_server.log")
	log.Info("Hello")

	pServerId = flag.Int("server_id", 0, "业务服务器 Id")
	pBindHost = flag.String("bind_host", "127.0.0.1", "绑定主机地址")
	pBindPort = flag.Int("bind_port", 54321, "绑定端口号")
	pEtcdEndpointArray = flag.String("etcd_endpoint_array", "127.0.0.1:2379", "Etcd 节点地址数组")
	flag.Parse() // 解析命令行参数

	ectdEndpointArray := strings.Split(*pEtcdEndpointArray, ",")
	cluster.InitEtcdCli(ectdEndpointArray)

	listenUserConnTransferCmd()
	biz_server_finder.FindNewBizServer()

	log.Info("启动网关服务器, serverId = %d, serverAddr = %s:%d", *pServerId, *pBindHost, *pBindPort)

	http.HandleFunc("/websocket", webSocketHandshake)
	_ = http.ListenAndServe(fmt.Sprintf("%s:%d", *pBindHost, *pBindPort), nil)
}

func webSocketHandshake(w http.ResponseWriter, r *http.Request) {
	if nil == w ||
		nil == r {
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)

	if nil != err {
		log.Error("WebSocket upgrade error, %v+", err)
		return
	}

	defer func() {
		_ = conn.Close()
	}()

	log.Info("有新客户端连入")

	sessionId += 1

	cmdCtx := &mywebsocket.CmdContextImpl{
		Conn:            conn,
		GatewayServerId: int32(*pServerId),
		SessionId:       sessionId,
	}

	base.GetCmdContextImplGroup().Add(cmdCtx.SessionId, cmdCtx)
	defer base.GetCmdContextImplGroup().RemoveBySessionId(cmdCtx.SessionId)

	cmdCtx.LoopSendMsg()
	cmdCtx.LoopReadMsg()
}

// 监听用户连接转移通知
func listenUserConnTransferCmd() {
	go func() {
		watchChan := cluster.GetEtcdCli().Watch(
			context.TODO(),
			"hero_story.go_server/publish/user_conn_transfer",
		)

		for resp := range watchChan {
			for _, event := range resp.Events {
				switch event.Type {
				case 0: // PUT
					jsonStr := string(event.Kv.Value)
					log.Info("收到玩家连接转移通知: " + jsonStr)

					userConnTransferCmd := &base.UserConnTransferCmd{}

					err := json.Unmarshal([]byte(jsonStr), userConnTransferCmd)

					if nil != err {
						log.Error("%+v", err)
						continue
					}

					userId := userConnTransferCmd.UserId
					cmdCtx := base.GetCmdContextImplGroup().GetByUserId(userId) // 没有这个函数

					if nil == cmdCtx {
						log.Error("未找到玩家上下文, userId = %d", userId)
						mywebsocket.UnbindUserId(userId, userConnTransferCmd.GatewayServerId) // 也没有这个函数
						continue
					}

					log.Info("令玩家断线, userId = %d", userId)
					cmdCtx.Disconnect()
				case 1: // DELETE
				}
			}
		}
	}()
}
