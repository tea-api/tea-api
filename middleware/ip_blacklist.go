package middleware

import (
	"fmt"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"tea-api/common"
	"tea-api/setting"
)

// IPBlacklistEntry IP黑名单条目
type IPBlacklistEntry struct {
	IP          string
	Reason      string
	BlockedAt   time.Time
	ExpiresAt   time.Time
	ViolationCount int
	IsTemporary bool
}

// IPBlacklistManager IP黑名单管理器
type IPBlacklistManager struct {
	mu        sync.RWMutex
	blacklist map[string]*IPBlacklistEntry
	whitelist map[string]bool
}

var ipBlacklistManager = &IPBlacklistManager{
	blacklist: make(map[string]*IPBlacklistEntry),
	whitelist: make(map[string]bool),
}

// 黑名单配置
const (
	TempBlockDuration     = 1 * time.Hour    // 临时封禁时长
	PermanentBlockDuration = 24 * time.Hour  // 永久封禁时长
	MaxViolations         = 5                // 最大违规次数
	CleanupInterval       = 10 * time.Minute // 清理间隔
)

// IPBlacklist IP黑名单中间件
func IPBlacklist() gin.HandlerFunc {
	// 启动清理协程
	go ipBlacklistManager.cleanupRoutine()
	
	return func(c *gin.Context) {
		clientIP := c.ClientIP()

		// 检查白名单
		if ipBlacklistManager.isWhitelisted(clientIP) {
			c.Next()
			return
		}

		// 检查黑名单
		if entry := ipBlacklistManager.isBlacklisted(clientIP); entry != nil {
			common.SysLog(fmt.Sprintf("blocked request from blacklisted IP %s: %s", clientIP, entry.Reason))
			abortWithBlacklistError(c, entry)
			return
		}
		c.Next()
	}
}

// isWhitelisted 检查IP是否在白名单中
func (manager *IPBlacklistManager) isWhitelisted(ip string) bool {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	
	// 检查精确匹配
	if manager.whitelist[ip] {
		return true
	}
	
	// 检查内网IP (仅在生产环境中自动白名单)
	// 在测试环境中，我们需要能够封禁内网IP进行测试
	// if isPrivateIP(ip) {
	//     return true
	// }
	
	return false
}

// isBlacklisted 检查IP是否在黑名单中
func (manager *IPBlacklistManager) isBlacklisted(ip string) *IPBlacklistEntry {
	manager.mu.RLock()
	defer manager.mu.RUnlock()
	
	entry, exists := manager.blacklist[ip]
	if !exists {
		return nil
	}
	
	// 检查是否已过期
	if entry.IsTemporary && time.Now().After(entry.ExpiresAt) {
		return nil
	}
	
	return entry
}

// AddToBlacklist 添加IP到黑名单
func (manager *IPBlacklistManager) AddToBlacklist(ip, reason string, temporary bool) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	
	now := time.Now()
	var expiresAt time.Time
	
	if temporary {
		expiresAt = now.Add(TempBlockDuration)
	} else {
		expiresAt = now.Add(PermanentBlockDuration)
	}
	
	// 如果已存在，增加违规次数
	if existing, exists := manager.blacklist[ip]; exists {
		existing.ViolationCount++
		existing.Reason = reason
		existing.BlockedAt = now
		existing.ExpiresAt = expiresAt
		
		// 违规次数过多，转为永久封禁
		if existing.ViolationCount >= MaxViolations {
			existing.IsTemporary = false
			existing.ExpiresAt = now.Add(PermanentBlockDuration)
		}
	} else {
		manager.blacklist[ip] = &IPBlacklistEntry{
			IP:             ip,
			Reason:         reason,
			BlockedAt:      now,
			ExpiresAt:      expiresAt,
			ViolationCount: 1,
			IsTemporary:    temporary,
		}
	}

	common.SysLog(fmt.Sprintf("added IP %s to blacklist: %s (temporary: %v)", ip, reason, temporary))

	// 记录安全日志
	setting.AddSecurityLog("ip_blacklist", ip,
		fmt.Sprintf("IP已加入黑名单: %s", reason), "blacklisted",
		map[string]interface{}{
			"reason": reason,
			"temporary": temporary,
			"violations": manager.blacklist[ip].ViolationCount,
		})
}

// RemoveFromBlacklist 从黑名单移除IP
func (manager *IPBlacklistManager) RemoveFromBlacklist(ip string) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	
	if _, exists := manager.blacklist[ip]; exists {
		delete(manager.blacklist, ip)
		common.SysLog(fmt.Sprintf("removed IP %s from blacklist", ip))
	}
}

// AddToWhitelist 添加IP到白名单
func (manager *IPBlacklistManager) AddToWhitelist(ip string) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	
	manager.whitelist[ip] = true
	common.SysLog(fmt.Sprintf("added IP %s to whitelist", ip))
}

// RemoveFromWhitelist 从白名单移除IP
func (manager *IPBlacklistManager) RemoveFromWhitelist(ip string) {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	
	delete(manager.whitelist, ip)
	common.SysLog(fmt.Sprintf("removed IP %s from whitelist", ip))
}

// cleanupRoutine 清理过期的黑名单条目
func (manager *IPBlacklistManager) cleanupRoutine() {
	ticker := time.NewTicker(CleanupInterval)
	defer ticker.Stop()
	
	for range ticker.C {
		manager.cleanupExpiredEntries()
	}
}

// cleanupExpiredEntries 清理过期条目
func (manager *IPBlacklistManager) cleanupExpiredEntries() {
	manager.mu.Lock()
	defer manager.mu.Unlock()
	
	now := time.Now()
	for ip, entry := range manager.blacklist {
		if entry.IsTemporary && now.After(entry.ExpiresAt) {
			delete(manager.blacklist, ip)
			common.SysLog(fmt.Sprintf("removed expired blacklist entry for IP %s", ip))
		}
	}
}

// isPrivateIP 检查是否为内网IP
func isPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	
	// 检查是否为私有网络
	privateNetworks := []string{
		"10.0.0.0/8",
		"172.16.0.0/12",
		"192.168.0.0/16",
		"127.0.0.0/8",
		"::1/128",
		"fc00::/7",
	}
	
	for _, network := range privateNetworks {
		_, cidr, err := net.ParseCIDR(network)
		if err != nil {
			continue
		}
		if cidr.Contains(parsedIP) {
			return true
		}
	}
	
	return false
}

// abortWithBlacklistError 中止请求并返回黑名单错误
func abortWithBlacklistError(c *gin.Context, entry *IPBlacklistEntry) {
	var message string
	if entry.IsTemporary {
		remaining := entry.ExpiresAt.Sub(time.Now())
		message = fmt.Sprintf("您的IP已被临时封禁，原因：%s，剩余时间：%v", entry.Reason, remaining.Round(time.Minute))
	} else {
		message = fmt.Sprintf("您的IP已被永久封禁，原因：%s", entry.Reason)
	}
	
	c.JSON(http.StatusForbidden, gin.H{
		"error": gin.H{
			"message":    message,
			"type":       "ip_blocked",
			"code":       "ip_blacklisted",
			"blocked_at": entry.BlockedAt.Format(time.RFC3339),
			"expires_at": entry.ExpiresAt.Format(time.RFC3339),
			"violations": entry.ViolationCount,
		},
	})
	c.Abort()
}

// GetBlacklistStats 获取黑名单统计信息
func GetBlacklistStats() map[string]interface{} {
	ipBlacklistManager.mu.RLock()
	defer ipBlacklistManager.mu.RUnlock()
	
	stats := map[string]interface{}{
		"total_blacklisted": len(ipBlacklistManager.blacklist),
		"total_whitelisted": len(ipBlacklistManager.whitelist),
		"temporary_blocks":  0,
		"permanent_blocks":  0,
	}
	
	for _, entry := range ipBlacklistManager.blacklist {
		if entry.IsTemporary {
			stats["temporary_blocks"] = stats["temporary_blocks"].(int) + 1
		} else {
			stats["permanent_blocks"] = stats["permanent_blocks"].(int) + 1
		}
	}
	
	return stats
}

// AutoBlacklistIP 自动封禁IP（供其他中间件调用）
func AutoBlacklistIP(ip, reason string) {
	ipBlacklistManager.AddToBlacklist(ip, reason, true)
}

// AutoPermanentBlacklistIP 自动永久封禁IP
func AutoPermanentBlacklistIP(ip, reason string) {
	ipBlacklistManager.AddToBlacklist(ip, reason, false)
}

// GetBlacklistData 获取黑名单详细数据
func (manager *IPBlacklistManager) GetBlacklistData() []map[string]interface{} {
	manager.mu.RLock()
	defer manager.mu.RUnlock()

	var blacklistData []map[string]interface{}
	for _, entry := range manager.blacklist {
		data := map[string]interface{}{
			"ip":           entry.IP,
			"reason":       entry.Reason,
			"blocked_at":   entry.BlockedAt.Format(time.RFC3339),
			"expires_at":   entry.ExpiresAt.Format(time.RFC3339),
			"violations":   entry.ViolationCount,
			"is_temporary": entry.IsTemporary,
		}
		blacklistData = append(blacklistData, data)
	}

	return blacklistData
}

// GetBlacklistManager 获取黑名单管理器实例
func GetBlacklistManager() *IPBlacklistManager {
	return ipBlacklistManager
}
