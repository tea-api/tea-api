package controller

import (
	"encoding/json"
	"fmt"
	"tea-api/common"
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
		// 当输入价格和输出价格相同时，补全倍率应为1
		if inputPerM == outputPerM {
			completionRatio = 1.0
		} else {
			completionRatio = outputPerM / inputPerM
		}
	}

	// 打印日志，帮助调试
	common.SysLog(fmt.Sprintf("更新模型价格 - 模型: %s, 输入价格: %.3f, 输出价格: %.3f, 计算的补全倍率: %.3f",
		req.ModelName, req.InputPrice, req.OutputPrice, completionRatio))

	// 获取当前模型信息，确定是按量计费还是按次计费
	pricing := model.GetPricing()
	var currentModel *model.Pricing
	for _, p := range pricing {
		if p.ModelName == req.ModelName {
			currentModel = &p
			break
		}
	}

	// 打印当前模型信息
	if currentModel != nil {
		common.SysLog(fmt.Sprintf("当前模型信息 - 名称: %s, 类型: %d, 当前补全倍率: %.3f",
			currentModel.ModelName, currentModel.QuotaType, currentModel.CompletionRatio))
	}

	// 获取当前补全倍率映射
	oldCompMap := operation_setting.GetCompletionRatioMap()
	oldRatio, exists := oldCompMap[req.ModelName]
	if exists {
		common.SysLog(fmt.Sprintf("当前内存中的补全倍率: %.3f", oldRatio))
	} else {
		common.SysLog("当前内存中没有该模型的补全倍率记录")
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

	// 强制设置补全倍率，确保更新成功
	compMap[req.ModelName] = completionRatio
	compBytes, _ := json.Marshal(compMap)

	common.SysLog(fmt.Sprintf("即将更新补全倍率 - 模型: %s, 新补全倍率: %.3f", req.ModelName, completionRatio))

	priceMap := operation_setting.GetModelPriceMap()
	newPriceMap := make(map[string]float64)
	for k, v := range priceMap {
		if k != req.ModelName {
			newPriceMap[k] = v
		}
	}

	// 如果是按次计费模型，则保存价格
	if currentModel != nil && currentModel.QuotaType == 1 {
		newPriceMap[req.ModelName] = req.InputPrice
	}

	priceBytes, _ := json.Marshal(newPriceMap)

	// 先更新内存中的数据，确保立即生效
	// 更新内存中的模型比率和补全比率
	if err := operation_setting.UpdateModelRatioByJSONString(string(ratioBytes)); err != nil {
		c.JSON(200, gin.H{"success": false, "message": "更新内存失败: " + err.Error()})
		return
	}
	if err := operation_setting.UpdateCompletionRatioByJSONString(string(compBytes)); err != nil {
		c.JSON(200, gin.H{"success": false, "message": "更新内存失败: " + err.Error()})
		return
	}
	if err := operation_setting.UpdateModelPriceByJSONString(string(priceBytes)); err != nil {
		c.JSON(200, gin.H{"success": false, "message": "更新内存失败: " + err.Error()})
		return
	}

	// 然后更新数据库
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

	// 强制清除缓存，确保下次获取价格时是最新的
	model.ClearPricingCache()

	// 验证更新是否成功
	newCompMap := operation_setting.GetCompletionRatioMap()
	newRatio, newExists := newCompMap[req.ModelName]
	if newExists {
		common.SysLog(fmt.Sprintf("更新后内存中的补全倍率: %.3f", newRatio))
	} else {
		common.SysLog("更新后内存中仍然没有该模型的补全倍率记录")
	}

	// 返回更新后的补全倍率，便于前端显示
	c.JSON(200, gin.H{
		"success":          true,
		"completion_ratio": completionRatio,
	})
}

// 新增直接更新补全倍率的API
type UpdateCompletionRatioRequest struct {
	ModelName       string  `json:"model_name"`
	CompletionRatio float64 `json:"completion_ratio"`
}

func UpdateCompletionRatio(c *gin.Context) {
	// 验证请求
	var req UpdateCompletionRatioRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(200, gin.H{"success": false, "message": "无效请求"})
		return
	}

	// 验证模型是否存在
	pricing := model.GetPricing()
	var currentModel *model.Pricing
	for _, p := range pricing {
		if p.ModelName == req.ModelName {
			currentModel = &p
			break
		}
	}

	if currentModel == nil {
		c.JSON(200, gin.H{"success": false, "message": "模型不存在"})
		return
	}

	// 只允许为按量计费模型设置补全倍率
	if currentModel.QuotaType != 0 {
		c.JSON(200, gin.H{"success": false, "message": "只能为按量计费模型设置补全倍率"})
		return
	}

	// 获取当前的补全倍率映射
	var compMap map[string]float64
	_ = json.Unmarshal([]byte(operation_setting.CompletionRatio2JSONString()), &compMap)
	if compMap == nil {
		compMap = make(map[string]float64)
	}

	// 更新补全倍率
	compMap[req.ModelName] = req.CompletionRatio
	compBytes, _ := json.Marshal(compMap)

	// 先更新内存中的数据，确保立即生效
	if err := operation_setting.UpdateCompletionRatioByJSONString(string(compBytes)); err != nil {
		c.JSON(200, gin.H{"success": false, "message": "更新内存失败: " + err.Error()})
		return
	}

	// 然后更新数据库
	if err := model.UpdateOption("CompletionRatio", string(compBytes)); err != nil {
		c.JSON(200, gin.H{"success": false, "message": err.Error()})
		return
	}

	// 强制清除缓存，确保下次获取价格时是最新的
	model.ClearPricingCache()

	c.JSON(200, gin.H{
		"success": true,
		"message": "补全倍率已更新",
	})
}
