package controller

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"tea-api/model"
	"tea-api/setting"
	"tea-api/setting/operation_setting"
)

func GetPricing(c *gin.Context) {
	pricing := model.GetPricing()
	userId, exists := c.Get("id")
	usableGroup := map[string]string{}
	groupRatio := map[string]float64{}
	for s, f := range setting.GetGroupRatioCopy() {
		groupRatio[s] = f
	}
	var group string
	if exists {
		user, err := model.GetUserCache(userId.(int))
		if err == nil {
			group = user.Group
		}
	}

	usableGroup = setting.GetUserUsableGroups(group)
	// check groupRatio contains usableGroup
	for group := range setting.GetGroupRatioCopy() {
		if _, ok := usableGroup[group]; !ok {
			delete(groupRatio, group)
		}
	}

	c.JSON(200, gin.H{
		"success":      true,
		"data":         pricing,
		"group_ratio":  groupRatio,
		"usable_group": usableGroup,
	})
}

func ResetModelRatio(c *gin.Context) {
	defaultStr := operation_setting.DefaultModelRatio2JSONString()
	err := model.UpdateOption("ModelRatio", defaultStr)
	if err != nil {
		c.JSON(200, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	err = operation_setting.UpdateModelRatioByJSONString(defaultStr)
	if err != nil {
		c.JSON(200, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	c.JSON(200, gin.H{
		"success": true,
		"message": "重置模型倍率成功",
	})
}

type UpdatePricingRequest struct {
	ModelRatios      map[string]float64 `json:"modelRatios"`
	CompletionRatios map[string]float64 `json:"completionRatios"`
}

func UpdatePricing(c *gin.Context) {
	var req UpdatePricingRequest
	err := c.ShouldBindJSON(&req)
	if err != nil {
		c.JSON(200, gin.H{
			"success": false,
			"message": "参数错误: " + err.Error(),
		})
		return
	}

	// 更新模型倍率
	if len(req.ModelRatios) > 0 {
		// 获取当前的模型倍率配置
		currentRatios := operation_setting.GetModelRatioMap()

		// 更新指定的模型倍率
		for modelName, ratio := range req.ModelRatios {
			currentRatios[modelName] = ratio
		}

		// 转换为JSON字符串并保存
		ratioBytes, err := json.Marshal(currentRatios)
		if err != nil {
			c.JSON(200, gin.H{
				"success": false,
				"message": "序列化模型倍率失败: " + err.Error(),
			})
			return
		}

		err = model.UpdateOption("ModelRatio", string(ratioBytes))
		if err != nil {
			c.JSON(200, gin.H{
				"success": false,
				"message": "保存模型倍率失败: " + err.Error(),
			})
			return
		}

		err = operation_setting.UpdateModelRatioByJSONString(string(ratioBytes))
		if err != nil {
			c.JSON(200, gin.H{
				"success": false,
				"message": "更新模型倍率失败: " + err.Error(),
			})
			return
		}
	}

	// 更新补全倍率
	if len(req.CompletionRatios) > 0 {
		// 获取当前的补全倍率配置
		currentCompletionRatios := operation_setting.GetCompletionRatioMap()

		// 更新指定的补全倍率
		for modelName, ratio := range req.CompletionRatios {
			currentCompletionRatios[modelName] = ratio
		}

		// 转换为JSON字符串并保存
		completionRatioBytes, err := json.Marshal(currentCompletionRatios)
		if err != nil {
			c.JSON(200, gin.H{
				"success": false,
				"message": "序列化补全倍率失败: " + err.Error(),
			})
			return
		}

		err = model.UpdateOption("CompletionRatio", string(completionRatioBytes))
		if err != nil {
			c.JSON(200, gin.H{
				"success": false,
				"message": "保存补全倍率失败: " + err.Error(),
			})
			return
		}

		err = operation_setting.UpdateCompletionRatioByJSONString(string(completionRatioBytes))
		if err != nil {
			c.JSON(200, gin.H{
				"success": false,
				"message": "更新补全倍率失败: " + err.Error(),
			})
			return
		}
	}

	c.JSON(200, gin.H{
		"success": true,
		"message": "价格更新成功",
	})
}
