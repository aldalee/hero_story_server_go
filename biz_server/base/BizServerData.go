package base

type BizServerData struct {
	ServerId   int32           `json:"serverId"`
	ServerAddr string          `json:"serverAddr"`
	SjtArray   []ServerJobType `json:"sjtArray"`
	LoadCount  int32           `json:"loadCount"`
}
