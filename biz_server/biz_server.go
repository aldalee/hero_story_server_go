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
	"time"

	"github.com/gorilla/websocket"
	clientv3 "go.etcd.io/etcd/client/v3"
	"hero_story.go_server/biz_server/base"
	mywebsocket "hero_story.go_server/biz_server/network/websocket"
	"hero_story.go_server/comm/cluster"
	"hero_story.go_server/comm/log"
)

var pServerId *int
var pBindHost *string
var pBindPort *int
var pSjtArray *string
var pEtcdEndpointArray *string

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  2048,
	WriteBufferSize: 2048,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var sessionId int32 = 0

// 启动业务服务器
func main() {
	fmt.Println("start bizServer")

	ex, err := os.Executable()

	if nil != err {
		panic(err)
	}

	log.Config(path.Dir(ex) + "/log/biz_server.log")
	log.Info("Hello")

	pServerId = flag.Int("server_id", 0, "业务服务器 Id")
	pBindHost = flag.String("bind_host", "127.0.0.1", "绑定主机地址")
	pBindPort = flag.Int("bind_port", 12345, "绑定端口号")
	pSjtArray = flag.String("server_job_type_array", "", "服务器职责类型数组")
	pEtcdEndpointArray = flag.String("etcd_endpoint_array", "127.0.0.1:2379", "Etcd 节点地址数组")
	flag.Parse() // 解析命令行参数

	base.G_serverId = *pServerId

	sjtArray := base.StringToServerJobTypeArray(*pSjtArray)
	etcdEndpointArray := strings.Split(*pEtcdEndpointArray, ",")

	cluster.InitEtcdCli(etcdEndpointArray)

	grantLeaseForServerLife()
	registerTheServer(*pServerId, *pBindHost, *pBindPort, sjtArray)

	//
	// 可以用 `ab -n 10000 -c 8 http://127.0.0.1/websocket` 来测试一下性能
	//
	//http.HandleFunc("/websocket", func(w http.ResponseWriter, r *http.Request) {
	//	_, _ = w.Write([]byte("Hello, the World!"))
	//})
	//
	log.Info(
		"启动业务服务器, serverId = %d, serverAddr = %s:%d, serverJobTypeArray = %s",
		*pServerId, *pBindHost, *pBindPort, *pSjtArray,
	)

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

	//ctx := &mywebsocket.CmdContextImpl{
	//	Conn:      conn,
	//	SessionId: sessionId,
	//}
	//
	//// 将指令上下文添加到分组,
	//// 当断开连接时移除指令上下文...
	//broadcaster.AddCmdCtx(sessionId, ctx)
	//defer broadcaster.RemoveCmdCtxBySessionId(sessionId)
	//
	//// 循环发送消息
	//ctx.LoopSendMsg()
	//// 开始循环读取消息
	//ctx.LoopReadMsg()

	myConn := &mywebsocket.GatewayServerConn{
		WsConn: conn,
	}

	myConn.LoopSendMsg()
	myConn.LoopReadMsg()
}

// 创建租约
func grantLeaseForServerLife() {
	leaseGrantResp, err := cluster.GetEtcdCli().Grant(context.TODO(), 5)

	if nil != err {
		log.Error("%+v", err)
		return
	}

	// 服务器不死, 就一直续约
	ch, err := cluster.GetEtcdCli().KeepAlive(
		context.TODO(),
		leaseGrantResp.ID,
	)

	if nil != err {
		log.Error("%+v", err)
		return
	}

	go func() {
		for {
			<-ch // ch ==> BlockingQueue.poll
		}
	}()

	base.G_leaseIdForServerLife = leaseGrantResp.ID
}

// 注册本服务器
func registerTheServer(serverId int, bindHost string, bindPort int, sjtArray []base.ServerJobType) {
	reportData := &base.BizServerData{
		ServerId:   int32(serverId),
		ServerAddr: fmt.Sprintf("%s:%d", bindHost, bindPort),
		SjtArray:   sjtArray,
	}

	go func() {
		grantResp, _ := cluster.GetEtcdCli().Grant(context.TODO(), 10)

		for {
			time.Sleep(5 * time.Second)

			// 更新负载数 ( 总人数 )
			reportData.LoadCount = base.GetLoadStat().GetTotalCount()

			leaseKeepLiveResp, _ := cluster.GetEtcdCli().KeepAliveOnce(context.TODO(), grantResp.ID)
			byteArray, _ := json.Marshal(reportData)

			_, _ = cluster.GetEtcdCli().Put(
				context.TODO(),
				fmt.Sprintf("hero_story.go_server/biz_server_%d", serverId), // hero_story.go_server/biz_server_1001
				string(byteArray),
				clientv3.WithLease(leaseKeepLiveResp.ID),
			)
		}
	}()

	go func() {
		watchChan := cluster.GetEtcdCli().Watch(
			context.TODO(),
			"hero_story.go_server/publish/user_offline",
		)

		for resp := range watchChan {
			for _, event := range resp.Events {
				switch event.Type {
				case 0: // PUT
					strVal := string(event.Kv.Value)
					log.Info("收到玩家下线通知: " + strVal)

					userOfflineEvent := &base.UserOfflineEvent{}
					_ = json.Unmarshal([]byte(strVal), userOfflineEvent)

					grantResp, _ := cluster.GetEtcdCli().Grant(context.TODO(), 50)
					offlineLockName := fmt.Sprintf(
						"user_offline_%d_%d",
						userOfflineEvent.UserId,
						*pServerId,
					)

					cluster.GetEtcdCli().Put(
						context.TODO(), offlineLockName, "1",
						clientv3.WithLease(grantResp.ID),
					)

					base.GetLoadStat().DeleteUserId(
						userOfflineEvent.GatewayServerId,
						userOfflineEvent.UserId,
					)

					cluster.GetEtcdCli().Revoke(context.TODO(), grantResp.ID)
				case 1: // DELETE
				}
			}
		}
	}()
}
