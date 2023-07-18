package base

import clientv3 "go.etcd.io/etcd/client/v3"

// 服务器 Id
var G_serverId int

// 租约 Id, 服务器一直活着, 这个租约就一直有效
var G_leaseIdForServerLife clientv3.LeaseID
