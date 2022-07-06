package service

/**
 * 处理相关params 和 condition 数据
 * Create by :liming10
 * Created on: 2022/04/10 15:08
 */
import (
	"fmt"
	"runtime"

	//"wg_goservice/pkg/headers"
	"wg_goservice/pkg/json"
	"wg_goservice/pkg/result"
	"wg_goservice/rediscache"

	"github.com/gin-gonic/gin"
	"github.com/sohaha/zlsgo/ztype"
)

/*
*获取params
*headerMap  header参数
*reqBody   	body和query参数
 */
func GetParams(params map[string]interface{}, conMap map[string]interface{}, headerMap map[string]string, reqBody interface{}, reqData interface{}, bodyMap map[string]interface{}) map[string]interface{} {
	//提取params中的字段
	paramsVal := make(map[string]interface{})
	json.UnmarshalInterface(params["params"], &paramsVal)
	//处理query数据
	reqDataVal := make(map[string]interface{})
	json.UnmarshalInterface(reqData, &reqDataVal)
	nparams := make(map[string]interface{})
	reqBodyVal := make(map[string]interface{})
	json.UnmarshalInterface(reqBody, &reqBodyVal)
	for pk, pv := range paramsVal {
		pval := make(map[string]interface{})
		json.UnmarshalInterface(pv, &pval)
		//处理切片数据
		for sk, sv := range pval {
			svstr := ztype.ToString(sv)
			//处理相关数据
			if pk == "body" {
				fmt.Println("这是body数据", reqBodyVal)
				if reqDataVal[svstr] != nil {
					if reqDataVal[svstr] != "" {
						nparams[sk] = reqDataVal[svstr]
					}
				}
			} else if pk == "header" {
				if headerMap[svstr] != "" {
					nparams[sk] = headerMap[svstr]
				}
			} else if pk == "query" {
				if reqDataVal[svstr] != nil {
					if reqDataVal[svstr] != "" {
						nparams[sk] = reqDataVal[svstr]
					}
				}
			} else {
				if bodyMap[pk] != nil {
					//处理非body,header,query相关数据
					bodyMapVal := make(map[string]interface{})
					json.UnmarshalInterface(bodyMap[pk], &bodyMapVal)
					dataMapVal2 := make(map[string]interface{})
					json.UnmarshalInterface(bodyMapVal["data"], &dataMapVal2)
					if dataMapVal2[svstr] != "" {
						nparams[sk] = dataMapVal2[svstr]
					}
				}
			}
		}
	}
	return nparams

}

func GetCondition(params map[string]interface{}, conmap map[string]interface{}, reqBody interface{}, reqData interface{}) bool {
	nconmapVal := make(map[string]interface{})
	json.UnmarshalInterface(conmap, &nconmapVal)

	datas := make(map[string]interface{})
	json.UnmarshalInterface(nconmapVal["data"], &datas)
	fmt.Println("condition", params["condition"])
	if params["condition"] == nil {
		return true
	}
	var isTrue bool
	isTrue = false
	for _, sv := range params["condition"].([]interface{}) {
		sVal := make(map[string]interface{})
		field := ztype.ToString(sVal["field"])
		json.UnmarshalInterface(sv, &sVal)
		if sVal["field"] == "code" {
			if nconmapVal[field] != nil {
				//转化成ToFloat64
				if nconmapVal[field] == ztype.ToFloat64(sVal["value"]) {
					isTrue = true
				}
			} else {
				isTrue = true
			}

		} else {
			//判断符号  不能直接赋值 判断  有数据类型限制
			if sVal["operation"] == "=" {
				if datas[field] == ztype.ToString(sVal["value"]) {
					isTrue = true
				} else {
					return false
				}
			} else if sVal["operation"] == ">" {
				if ztype.ToFloat64(datas[field]) > ztype.ToFloat64(sVal["value"]) {
					isTrue = true
				} else {
					return false
				}
			} else if sVal["operation"] == "<" {
				if ztype.ToFloat64(datas[field]) < ztype.ToFloat64(sVal["value"]) {
					isTrue = true
				} else {
					return false
				}
			} else if sVal["operation"] == ">=" {
				if ztype.ToFloat64(datas[field]) >= ztype.ToFloat64(sVal["value"]) {
					isTrue = true
				} else {
					return false
				}
			} else if sVal["operation"] == "<=" {
				if ztype.ToFloat64(datas[field]) <= ztype.ToFloat64(sVal["value"]) {
					isTrue = true
				} else {
					return false
				}
			} else if sVal["operation"] == "<>" {
				if ztype.ToString(datas[field]) != ztype.ToString(sVal["value"]) {
					isTrue = true
				} else {
					return false
				}
			} else if sVal["operation"] == "empty" {
				if datas[field] == "" {
					isTrue = true
				} else {
					return false
				}
			} else if sVal["operation"] == "not empty" {
				if datas[field] != "" {
					isTrue = true
				} else {
					return false
				}
			}
		}
	}
	return isTrue
}

/*
*字段封包
*headerMap  header参数
*datas   	要返回数据
 */
func GetFiledData(c *gin.Context, Method string, datas map[string]interface{}) map[string]interface{} {
	result := result.NewResult(c)
	fileds, err := rediscache.GetLayoutReturnCache(Method)
	if err != nil {
		result.Error(1408, "获取return_data缓存数据错误")
	}
	returnDatas := make(map[string]interface{})
	for sk, sv := range datas {
		ndatas := make(map[string]interface{})
		tempDatas := make(map[string]interface{})
		json.UnmarshalInterface(sv, &ndatas)
	OuterLoop:
		for _, fsv := range fileds[sk].([]interface{}) {
			if fsv == "*" {
				returnDatas[sk] = sv
				//跳出OuterLoop 循环
				continue OuterLoop
			} else {
				tempDatas[ztype.ToString(fsv)] = ndatas[ztype.ToString(fsv)]
			}
			returnDatas[sk] = tempDatas
		}
		tempDatas = nil
	}
	//GC释放map内存
	runtime.GC()
	return returnDatas
}
