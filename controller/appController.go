package controller

/*
*一定要注意
*协程的开启和关闭
*channel的开启和关闭
*免得协程泄漏，内存泄漏
*还有就是数据，最好用结构体，指针，别用map，并发读写不安全 这也是常踩到坑
*用map 一定注意独立，不能并发读写
*注意copy
*map 使用前要判断 单元是否存在
*不要通过共享内存来通信，而通过通信来共享内存。
*
 */
import (
	"fmt"
	"sync"
	"wg_goservice/global"
	"wg_goservice/pkg/headers"
	"wg_goservice/pkg/http"
	"wg_goservice/pkg/json"
	"wg_goservice/pkg/result"
	"wg_goservice/pkg/service"
	"wg_goservice/rediscache"

	"github.com/gin-gonic/gin"
	"github.com/valyala/fasthttp"
)

type AppController struct{}

func NewAppController() AppController {
	return AppController{}
}

//这里使用了sync.WaitGroup来实现goroutine的同步
var wg sync.WaitGroup

//所有的请方式操纵
func (a *AppController) RequestInfo(c *gin.Context) {
	//获取get的请求数据
	fmt.Println(c.Request.URL.Query())
	c.Request.ParseMultipartForm(128) //保存表单缓存的内存大小128M
	data := c.Request.Form
	//定义一个map数据
	reqData := make(map[string]interface{})
	//处理post数据到map
	for key, value := range data {
		reqData[key] = value[0]
	}
	//处理get数据到map
	for key, value := range c.Request.URL.Query() {
		reqData[key] = value[0]
	}
	result := result.NewResult(c)
	//获取头部信息
	//获取标识
	WgMethod := c.Request.Header.Get("Wg-Method")
	if WgMethod == "" {
		result.Error(1404, "Wg-Method字段缺失")
		return
	}

	fmt.Println(global.LogsSetting.LogPath)
	//根据标识 获取配置信息
	layout, err := rediscache.GetLayoutRedisCache(WgMethod)
	if err != nil {
		result.Error(1405, "数据查询错误")
		return
	}
	//定义一个map数据
	var returnmap map[string]interface{}
	returnmap = make(map[string]interface{})

	//定义一个map数据 用来判断提交 可能会内存溢出 记得GC内存释放
	var conmap map[string]interface{}
	conmap = make(map[string]interface{})
	bodyMap := make(map[string]interface{})
	//创建通道
	var ch chan interface{}
	var ch2 chan interface{}
	//有缓冲的通道  防止阻塞 如果不给1个缓存位  程序不能执行 直接阻塞
	ch = make(chan interface{}, 1)
	ch2 = make(chan interface{}, 1)
	//定义map
	bs := make(map[string]string)
	ch <- reqData
	//获取map数据 获取每一个step的数据
	for lk, lv := range layout {
		//lv  转化成map
		//如果大于1  开启go 的协程
		if len(lv.(map[string]interface{})) > 1 {
			reqBody := <-ch
		OuterLoop:
			for sk, sv := range lv.(map[string]interface{}) {
				svinfo := make(map[string]interface{})
				json.UnmarshalInterface(sv, &svinfo)
				fmt.Println("*************************************")
				//获取配置
				NewMethod, _ := rediscache.GetNewMethodRedisCache(sk)

				//开始处理相关数据 获取单个借口具体配置
				wg.Add(1) //启用一个gorouteine  就+1
				headerMap := headers.GetHeader(c, NewMethod)
				//获取将要请求的参数信息
				params := service.GetParams(svinfo, conmap, headerMap, reqBody, reqData, bodyMap)
				//条件判断 如果满足
				conditions := service.GetCondition(svinfo, conmap, reqBody, reqData)

				if conditions == false {
					//这里要注意  跳出循环 结束协程 防止阻塞
					wg.Done()
					//跳出OuterLoop 循环
					continue OuterLoop
				}
				go GetData(headerMap, params, ch2)
				//拿到协程的值 后 合并处理 赋值nretmap
				nretmap := <-ch2
				retmap := make(map[string]interface{})
				err = json.UnmarshalInterface(nretmap, &retmap)
				returnmap[sk] = retmap["data"]
				//用来条件判断
				conmap = retmap
				bodyMap[lk+"."+sk] = retmap
				retmap2 := make(map[string]interface{})
				err = json.UnmarshalInterface(retmap["data"], &retmap2)
				for msk, msv := range retmap2 {
					bs[msk] = headers.Strval(msv)
				}
			}
			//map 转成 interface
			var newface map[string]interface{}
			newface = make(map[string]interface{})
			for bsk, bsv := range bs {
				newface[bsk] = bsv
			}
			//合并协程字段  发送给管道
			ch <- newface
			wg.Wait() // 等待所有登记的goroutine都结束
		} else {
		OuterLoopFor:
			for sk, sv := range lv.(map[string]interface{}) {
				//获取配置
				NewMethod, _ := rediscache.GetNewMethodRedisCache(sk)
				//开始处理相关数据 获取单个借口具体配置
				headerMap := headers.GetHeader(c, NewMethod)
				svinfo := make(map[string]interface{})
				json.UnmarshalInterface(sv, &svinfo)
				//定义相关变量  防止在判断里面使用:= 出现  resp declared but not used
				var resp []byte
				var status int
				var err error
				reqBody := <-ch // 从ch中接收值并赋值给变量
				params := service.GetParams(svinfo, conmap, headerMap, reqBody, reqData, bodyMap)
				conditions := service.GetCondition(svinfo, conmap, reqBody, reqData)
				if conditions == false {
					//跳出OuterLoop 循环
					continue OuterLoopFor
				}
				if headerMap["method"] == "POST" {
					resp, status, err = http.NewRequest().SetHeader(headerMap).Post(headerMap["Wg-Api-Url"], params)
				} else {
					resp, status, err = http.NewRequest().SetHeader(headerMap).Get(headerMap["Wg-Api-Url"], params)
				}
				if status != fasthttp.StatusOK || err != nil {
					err = fmt.Errorf("接口请求失败，HTTP状态码: %d", status)
					return
				}
				var datas service.Result
				json.Unmarshal(resp, &datas)
				retmap := make(map[string]interface{})
				err = json.UnmarshalInterface(datas, &retmap)
				//收集返回的数==0-9654321`=-09875=
				conmap = retmap
				returnmap[sk] = retmap["data"]
				bodyMap[lk+"."+sk] = retmap
				//将一个值发送到通道中。
				ch <- retmap["data"]

			}
		}
	}
	//关闭通道
	close(ch)
	returnmap = service.GetFiledData(c, WgMethod, returnmap)
	result.Success(returnmap)
}

func GetData(headerMap map[string]string, reqBody interface{}, ch2 chan interface{}) map[string]interface{} {
	var resp []byte
	var status int
	var err error
	//收集返回的数据
	defer wg.Done() // goroutine结束就登记-1
	if headerMap["method"] == "POST" {
		resp, status, err = http.NewRequest().SetHeader(headerMap).Post(headerMap["Wg-Api-Url"], reqBody)
	} else {
		resp, status, err = http.NewRequest().SetHeader(headerMap).Get(headerMap["Wg-Api-Url"], reqBody)
	}
	if status != fasthttp.StatusOK || err != nil {
		err = fmt.Errorf("接口请求失败，HTTP状态码: %d", status)
	}

	var datas service.Result
	json.Unmarshal(resp, &datas)
	retmap := make(map[string]interface{})
	err = json.UnmarshalInterface(datas, &retmap)
	ch2 <- datas
	return retmap
}
