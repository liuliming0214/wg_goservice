package service

// Host 服务地址
type Host string

func (h Host) String() string {
	return string(h)
}

// Result 响应结构体(请继承该结构体)
type Result struct {
	Code int         `json:"code"` //提示代码
	Msg  string      `json:"msg"`  //提示信息
	Data interface{} `json:"data"` //出错
}

type ResultPmsService struct {
	ErrorCode int      `json:"errorCode"`
	Data      struct{} `json:"data"`
	Ret       int      `json:"ret"`
	Message   string   `json:"message"`
	Date      string   `json:"date"`
}
