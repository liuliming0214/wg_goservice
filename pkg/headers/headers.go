package headers

/**
 * 获取每一个请求的header签名
 * Create by :liming10
 * Created on: 2022/03/28 15:08
 */
import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"
	"wg_goservice/pkg/json"

	"github.com/gin-gonic/gin"
)

//请求之前  需要封装 相关的header信息
func GetHeader(c *gin.Context, appInfo map[string]interface{}) map[string]string {
	headerMap := make(map[string]string)
	Header := c.Request.Header
	for k, _ := range Header {
		headerMap[k] = c.Request.Header.Get(k)
	}
	var sys_app_secret string
	for sk, sv := range appInfo {
		if sk == "sys_app_id" {
			headerMap["Wg-Sys-App-Id"] = Strval(sv)
		}
		if sk == "method" {
			headerMap["Wg-Method"] = Strval(sv)
		}
		if sk == "sys_app_secret" {
			sys_app_secret = Strval(sv)
		}
		if sk == "api_url" {
			headerMap["Wg-Api-Url"] = Strval(sv)
		}
	}
	headerMap["method"] = Strval(appInfo["method"])
	newSign := GetSign(headerMap, sys_app_secret)
	//重新生成sign
	headerMap["Wg-Sign"] = Strval(newSign)
	return headerMap
}

/**liming10@leju.com
--检验请求的sign签名是否正确
--headerMap:传入的参数值组成的table，不含传入的签名值
--secret:加入到签名字符串中的混杂字符
*/
func GetSign(headerMap map[string]string, secret string) string {
	params := make(map[string]string)
	params["sys_app_id"] = headerMap["Wg-Sys-App-Id"]
	params["timestamp"] = headerMap["Wg-Timestamp"]
	params["app_id"] = headerMap["Wg-App-Id"]
	params["method"] = headerMap["Wg-Method"]
	params["x_request_id"] = headerMap["Wg-X-Request-Id"]
	// 排序
	var dataParams string
	//ksort
	var keys []string
	for k := range params {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	//拼接 提出所有的键名并按字符顺序排序
	for _, k := range keys {
		dataParams = dataParams + k + "=" + params[k] + "&"
	}
	dataParams = dataParams + secret

	h := md5.New()
	h.Write([]byte(strings.TrimRight(dataParams, "\n")))
	return hex.EncodeToString(h.Sum(nil))

}

// Strval 获取变量的字符串值
// 浮点型 3.0将会转换成字符串3, "3"
// 非数值或字符类型的变量将会被转换成JSON格式字符串
func Strval(value interface{}) string {
	// interface 转 string
	var key string
	if value == nil {
		return key
	}

	switch value.(type) {
	case float64:
		ft := value.(float64)
		key = strconv.FormatFloat(ft, 'f', -1, 64)
	case float32:
		ft := value.(float32)
		key = strconv.FormatFloat(float64(ft), 'f', -1, 64)
	case int:
		it := value.(int)
		key = strconv.Itoa(it)
	case uint:
		it := value.(uint)
		key = strconv.Itoa(int(it))
	case int8:
		it := value.(int8)
		key = strconv.Itoa(int(it))
	case uint8:
		it := value.(uint8)
		key = strconv.Itoa(int(it))
	case int16:
		it := value.(int16)
		key = strconv.Itoa(int(it))
	case uint16:
		it := value.(uint16)
		key = strconv.Itoa(int(it))
	case int32:
		it := value.(int32)
		key = strconv.Itoa(int(it))
	case uint32:
		it := value.(uint32)
		key = strconv.Itoa(int(it))
	case int64:
		it := value.(int64)
		key = strconv.FormatInt(it, 10)
	case uint64:
		it := value.(uint64)
		key = strconv.FormatUint(it, 10)
	case string:
		key = value.(string)
	case []byte:
		key = string(value.([]byte))
	default:
		newValue, _ := json.Marshal(value)
		key = string(newValue)
	}

	return key
}
