package json

import (
	"github.com/json-iterator/go"
	"github.com/json-iterator/go/extra"
)

var json jsoniter.API

func init() {
	// [模糊模式]
	// 1. 兼容空数组作为对象，比如PHP中空数组序列化后为[]，对此非对象将兼容
	// 2. 兼容字符串与数字类型，比如字段定义为int，而JSON里为字符串格式
	extra.RegisterFuzzyDecoders()

	json = jsoniter.Config{
		EscapeHTML:             true,
		SortMapKeys:            true,
		ValidateJsonRawMessage: true,
	}.Froze()
}

// API 返回JSON实例
func API() jsoniter.API {
	return json
}

// MarshalToString 序列化成JSON字符串
func MarshalToString(v interface{}) (string, error) {
	return json.MarshalToString(&v)
}

// Marshal 序列化成字节切片
func Marshal(v interface{}) ([]byte, error) {
	return json.Marshal(&v)
}

// UnmarshalFromString 从字符串反序列化
func UnmarshalFromString(str string, v interface{}) error {
	return json.UnmarshalFromString(str, &v)
}

// Unmarshal 从字节切片反序列化
func Unmarshal(data []byte, v interface{}) error {
	return json.Unmarshal(data, &v)
}

// UnmarshalInterface 从结构体的interface{}到interface{}的反序列化
func UnmarshalInterface(data interface{}, v interface{}) error {
	j, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return json.Unmarshal(j, &v)
}

// Valid 验证字节切片是否为标准的JSON
func Valid(data []byte) bool {
	return json.Valid(data)
}

// Get 根据PATH查询JSON字段
func Get(data []byte, path ...interface{}) jsoniter.Any {
	return json.Get(data, path)
}
