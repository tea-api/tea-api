package model

// CheckinRecord stores daily sign-in information.
type CheckinRecord struct {
	Id          int    `json:"id" gorm:"primaryKey"`
	UserId      int    `json:"user_id" gorm:"index"`
	Date        string `json:"date" gorm:"index"`
	Continuous  int    `json:"continuous"`
	CreatedTime int64  `json:"created_time" gorm:"autoCreateTime:milli"`
}

func CreateCheckin(record *CheckinRecord) error {
	return DB.Create(record).Error
}

func GetLastCheckin(userId int) (*CheckinRecord, error) {
	var rec CheckinRecord
	err := DB.Where("user_id = ?", userId).Order("date desc").First(&rec).Error
	return &rec, err
}
