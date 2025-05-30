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

	// 计算奖励金额以显示给用户
	reward := common.BaseCheckinReward
	if days > 1 {
		extraDays := days - 1
		if extraDays > common.MaxContinuousRewardDays {
			extraDays = common.MaxContinuousRewardDays
		}
		reward += extraDays * common.ContinuousCheckinReward
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"continuous": days,
			"reward":     reward,
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
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data": gin.H{
			"checked_today": checkedToday,
			"continuous":    continuous,
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
			},
		},
	})
}
