package service

import (
	"tea-api/common"
	"tea-api/model"
	"time"
)

// Checkin handles daily check-in logic.
// It will create a new record if the user has not checked in today and return
// the current continuous days.
func Checkin(userId int) (int, error) {
	today := time.Now().Format("2006-01-02")
	last, err := model.GetLastCheckin(userId)
	if err != nil && err.Error() != "record not found" {
		return 0, err
	}
	continuous := 1
	if last != nil && last.Date == today {
		return last.Continuous, nil
	}
	if last != nil && last.Date == time.Now().AddDate(0, 0, -1).Format("2006-01-02") {
		continuous = last.Continuous + 1
	}
	rec := &model.CheckinRecord{
		UserId:     userId,
		Date:       today,
		Continuous: continuous,
	}
	if err := model.CreateCheckin(rec); err != nil {
		return 0, err
	}

	// 计算并发放奖励
	reward := calculateReward(continuous)
	if reward > 0 {
		// 增加用户配额
		if err := model.IncreaseUserQuota(userId, reward, true); err != nil {
			common.SysError("签到奖励发放失败: " + err.Error())
			// 但不影响签到功能本身
		}
	}

	return continuous, nil
}

// GetCheckinStatus 获取用户当天是否已签到
// 返回值：是否今天已签到，连续签到天数，错误
func GetCheckinStatus(userId int) (bool, int, error) {
	today := time.Now().Format("2006-01-02")
	last, err := model.GetLastCheckin(userId)
	if err != nil {
		if err.Error() == "record not found" {
			// 用户从未签到过
			return false, 0, nil
		}
		return false, 0, err
	}

	// 检查是否是今天签到的
	checkedToday := last.Date == today
	return checkedToday, last.Continuous, nil
}

// calculateReward 计算签到奖励
func calculateReward(continuous int) int {
	// 基础奖励
	reward := common.BaseCheckinReward

	// 连续签到额外奖励
	if continuous > 1 {
		// 计算额外奖励天数，但不超过上限
		extraDays := continuous - 1
		if extraDays > common.MaxContinuousRewardDays {
			extraDays = common.MaxContinuousRewardDays
		}
		reward += extraDays * common.ContinuousCheckinReward
	}

	return reward
}
