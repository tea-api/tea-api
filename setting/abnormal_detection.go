package setting

import "tea-api/setting/config"

type RateSpikeRule struct {
	Enabled           bool `json:"enabled"`
	WindowSeconds     int  `json:"window_seconds"`
	SpikePercent      int  `json:"spike_percent"`
	NeighborDiffRatio int  `json:"neighbor_diff_ratio"`
}

type HighFrequencyRule struct {
	Enabled              bool `json:"enabled"`
	MaxRequestsPerSecond int  `json:"max_requests_per_second"`
}

type SecurityPolicy struct {
	SleepSeconds int `json:"sleep_seconds"`
}

type AbnormalRules struct {
	RateSpike     RateSpikeRule     `json:"rate_spike"`
	HighFrequency HighFrequencyRule `json:"high_frequency"`
}

type AbnormalDetectionConfig struct {
	Enabled  bool           `json:"enabled"`
	Rules    AbnormalRules  `json:"rules"`
	Security SecurityPolicy `json:"security"`
}

var defaultAbnormalDetection = AbnormalDetectionConfig{
	Enabled: false,
	Rules: AbnormalRules{
		RateSpike:     RateSpikeRule{Enabled: false, WindowSeconds: 60, SpikePercent: 500, NeighborDiffRatio: 10},
		HighFrequency: HighFrequencyRule{Enabled: false, MaxRequestsPerSecond: 50},
	},
	Security: SecurityPolicy{SleepSeconds: 0},
}

func init() {
	config.GlobalConfig.Register("abnormal_detection", &defaultAbnormalDetection)
}

func GetAbnormalDetection() *AbnormalDetectionConfig {
	return &defaultAbnormalDetection
}
