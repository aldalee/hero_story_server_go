package async_op

import "hash/crc32"

// StrToBindId 将字符串转换成绑定 Id
func StrToBindId(strVal string) int {
	v := int(crc32.ChecksumIEEE([]byte(strVal)))

	if v >= 0 {
		return v
	} else {
		return -v
	}
}
