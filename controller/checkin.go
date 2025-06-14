package controller

import (
	"net/http"
	"tea-api/common"
	"tea-api/service"

	"github.com/gin-gonic/gin"
)

func Checkin(c *gin.Context) {
	if !common.CheckinEnabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "签到功能未开启",
		})
		return
	}
	id := c.GetInt("id")
	days, err := service.Checkin(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// 获取奖励信息
	rewardInfo := service.GetCheckinRewardInfo(days)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"continuous":        days,
			"reward":            rewardInfo["reward"],
			"is_special_reward": rewardInfo["is_special_reward"],
		},
	})
}

// GetCheckinStatus 获取用户当天的签到状态
func GetCheckinStatus(c *gin.Context) {
	if !common.CheckinEnabled {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "签到功能未开启",
		})
		return
	}
	id := c.GetInt("id")
	checkedToday, continuous, err := service.GetCheckinStatus(id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}

	// 获取下一次签到可获得的奖励信息
	nextDays := continuous
	if !checkedToday {
		nextDays++ // 如果今天没签到，计算下一次签到的天数
	}
	rewardInfo := service.GetCheckinRewardInfo(nextDays)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"checked_today":   checkedToday,
			"continuous":      continuous,
			"next_reward":     rewardInfo["reward"],
			"is_special_next": rewardInfo["is_special_reward"],
		},
	})
}

// GetCheckinConfig 获取签到配置信息
func GetCheckinConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"checkin_enabled": common.CheckinEnabled,
			"checkin_config": gin.H{
				"base_reward":       common.BaseCheckinReward,
				"continuous_reward": common.ContinuousCheckinReward,
				"max_days":          common.MaxContinuousRewardDays,
				"special_days":      common.SpecialRewardDays,
				"special_rewards":   common.SpecialRewards,
				"streak_reset":      common.CheckinStreakReset,
			},
		},
	})
}

// UpdateCheckinConfig 更新签到配置
func UpdateCheckinConfig(c *gin.Context) {
	// 验证管理员权限
	if c.GetInt("role") < common.RoleAdminUser {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "无权操作",
		})
		return
	}

	var config struct {
		BaseReward       int   `json:"base_reward"`
		ContinuousReward int   `json:"continuous_reward"`
		MaxDays          int   `json:"max_days"`
		SpecialDays      []int `json:"special_days"`
		SpecialRewards   []int `json:"special_rewards"`
		StreakReset      bool  `json:"streak_reset"`
		CheckinEnabled   bool  `json:"checkin_enabled"`
	}

	if err := c.ShouldBindJSON(&config); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "参数错误",
		})
		return
	}

	// 更新配置
	common.CheckinEnabled = config.CheckinEnabled
	common.BaseCheckinReward = config.BaseReward
	common.ContinuousCheckinReward = config.ContinuousReward
	common.MaxContinuousRewardDays = config.MaxDays
	common.SpecialRewardDays = config.SpecialDays
	common.SpecialRewards = config.SpecialRewards
	common.CheckinStreakReset = config.StreakReset

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "更新成功",
	})
}
