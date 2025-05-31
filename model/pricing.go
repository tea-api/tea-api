package model

import (
	"strconv"
	"sync"
	"tea-api/common"
	"tea-api/setting/operation_setting"
	"time"
)

type Pricing struct {
	ModelName       string   `json:"model_name"`
	QuotaType       int      `json:"quota_type"`
	ModelRatio      float64  `json:"model_ratio"`
	ModelPrice      float64  `json:"model_price"`
	OwnerBy         string   `json:"owner_by"`
	CompletionRatio float64  `json:"completion_ratio"`
	EnableGroup     []string `json:"enable_groups,omitempty"`
}

var (
	pricingMap         []Pricing
	lastGetPricingTime time.Time
	updatePricingLock  sync.Mutex
)

func GetPricing() []Pricing {
	updatePricingLock.Lock()
	defer updatePricingLock.Unlock()

	if time.Since(lastGetPricingTime) > time.Minute*1 || len(pricingMap) == 0 {
		updatePricing()
	}
	//if group != "" {
	//	userPricingMap := make([]Pricing, 0)
	//	models := GetGroupModels(group)
	//	for _, pricing := range pricingMap {
	//		if !common.StringsContains(models, pricing.ModelName) {
	//			pricing.Available = false
	//		}
	//		userPricingMap = append(userPricingMap, pricing)
	//	}
	//	return userPricingMap
	//}
	return pricingMap
}

// ClearPricingCache 清除价格缓存，强制下次获取时重新加载
func ClearPricingCache() {
	updatePricingLock.Lock()
	defer updatePricingLock.Unlock()

	// 重置缓存时间，强制下次获取时更新
	lastGetPricingTime = time.Time{}
	// 清空缓存数据
	pricingMap = nil
}

func updatePricing() {
	//modelRatios := common.GetModelRatios()
	enableAbilities := GetAllEnableAbilities()
	modelGroupsMap := make(map[string][]string)
	for _, ability := range enableAbilities {
		groups := modelGroupsMap[ability.Model]
		if groups == nil {
			groups = make([]string, 0)
		}
		if !common.StringsContains(groups, ability.Group) {
			groups = append(groups, ability.Group)
		}
		modelGroupsMap[ability.Model] = groups
	}

	common.SysLog("开始更新价格信息...")

	// 获取当前的补全倍率映射，用于调试
	compRatioMap := operation_setting.GetCompletionRatioMap()
	common.SysLog("当前内存中的补全倍率映射:")
	for model, ratio := range compRatioMap {
		common.SysLog("模型: " + model + ", 补全倍率: " + strconv.FormatFloat(ratio, 'f', 3, 64))
	}

	pricingMap = make([]Pricing, 0)
	for model, groups := range modelGroupsMap {
		pricing := Pricing{
			ModelName:   model,
			EnableGroup: groups,
		}
		modelPrice, findPrice := operation_setting.GetModelPrice(model, false)
		if findPrice {
			pricing.ModelPrice = modelPrice
			pricing.QuotaType = 1
		} else {
			modelRatio, _ := operation_setting.GetModelRatio(model)
			pricing.ModelRatio = modelRatio

			// 直接从补全倍率映射中获取，避免硬编码覆盖
			if ratio, ok := compRatioMap[model]; ok {
				pricing.CompletionRatio = ratio
				common.SysLog("使用自定义补全倍率 - 模型: " + model + ", 补全倍率: " + strconv.FormatFloat(ratio, 'f', 3, 64))
			} else {
				pricing.CompletionRatio = operation_setting.GetCompletionRatio(model)
				common.SysLog("使用默认补全倍率 - 模型: " + model + ", 补全倍率: " + strconv.FormatFloat(pricing.CompletionRatio, 'f', 3, 64))
			}

			pricing.QuotaType = 0
		}
		pricingMap = append(pricingMap, pricing)
	}
	lastGetPricingTime = time.Now()

	common.SysLog("价格信息更新完成，共 " + strconv.Itoa(len(pricingMap)) + " 个模型")
}
