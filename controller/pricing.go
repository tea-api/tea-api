package controller

import (
	"encoding/json"
	"tea-api/model"
	"tea-api/setting"
	"tea-api/setting/operation_setting"

	"github.com/gin-gonic/gin"
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

type UpdateModelPricingRequest struct {
	ModelName   string  `json:"model_name"`
	InputPrice  float64 `json:"input_price"`
	OutputPrice float64 `json:"output_price"`
	Unit        string  `json:"unit"`
}

func UpdateModelPricing(c *gin.Context) {
	var req UpdateModelPricingRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, gin.H{"success": false, "message": "invalid request"})
		return
	}

	multiplier := 1.0
	if req.Unit == "1k" || req.Unit == "k" {
		multiplier = 1000.0
	}

	inputPerM := req.InputPrice * multiplier
	outputPerM := req.OutputPrice * multiplier

	basePerM := 2.0
	ratio := 0.0
	if basePerM != 0 {
		ratio = inputPerM / basePerM
	}
	completionRatio := 0.0
	if inputPerM != 0 {
		completionRatio = outputPerM / inputPerM
	}

	var ratioMap map[string]float64
	_ = json.Unmarshal([]byte(operation_setting.ModelRatio2JSONString()), &ratioMap)
	if ratioMap == nil {
		ratioMap = make(map[string]float64)
	}
	ratioMap[req.ModelName] = ratio
	ratioBytes, _ := json.Marshal(ratioMap)

	var compMap map[string]float64
	_ = json.Unmarshal([]byte(operation_setting.CompletionRatio2JSONString()), &compMap)
	if compMap == nil {
		compMap = make(map[string]float64)
	}
	compMap[req.ModelName] = completionRatio
	compBytes, _ := json.Marshal(compMap)

	priceMap := operation_setting.GetModelPriceMap()
	newPriceMap := make(map[string]float64)
	for k, v := range priceMap {
		if k != req.ModelName {
			newPriceMap[k] = v
		}
	}
	priceBytes, _ := json.Marshal(newPriceMap)

	if err := model.UpdateOption("ModelRatio", string(ratioBytes)); err != nil {
		c.JSON(200, gin.H{"success": false, "message": err.Error()})
		return
	}
	if err := model.UpdateOption("CompletionRatio", string(compBytes)); err != nil {
		c.JSON(200, gin.H{"success": false, "message": err.Error()})
		return
	}
	if err := model.UpdateOption("ModelPrice", string(priceBytes)); err != nil {
		c.JSON(200, gin.H{"success": false, "message": err.Error()})
		return
	}

	// 更新内存中的模型比率和补全比率
	if err := operation_setting.UpdateModelRatioByJSONString(string(ratioBytes)); err != nil {
		c.JSON(200, gin.H{"success": false, "message": "保存到数据库成功，但更新内存失败: " + err.Error()})
		return
	}
	if err := operation_setting.UpdateCompletionRatioByJSONString(string(compBytes)); err != nil {
		c.JSON(200, gin.H{"success": false, "message": "保存到数据库成功，但更新内存失败: " + err.Error()})
		return
	}
	if err := operation_setting.UpdateModelPriceByJSONString(string(priceBytes)); err != nil {
		c.JSON(200, gin.H{"success": false, "message": "保存到数据库成功，但更新内存失败: " + err.Error()})
		return
	}

	c.JSON(200, gin.H{"success": true})
}
