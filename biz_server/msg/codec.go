package msg

import (
	"encoding/binary"
	"errors"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

// Decode 根据消息代号解码字节数组
func Decode(msgData []byte, msgCode int16) (*dynamicpb.Message, error) {
	if nil == msgData {
		return nil, errors.New("消息数据为空")
	}

	msgDesc, err := getMsgDescByMsgCode(msgCode)

	if nil != err {
		return nil, err
	}

	newMsg := dynamicpb.NewMessage(msgDesc)

	if err := proto.Unmarshal(msgData, newMsg); nil != err {
		return nil, err
	}

	return newMsg, nil
}

// Encode 将消息对象编码成字节数组
func Encode(msgObj protoreflect.ProtoMessage) ([]byte, error) {
	if nil == msgObj {
		return nil, errors.New("消息对象为空")
	}

	msgCode, err := getMsgCodeByMsgName(string(msgObj.ProtoReflect().Descriptor().Name()))

	if nil != err {
		return nil, err
	}

	msgSizeByteArray := make([]byte, 2)
	binary.BigEndian.PutUint16(msgSizeByteArray, 0)

	msgCodeByteArray := make([]byte, 2)
	binary.BigEndian.PutUint16(msgCodeByteArray, uint16(msgCode))

	msgBodyArray, err := proto.Marshal(msgObj)

	if nil != err {
		return nil, err
	}

	completeMsg := append(msgSizeByteArray, msgCodeByteArray...)
	completeMsg = append(completeMsg, msgBodyArray...)

	return completeMsg, nil
}
