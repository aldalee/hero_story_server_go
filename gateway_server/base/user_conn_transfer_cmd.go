package base

type UserConnTransferCmd struct {
	UserId          int64 `json:"userId"`
	GatewayServerId int32 `json:"gatewayServerId"`
}
