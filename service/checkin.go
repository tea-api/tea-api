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

	// 处理连续签到计数
	if last != nil {
		// 检查是否是连续签到
		lastDate, _ := time.Parse("2006-01-02", last.Date)
		todayDate, _ := time.Parse("2006-01-02", today)
		daysDiff := int(todayDate.Sub(lastDate).Hours() / 24)

		if daysDiff == 1 {
			// 连续签到，增加连续天数
			continuous = last.Continuous + 1
		} else if daysDiff > 1 && !common.CheckinStreakReset {
			// 如果配置了不重置连续签到，则保持原有连续天数
			continuous = last.Continuous
		}
		// 其他情况保持默认值 continuous = 1 (重置签到计数)
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
	reward, _ := calculateReward(continuous)
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
// 返回总奖励和是否命中特殊奖励
func calculateReward(continuous int) (int, bool) {
	// 基础奖励
	reward := common.BaseCheckinReward
	hitSpecial := false

	// 连续签到额外奖励
	if continuous > 1 {
		// 计算额外奖励天数，但不超过上限
		extraDays := continuous - 1
		if extraDays > common.MaxContinuousRewardDays {
			extraDays = common.MaxContinuousRewardDays
		}
		reward += extraDays * common.ContinuousCheckinReward
	}

	// 检查是否命中特殊奖励日
	for i, day := range common.SpecialRewardDays {
		if continuous == day && i < len(common.SpecialRewards) {
			reward += common.SpecialRewards[i]
			hitSpecial = true
			break
		}
	}

	return reward, hitSpecial
}

// GetCheckinRewardInfo 获取签到奖励信息
func GetCheckinRewardInfo(continuous int) map[string]interface{} {
	reward, isSpecial := calculateReward(continuous)

	return map[string]interface{}{
		"reward":              reward,
		"is_special_reward":   isSpecial,
		"base_reward":         common.BaseCheckinReward,
		"continuous_reward":   common.ContinuousCheckinReward,
		"max_continuous_days": common.MaxContinuousRewardDays,
		"special_days":        common.SpecialRewardDays,
		"special_rewards":     common.SpecialRewards,
		"streak_reset":        common.CheckinStreakReset,
	}
}
