package base

import (
	"strconv"
	"strings"
)

// ServerJobType 服务器职责类型
type ServerJobType int32

const (
	ServerJobTypeLOGIN ServerJobType = iota + 1 // 1
	ServerJobTypeGAME                           // 2
)

func (sjt ServerJobType) ToString() string {
	switch sjt {
	case ServerJobTypeLOGIN:
		return "LOGIN"
	case ServerJobTypeGAME:
		return "GAME"
	default:
		return ""
	}
}

func StringToServerJobType(strVal string) ServerJobType {
	if strings.EqualFold(strVal, "LOGIN") || // 和 "LOGIN" 比较
		strings.EqualFold(strVal, strconv.Itoa(int(ServerJobTypeLOGIN))) { // 和 "1" 比较
		return ServerJobTypeLOGIN
	}

	if strings.EqualFold(strVal, "GAME") || // 和 "GAME" 比较
		strings.EqualFold(strVal, strconv.Itoa(int(ServerJobTypeGAME))) { // 和 "2" 比较
		return ServerJobTypeGAME
	}

	panic("无法转型为服务器职责类型")
}

func StringToServerJobTypeArray(strVal string) []ServerJobType {
	if len(strVal) <= 0 {
		return nil
	}

	strArray := strings.Split(strVal, ",")
	var enumArray []ServerJobType

	for _, currStr := range strArray {
		sjt := StringToServerJobType(currStr)
		enumArray = append(enumArray, sjt)
	}

	return enumArray
}
