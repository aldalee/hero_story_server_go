package base

type UserOfflineEvent struct {
	GatewayServerId int32
	SessionId       int32
	UserId          int64
}
