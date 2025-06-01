package system_setting

import "tea-api/setting/config"

// LinuxDOTrustLevelSettings 存储不同L站账户信任等级注册时送的额度设置
type LinuxDOTrustLevelSettings struct {
	Enabled     bool `json:"enabled"`
	TrustLevel0 int  `json:"trust_level_0"`
	TrustLevel1 int  `json:"trust_level_1"`
	TrustLevel2 int  `json:"trust_level_2"`
	TrustLevel3 int  `json:"trust_level_3"`
	TrustLevel4 int  `json:"trust_level_4"`
}

// 默认配置
var defaultLinuxDOTrustLevelSettings = LinuxDOTrustLevelSettings{
	Enabled:     false,
	TrustLevel0: 0,
	TrustLevel1: 1000,
	TrustLevel2: 2000,
	TrustLevel3: 5000,
	TrustLevel4: 10000,
}

func init() {
	// 注册到全局配置管理器
	config.GlobalConfig.Register("linuxdo_trust_level", &defaultLinuxDOTrustLevelSettings)
}

func GetLinuxDOTrustLevelSettings() *LinuxDOTrustLevelSettings {
	return &defaultLinuxDOTrustLevelSettings
}
