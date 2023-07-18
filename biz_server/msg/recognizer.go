package msg

import (
	"errors"
	"google.golang.org/protobuf/reflect/protoreflect"
	"strings"
	"sync"
)

var msgCodeAndMsgDescMap = make(map[int16]protoreflect.MessageDescriptor)
var msgNameAndMsgCodeMap = make(map[string]int16)

var locker = &sync.Mutex{}

func getMsgDescByMsgCode(msgCode int16) (protoreflect.MessageDescriptor, error) {
	if msgCode < 0 {
		return nil, errors.New("消息代号无效")
	}

	if len(msgCodeAndMsgDescMap) <= 0 {
		init2Map()
	}

	return msgCodeAndMsgDescMap[msgCode], nil
}

func getMsgCodeByMsgName(msgName string) (int16, error) {
	if len(msgName) <= 0 {
		return -1, errors.New("消息名称为空")
	}

	if len(msgNameAndMsgCodeMap) <= 0 {
		init2Map()
	}

	msgName = strings.ToLower(
		strings.Replace(msgName, "_", "", -1),
	)

	return msgNameAndMsgCodeMap[msgName], nil
}

// 初始化两个字典
func init2Map() {
	locker.Lock()
	defer locker.Unlock()

	if len(msgCodeAndMsgDescMap) > 0 &&
		len(msgNameAndMsgCodeMap) > 0 {
		return
	}

	// 先往 msgNameAndMsgCodeMap "名称 --> 代号" 这个字典里填数据

	for k, v := range MsgCode_value {
		// USER_LOGIN_CMD ==> userlogincmd
		msgName := strings.ToLower(
			strings.Replace(k, "_", "", -1),
		)

		msgNameAndMsgCodeMap[msgName] = int16(v)
	}

	msgDescList := File_GameMsgProtocol_proto.Messages()

	for i := 0; i < msgDescList.Len(); i++ {
		msgDesc := msgDescList.Get(i)
		msgName := strings.ToLower(
			strings.Replace(string(msgDesc.Name()), "_", "", -1),
		)

		msgCode := msgNameAndMsgCodeMap[msgName]
		msgCodeAndMsgDescMap[msgCode] = msgDesc
	}
}
