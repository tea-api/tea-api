package model

import (
	"errors"
	"github.com/bytedance/gopkg/util/gopool"
	"gorm.io/gorm"
	"sync"
	"tea-api/common"
	"time"
)

const (
	BatchUpdateTypeUserQuota = iota
	BatchUpdateTypeTokenQuota
	BatchUpdateTypeUsedQuota
	BatchUpdateTypeChannelUsedQuota
	BatchUpdateTypeRequestCount
	BatchUpdateTypeCount // if you add a new type, you need to add a new map and a new lock
)

var batchUpdateStores []map[int]int
var batchUpdateLocks []sync.Mutex

func init() {
	for i := 0; i < BatchUpdateTypeCount; i++ {
		batchUpdateStores = append(batchUpdateStores, make(map[int]int))
		batchUpdateLocks = append(batchUpdateLocks, sync.Mutex{})
	}
}

func InitBatchUpdater() {
	gopool.Go(func() {
		for {
			time.Sleep(time.Duration(common.BatchUpdateInterval) * time.Second)
			batchUpdate()
		}
	})
}

func addNewRecord(type_ int, id int, value int) {
	batchUpdateLocks[type_].Lock()
	defer batchUpdateLocks[type_].Unlock()
	if _, ok := batchUpdateStores[type_][id]; !ok {
		batchUpdateStores[type_][id] = value
	} else {
		batchUpdateStores[type_][id] += value
	}
}

func batchUpdate() {
	common.SysLog("batch update started")
	var wg sync.WaitGroup
	for i := 0; i < BatchUpdateTypeCount; i++ {
		batchUpdateLocks[i].Lock()
		store := batchUpdateStores[i]
		batchUpdateStores[i] = make(map[int]int)
		batchUpdateLocks[i].Unlock()
		// TODO: maybe we can combine updates with same key?
		for key, value := range store {
			wg.Add(1)
			key := key
			value := value
			typeIndex := i
			gopool.Go(func() {
				defer wg.Done()
				switch typeIndex {
				case BatchUpdateTypeUserQuota:
					if err := increaseUserQuota(key, value); err != nil {
						common.SysError("failed to batch update user quota: " + err.Error())
					}
				case BatchUpdateTypeTokenQuota:
					if err := increaseTokenQuota(key, value); err != nil {
						common.SysError("failed to batch update token quota: " + err.Error())
					}
				case BatchUpdateTypeUsedQuota:
					updateUserUsedQuota(key, value)
				case BatchUpdateTypeRequestCount:
					updateUserRequestCount(key, value)
				case BatchUpdateTypeChannelUsedQuota:
					updateChannelUsedQuota(key, value)
				}
			})
		}
	}
	wg.Wait()
	common.SysLog("batch update finished")
}

func RecordExist(err error) (bool, error) {
	if err == nil {
		return true, nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return false, nil
	}
	return false, err
}

func shouldUpdateRedis(fromDB bool, err error) bool {
	return (common.RedisEnabled || common.TokenMemoryCacheEnabled) && fromDB && err == nil
}
