package middleware

import (
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"tea-api/common"
	"tea-api/model"
)

// FastAuthCache 快速认证缓存
type FastAuthCache struct {
	mu    sync.RWMutex
	cache map[string]*CachedAuthInfo
}

// CachedAuthInfo 缓存的认证信息
type CachedAuthInfo struct {
	UserID       int
	TokenID      int
	TokenKey     string
	TokenName    string
	UnlimitedQuota bool
	RemainQuota  int64
	UserEnabled  bool
	ExpiresAt    time.Time
}

var fastAuthCache = &FastAuthCache{
	cache: make(map[string]*CachedAuthInfo),
}

// FastAuth 快速认证中间件，优化首字时延
func FastAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查是否启用快速路径
		config := common.GetLatencyOptConfig()
		if !config.EnableFastPath {
			// 回退到标准认证
			c.Next()
			return
		}

		key := c.Request.Header.Get("Authorization")
		if key == "" {
			key = c.Request.Header.Get("api-key")
		}
		if key == "" {
			abortWithOpenAiMessage(c, http.StatusUnauthorized, "API key is required")
			return
		}

		if strings.HasPrefix(key, "Bearer ") {
			key = key[7:]
		}

		// 尝试从快速缓存获取认证信息
		if authInfo := fastAuthCache.get(key); authInfo != nil {
			// 缓存命中，直接使用缓存的认证信息
			setAuthContext(c, authInfo)
			c.Next()
			return
		}

		// 缓存未命中，执行标准认证流程
		if err := performStandardAuth(c, key); err != nil {
			abortWithOpenAiMessage(c, http.StatusUnauthorized, err.Error())
			return
		}

		c.Next()
	}
}

// get 从缓存获取认证信息
func (cache *FastAuthCache) get(key string) *CachedAuthInfo {
	cache.mu.RLock()
	defer cache.mu.RUnlock()

	hashedKey := common.GenerateHMAC(key)
	authInfo, exists := cache.cache[hashedKey]
	if !exists {
		return nil
	}

	// 检查是否过期
	if time.Now().After(authInfo.ExpiresAt) {
		// 异步清理过期缓存
		go cache.delete(hashedKey)
		return nil
	}

	return authInfo
}

// set 设置缓存
func (cache *FastAuthCache) set(key string, authInfo *CachedAuthInfo) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	hashedKey := common.GenerateHMAC(key)
	cache.cache[hashedKey] = authInfo
}

// delete 删除缓存
func (cache *FastAuthCache) delete(hashedKey string) {
	cache.mu.Lock()
	defer cache.mu.Unlock()

	delete(cache.cache, hashedKey)
}

// performStandardAuth 执行标准认证流程
func performStandardAuth(c *gin.Context, key string) error {
	token, err := model.ValidateUserToken(key)
	if err != nil {
		return err
	}

	userCache, err := model.GetUserCache(token.UserId)
	if err != nil {
		return err
	}

	userEnabled := userCache.Status == common.UserStatusEnabled
	if !userEnabled {
		return common.NewError("用户已被封禁")
	}

	// 创建缓存条目
	authInfo := &CachedAuthInfo{
		UserID:         token.UserId,
		TokenID:        token.Id,
		TokenKey:       token.Key,
		TokenName:      token.Name,
		UnlimitedQuota: token.UnlimitedQuota,
		RemainQuota:    token.RemainQuota,
		UserEnabled:    userEnabled,
		ExpiresAt:      time.Now().Add(5 * time.Minute), // 5分钟缓存
	}

	// 设置到快速缓存
	fastAuthCache.set(key, authInfo)

	// 设置上下文
	setAuthContext(c, authInfo)
	userCache.WriteContext(c)

	return nil
}

// setAuthContext 设置认证上下文
func setAuthContext(c *gin.Context, authInfo *CachedAuthInfo) {
	c.Set("id", authInfo.UserID)
	c.Set("token_id", authInfo.TokenID)
	c.Set("token_key", authInfo.TokenKey)
	c.Set("token_name", authInfo.TokenName)
	c.Set("token_unlimited_quota", authInfo.UnlimitedQuota)
	if !authInfo.UnlimitedQuota {
		c.Set("token_quota", authInfo.RemainQuota)
	}
}

// CleanupExpiredCache 清理过期缓存
func CleanupExpiredCache() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			fastAuthCache.mu.Lock()
			now := time.Now()
			for key, authInfo := range fastAuthCache.cache {
				if now.After(authInfo.ExpiresAt) {
					delete(fastAuthCache.cache, key)
				}
			}
			fastAuthCache.mu.Unlock()
		}
	}
}

// init 初始化快速认证缓存清理
func init() {
	go CleanupExpiredCache()
}
