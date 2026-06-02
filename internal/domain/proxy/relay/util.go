// 通用工具函数
//
// 本文件提供 relay 包使用的通用工具函数：
//   - jsonMarshal: JSON 序列化封装
//   - jsonUnmarshal: JSON 反序列化封装
package relay

import "encoding/json"

// jsonMarshal JSON 序列化封装
func jsonMarshal(v any) ([]byte, error) {
	return json.Marshal(v)
}

// jsonUnmarshal JSON 反序列化封装
func jsonUnmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}
