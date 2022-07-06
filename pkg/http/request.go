package http

/**
 * http  使用工具类
 * Version   : 1.0
 * Create by :liming10
 * Created on: 2022/03/28 15:08
 */
import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"reflect"
	"time"

	"github.com/sohaha/zlsgo/ztype"
	"github.com/valyala/fasthttp"

	"wg_goservice/pkg/json"
	"wg_goservice/pkg/log"
)

var client = http.Client{
	Timeout: 3 * time.Second,
}

// Header http header config
type Header map[string]string

// Request HTTP请求
type Request struct {
	header  Header
	timeout int
	logMode uint8
}

const (
	LogInfo  uint8 = 1
	LogError uint8 = 2
)

// 请求日志
type reqLog struct {
	Method   string      `json:"method"`
	Url      string      `json:"url"`
	Header   Header      `json:"header"`
	ReqBody  interface{} `json:"req_body"`
	ResBody  interface{} `json:"res_body"`
	Duration string      `json:"duration"`
	Err      string      `json:"err"`
}

// NewRequest 创建HTTP请求实例
func NewRequest() *Request {
	return &Request{
		header: Header{
			"Content-Type": "application/x-www-form-urlencoded",
		},
		timeout: 3,
		logMode: LogInfo | LogError,
	}
}

// ResetHeader 重置Header
func (rq *Request) ResetHeader() *Request {
	rq.header = Header{}
	return rq
}

// SetHeader 设置Header
func (rq *Request) SetHeader(head Header) *Request {
	for k, v := range head {
		rq.header[k] = v
	}
	return rq
}

// LogMode 设置日志打印模式
// 0: 不打印日志
// 1: 仅打印请求成功日志
// 2: 仅打印请求失败日志
// 3: 所有日志均打印
func (rq *Request) LogMode(m uint8) *Request {
	rq.logMode = m
	return rq
}

// Post 发起POST请求
func (rq *Request) Post(url string, requestBody interface{}) ([]byte, int, error) {
	return rq.request(fasthttp.MethodPost, url, requestBody)
}

// Get 发起GET请求
func (rq *Request) Get(url string, requestBody interface{}) ([]byte, int, error) {
	return rq.request(fasthttp.MethodGet, url, requestBody)
}

// 使用标准包net/http
func (rq Request) request2(method string, url string, reqBody interface{}) (resBody []byte, statusCode int, err error) {
	// 请求体编码
	var reqBodyReader bytes.Buffer
	switch body := reqBody.(type) {
	case string:
		reqBodyReader.WriteString(body)
	case []byte:
		reqBodyReader.Write(body)
	default:
		var buf []byte
		buf, err = json.Marshal(reqBody)
		if err != nil {
			err = fmt.Errorf("JSON序列化失败: %v", err)
			return
		}
		reqBodyReader.Write(buf)
	}

	// 日志内容
	rl := reqLog{
		Method:   method,
		Url:      url,
		Header:   rq.header,
		ReqBody:  reqBody,
		ResBody:  "-",
		Duration: "-",
		Err:      "-",
	}

	// 构建请求
	req, err := http.NewRequest(method, url, &reqBodyReader)
	if err != nil {
		return
	}
	// 写入Header
	for k, v := range rq.header {
		req.Header.Set(k, v)
	}

	// 发送请求
	var startTime = time.Now().UnixNano()
	res, err := client.Do(req)
	rl.Duration = fmt.Sprintf("%vms", (time.Now().UnixNano()-startTime)/1e6) // 记录耗时
	if err != nil {
		if rq.logMode&LogError == LogError {
			rl.Err = err.Error()
			log.Error(rl)
		}
		return
	}

	// [必须]关闭资源
	defer func() {
		if res.Body != nil {
			_ = res.Body.Close()
		}
	}()

	// 读取响应数据
	statusCode = res.StatusCode
	if resBody, err = ioutil.ReadAll(res.Body); err != nil {
		err = fmt.Errorf("响应体读取失败, %w", err)
		rl.Err = err.Error()
		log.Error(rl)
		return
	}

	// 记录到日志
	_ = json.Unmarshal(resBody, &rl.ResBody)
	if rq.logMode&LogInfo == LogInfo {
		log.Info(rl)
	}

	return
}

// 基于fasthttp，但线上存在多个请求的响应数据交叉感染，导致JSON无法解析，根本原因暂时未知，本地没有复现
func (rq Request) request(method string, url string, reqBody interface{}) (respBody []byte, statusCode int, err error) {
	req := fasthttp.AcquireRequest()
	defer fasthttp.ReleaseRequest(req) // 用完需要释放资源

	// 设置请求头
	for k, v := range rq.header {
		req.Header.Set(k, v)
	}
	req.Header.SetMethod(method)
	req.SetRequestURI(url)
	var strs string
	//转化成字符串
	for sk, sv := range reqBody.(map[string]interface{}) {
		strs += ztype.ToString(sk) + "=" + ztype.ToString(sv) + "&"
	}
	// 请求体
	var requestBodyByte []byte
	switch reflect.ValueOf(reqBody).Kind() {
	case reflect.String:
		requestBodyByte = []byte(reflect.ValueOf(reqBody).String())
	case reflect.Slice:
		requestBodyByte = reflect.ValueOf(reqBody).Bytes()
	default:
		requestBodyByte, err = json.Marshal(reqBody)
		if err != nil {
			return respBody, statusCode, fmt.Errorf("JSON序列化失败: %v", err)
		}
	}
	fmt.Println(requestBodyByte, string(requestBodyByte))
	// 日志内容
	rl := reqLog{
		Method:   method,
		Url:      url,
		Header:   rq.header,
		ReqBody:  reqBody,
		ResBody:  "-",
		Duration: "-",
		Err:      "-",
	}
	// 发起请求
	req.SetBody([]byte(strs))
	resp := fasthttp.AcquireResponse()

	defer fasthttp.ReleaseResponse(resp) // 用完需要释放资源
	var startTime = time.Now().UnixNano()
	err = fasthttp.DoTimeout(req, resp, time.Second*time.Duration(rq.timeout))

	rl.Duration = fmt.Sprintf("%vms", (time.Now().UnixNano()-startTime)/1e6)
	if err != nil {
		if rq.logMode&LogError == LogError {
			rl.Err = err.Error()
			_ = json.Unmarshal(resp.Body(), &rl.ResBody)
			log.Info(rl)
		}

		return respBody, statusCode, err
	}

	// respBody = resp.Body()
	// 拷贝http响应的body
	respBody = make([]byte, len(resp.Body()))
	copy(respBody, resp.Body())

	statusCode = resp.StatusCode()

	_ = json.Unmarshal(resp.Body(), &rl.ResBody)

	if rq.logMode&LogInfo == LogInfo {
		log.Info(rl)
	}

	return respBody, statusCode, nil
}
