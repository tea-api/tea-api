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
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "",
		"data":    gin.H{"continuous": days},
	})
}
