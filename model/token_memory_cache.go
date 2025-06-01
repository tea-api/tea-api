package model

import (
	"strconv"
	"sync"
	"tea-api/common"
	"tea-api/constant"
	"time"
)

type tokenMemoryEntry struct {
	token    Token
	expireAt time.Time
}

var (
	tokenMemoryCache = make(map[string]tokenMemoryEntry)
	tokenMemoryLock  sync.RWMutex
)

func memoryCacheSetToken(t Token) {
	if !common.TokenMemoryCacheEnabled {
		return
	}
	key := t.Key
	t.Clean()
	tokenMemoryLock.Lock()
	tokenMemoryCache[key] = tokenMemoryEntry{
		token:    t,
		expireAt: time.Now().Add(time.Duration(constant.TokenCacheSeconds) * time.Second),
	}
	tokenMemoryLock.Unlock()
}

func memoryCacheDeleteToken(key string) {
	if !common.TokenMemoryCacheEnabled {
		return
	}
	tokenMemoryLock.Lock()
	delete(tokenMemoryCache, key)
	tokenMemoryLock.Unlock()
}

func memoryCacheIncrTokenQuota(key string, increment int64) {
	if !common.TokenMemoryCacheEnabled {
		return
	}
	tokenMemoryLock.Lock()
	entry, ok := tokenMemoryCache[key]
	if ok {
		entry.token.RemainQuota += int(increment)
		tokenMemoryCache[key] = entry
	}
	tokenMemoryLock.Unlock()
}

func memoryCacheSetTokenField(key string, field string, value string) {
	if !common.TokenMemoryCacheEnabled {
		return
	}
	tokenMemoryLock.Lock()
	entry, ok := tokenMemoryCache[key]
	if ok {
		switch field {
		case constant.TokenFiledRemainQuota:
			// value is string representation of int
			// parse to int
			if v, err := strconv.Atoi(value); err == nil {
				entry.token.RemainQuota = v
			}
		}
		tokenMemoryCache[key] = entry
	}
	tokenMemoryLock.Unlock()
}
func memoryCacheGetToken(key string) (*Token, error) {
	if !common.TokenMemoryCacheEnabled {
		return nil, nil
	}
	tokenMemoryLock.RLock()
	entry, ok := tokenMemoryCache[key]
	tokenMemoryLock.RUnlock()
	if !ok || time.Now().After(entry.expireAt) {
		if ok {
			tokenMemoryLock.Lock()
			delete(tokenMemoryCache, key)
			tokenMemoryLock.Unlock()
		}
		return nil, nil
	}
	t := entry.token
	t.Key = key
	return &t, nil
}
