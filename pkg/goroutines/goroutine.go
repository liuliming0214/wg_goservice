package goroutines

import (
	"fmt"
	"sync"
)

//这里使用了sync.WaitGroup来实现goroutine的同步
var wg sync.WaitGroup

//弃用
func GetData(headerMap map[string]string, reqBody interface{}) map[string]string {
	fmt.Println(headerMap)
	fmt.Println(reqBody)

	return headerMap

}
