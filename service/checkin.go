package service

import (
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
	return continuous, nil
}
