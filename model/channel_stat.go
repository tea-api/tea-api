package model

import "gorm.io/gorm"

// ChannelStat records usage statistics for a channel
// SuccessRate is calculated as Success/Total in service layer
// Continuous failures or successes can be extended later

type ChannelStat struct {
	ID        int   `json:"id" gorm:"primaryKey"`
	ChannelID int   `json:"channel_id" gorm:"index"`
	Total     int64 `json:"total"`
	Success   int64 `json:"success"`
	UpdatedAt int64 `json:"updated_at" gorm:"autoUpdateTime:milli"`
	CreatedAt int64 `json:"created_at" gorm:"autoCreateTime:milli"`
}

type ChannelStatDetail struct {
    ChannelStat
    Name string `json:"name"`
}

func GetChannelStatDetails() (stats []ChannelStatDetail, err error) {
    err = DB.Table("channel_stats").
        Select("channel_stats.*, channels.name").
        Joins("left join channels on channels.id = channel_stats.channel_id").
        Scan(&stats).Error
    return
}

func UpdateChannelStat(tx *gorm.DB, channelID int, success bool) error {
	if tx == nil {
		tx = DB
	}
	stat := ChannelStat{ChannelID: channelID}
	err := tx.FirstOrCreate(&stat, ChannelStat{ChannelID: channelID}).Error
	if err != nil {
		return err
	}
	updates := map[string]interface{}{"total": gorm.Expr("total + ?", 1)}
	if success {
		updates["success"] = gorm.Expr("success + ?", 1)
	}
	return tx.Model(&ChannelStat{}).Where("channel_id = ?", channelID).Updates(updates).Error
}
