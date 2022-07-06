package global

import (
	"github.com/allegro/bigcache"
	"log"
	"time"
)
//定义一个全局的bigcache
var (
	BigCache *bigcache.BigCache
)

//创建一个全局的bigcache
func SetupBigCache() (error) {
	config := bigcache.Config {
		Shards: 1024,   			// 存储的条目数量，值必须是2的幂
		LifeWindow: 3*time.Minute,	// 超时后条目被处理
		CleanWindow: 2*time.Minute, //处理超时条目的时间范围
		MaxEntriesInWindow: 0,		// 在 Life Window 中的最大数量，
		MaxEntrySize: 0,			// 条目最大尺寸，以字节为单位
		HardMaxCacheSize: 0,		// 设置缓存最大值，以MB为单位，超过了不在分配内存。0表示无限制分配
	}
    var initErr error
	BigCache, initErr = bigcache.NewBigCache(config)
	if initErr != nil {
		log.Fatal(initErr)
		return initErr
	}
	//BigCache.Stats().
	return nil
}

