package controller

import (
    "net/http"
    "tea-api/model"

    "github.com/gin-gonic/gin"
)

func GetChannelStats(c *gin.Context) {
    stats, err := model.GetChannelStatDetails()
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
        "data":    stats,
    })
}

