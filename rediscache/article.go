package rediscache

import (
	"fmt"
	"wg_goservice/global"
	"wg_goservice/model"
	"wg_goservice/pkg/json"

	"github.com/go-redis/redis"
)

//从cache得到method 信息
func GetMethodRedisCache(Method string) (*model.Method, error) {
	fmt.Println("redis:GetMethodRedisCache")
	key := fmt.Sprintf("wg.system.%s", Method)
	val, err := global.RedisDb.Get(key).Result()
	if err == redis.Nil || err != nil {
		return nil, err
	} else {
		method := model.Method{}
		if err := json.Unmarshal([]byte(val), &method); err != nil {
			return nil, err
		}
		return &method, nil
	}
}

//从cache得到编排配置信息
func GetLayoutRedisCache(Method string) (map[string]interface{}, error) {
	fmt.Println("redis:GetLayoutRedisCache")
	key := fmt.Sprintf("wg.layout.params.%s", Method)
	val, err := global.RedisDb.Get(key).Result()
	if err == redis.Nil || err != nil {
		return nil, err
	} else {
		var retmap map[string]interface{}
		if err := json.Unmarshal([]byte(val), &retmap); err != nil {
			return nil, err
		}
		return retmap, nil
	}
}

func GetNewMethodRedisCache(Method string) (map[string]interface{}, error) {
	fmt.Println("redis:GetNewMethodRedisCache")
	key := fmt.Sprintf("wg.system.%s", Method)
	val, err := global.RedisDb.Get(key).Result()
	if err == redis.Nil || err != nil {
		return nil, err
	} else {
		var retmap map[string]interface{}
		if err := json.Unmarshal([]byte(val), &retmap); err != nil {
			return nil, err
		}
		return retmap, nil
	}
}

func GetLayoutReturnCache(Method string) (map[string]interface{}, error) {
	key := fmt.Sprintf("wg.layout.return_data.%s", Method)
	val, err := global.RedisDb.Get(key).Result()
	if err == redis.Nil || err != nil {
		return nil, err
	} else {
		var retmap map[string]interface{}
		if err := json.Unmarshal([]byte(val), &retmap); err != nil {
			return nil, err
		}
		return retmap, nil
	}
}
