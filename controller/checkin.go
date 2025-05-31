package controller

import (
	"encoding/json"
	"net/http"
	"strconv"
	"tea-api/common"
	"tea-api/model"
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
	// 构建特殊奖励数据
	specialRewards := []map[string]interface{}{}
	for i, day := range common.SpecialRewardDays {
		if i < len(common.SpecialRewards) {
			reward := common.SpecialRewards[i]
			var name, description string

			// 根据天数设置不同的名称和描述
			if day == 7 {
				name = "周奖励"
				description = "连续签到7天可获得额外奖励"
			} else if day == 15 {
				name = "半月奖励"
				description = "连续签到15天可获得额外奖励"
			} else if day == 30 {
				name = "月奖励"
				description = "连续签到30天可获得额外奖励"
			} else {
				name = "特殊奖励"
				description = "连续签到" + strconv.Itoa(day) + "天可获得额外奖励"
			}

			specialRewards = append(specialRewards, map[string]interface{}{
				"name":        name,
				"description": description,
				"day":         day,
				"reward":      reward,
			})
		}
	}

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
				"special_rewards":   specialRewards,
				"weekly_bonus":      20000, // 周奖励
				"monthly_bonus":     50000, // 月奖励
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

	// 将配置保存到数据库中
	if err := model.UpdateOption("CheckinEnabled", strconv.FormatBool(config.CheckinEnabled)); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "保存失败: " + err.Error(),
		})
		return
	}
	if err := model.UpdateOption("BaseCheckinReward", strconv.Itoa(config.BaseReward)); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "保存失败: " + err.Error(),
		})
		return
	}
	if err := model.UpdateOption("ContinuousCheckinReward", strconv.Itoa(config.ContinuousReward)); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "保存失败: " + err.Error(),
		})
		return
	}
	if err := model.UpdateOption("MaxContinuousRewardDays", strconv.Itoa(config.MaxDays)); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "保存失败: " + err.Error(),
		})
		return
	}
	if err := model.UpdateOption("CheckinStreakReset", strconv.FormatBool(config.StreakReset)); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "保存失败: " + err.Error(),
		})
		return
	}

	// 保存特殊奖励日和奖励额度，需要转换为JSON字符串
	specialDaysBytes, err := json.Marshal(config.SpecialDays)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "特殊奖励日序列化失败: " + err.Error(),
		})
		return
	}
	if err := model.UpdateOption("SpecialRewardDays", string(specialDaysBytes)); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "保存失败: " + err.Error(),
		})
		return
	}

	specialRewardsBytes, err := json.Marshal(config.SpecialRewards)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "特殊奖励额度序列化失败: " + err.Error(),
		})
		return
	}
	if err := model.UpdateOption("SpecialRewards", string(specialRewardsBytes)); err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "保存失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "更新成功",
	})
}
