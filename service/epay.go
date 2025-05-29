package service

import (
	"tea-api/setting"
)

func GetCallbackAddress() string {
	if setting.CustomCallbackAddress == "" {
		return setting.ServerAddress
	}
	return setting.CustomCallbackAddress
}
