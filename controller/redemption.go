package controller

import (
	"net/http"
	"strconv"
	"strings"
	"tea-api/common"
	"tea-api/model"

	"github.com/gin-gonic/gin"
)

func GetAllRedemptions(c *gin.Context) {
	p, _ := strconv.Atoi(c.Query("p"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	if p < 0 {
		p = 0
	}
	if pageSize < 1 {
		pageSize = common.ItemsPerPage
	}
	redemptions, total, err := model.GetAllRedemptions((p-1)*pageSize, pageSize)
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
			"items":     redemptions,
			"total":     total,
			"page":      p,
			"page_size": pageSize,
		},
	})
	return
}

func SearchRedemptions(c *gin.Context) {
	keyword := c.Query("keyword")
	p, _ := strconv.Atoi(c.Query("p"))
	pageSize, _ := strconv.Atoi(c.Query("page_size"))
	if p < 0 {
		p = 0
	}
	if pageSize < 1 {
		pageSize = common.ItemsPerPage
	}
	redemptions, total, err := model.SearchRedemptions(keyword, (p-1)*pageSize, pageSize)
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
			"items":     redemptions,
			"total":     total,
			"page":      p,
			"page_size": pageSize,
		},
	})
	return
}

func GetRedemption(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	redemption, err := model.GetRedemptionById(id)
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
		"data":    redemption,
	})
	return
}

func AddRedemption(c *gin.Context) {
	redemption := model.Redemption{}
	err := c.ShouldBindJSON(&redemption)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if len(redemption.Name) == 0 || len(redemption.Name) > 20 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "兑换码名称长度必须在1-20之间",
		})
		return
	}

	// 检查是否是完全自定义的兑换码
	isFullyCustom := redemption.Key != "" && !strings.HasSuffix(redemption.Key, "*")

	// 检查是否是带前缀的批量生成
	isCustomPrefix := redemption.Key != "" && strings.HasSuffix(redemption.Key, "*")

	// 完全自定义的兑换码只能生成一个
	if isFullyCustom && redemption.Count > 1 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "自定义兑换码只能生成一个",
		})
		return
	}

	if redemption.Count <= 0 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "兑换码个数必须大于0",
		})
		return
	}
	if redemption.Count > 100 {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": "一次兑换码批量生成的个数不能大于 100",
		})
		return
	}
	if redemption.MaxTimes <= 0 {
		redemption.MaxTimes = 1
	}
	var keys []string
	total := redemption.Count
	if isFullyCustom {
		total = 1
	}

	for i := 0; i < total; i++ {
		var key string
		if isFullyCustom {
			// 完全自定义兑换码
			key = redemption.Key
		} else if isCustomPrefix {
			// 带前缀的批量生成
			prefix := strings.TrimSuffix(redemption.Key, "*")
			key = prefix + common.GetUUID()[:16] // 使用较短的UUID
		} else {
			// 完全随机生成
			key = common.GetUUID()
		}

		cleanRedemption := model.Redemption{
			UserId:      c.GetInt("id"),
			Name:        redemption.Name,
			Key:         key,
			CreatedTime: common.GetTimestamp(),
			Quota:       redemption.Quota,
			MaxTimes:    redemption.MaxTimes,
			ExpiredTime: redemption.ExpiredTime,
		}
		err = cleanRedemption.Insert()
		if err != nil {
			c.JSON(http.StatusOK, gin.H{
				"success": false,
				"message": err.Error(),
				"data":    keys,
			})
			return
		}
		keys = append(keys, key)
	}
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    keys,
	})
	return
}

func DeleteRedemption(c *gin.Context) {
	id, _ := strconv.Atoi(c.Param("id"))
	err := model.DeleteRedemptionById(id)
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
	})
	return
}

func UpdateRedemption(c *gin.Context) {
	statusOnly := c.Query("status_only")
	redemption := model.Redemption{}
	err := c.ShouldBindJSON(&redemption)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	cleanRedemption, err := model.GetRedemptionById(redemption.Id)
	if err != nil {
		c.JSON(http.StatusOK, gin.H{
			"success": false,
			"message": err.Error(),
		})
		return
	}
	if statusOnly != "" {
		cleanRedemption.Status = redemption.Status
	} else {
		// If you add more fields, please also update redemption.Update()
		cleanRedemption.Name = redemption.Name
		cleanRedemption.Quota = redemption.Quota
		if redemption.MaxTimes > 0 {
			cleanRedemption.MaxTimes = redemption.MaxTimes
		}
		cleanRedemption.ExpiredTime = redemption.ExpiredTime
	}
	err = cleanRedemption.Update()
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
		"data":    cleanRedemption,
	})
	return
}
