package service

import (
	"math/rand"
	"sync"
	"time"

	"tea-api/model"
)

// ChannelManager implements a basic weighted round-robin scheduler for channels.
type ChannelManager struct {
	mu       sync.RWMutex
	channels []*model.Channel
	index    int
	stats    map[int]*model.ChannelStat
}

func NewChannelManager() *ChannelManager {
	cm := &ChannelManager{stats: make(map[int]*model.ChannelStat)}
	cm.refresh()
	go cm.healthLoop()
	return cm
}

func (cm *ChannelManager) refresh() {
	chs, err := model.GetAllChannels(0, -1, true, false)
	if err != nil {
		return
	}
	cm.mu.Lock()
	cm.channels = chs
	cm.index = 0
	cm.mu.Unlock()
}

func (cm *ChannelManager) next() *model.Channel {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	if len(cm.channels) == 0 {
		return nil
	}
	ch := cm.channels[cm.index]
	cm.index = (cm.index + 1) % len(cm.channels)
	return ch
}

// Select returns a channel using weighted random algorithm.
func (cm *ChannelManager) Select() *model.Channel {
	cm.mu.RLock()
	list := cm.channels
	cm.mu.RUnlock()
	if len(list) == 0 {
		return nil
	}
	total := 0
	for _, c := range list {
		total += c.GetWeight() + 1
	}
	w := rand.Intn(total)
	for _, c := range list {
		w -= c.GetWeight() + 1
		if w < 0 {
			return c
		}
	}
	return list[0]
}

// Report records usage result of a channel.
func (cm *ChannelManager) Report(channelID int, success bool) {
	_ = model.UpdateChannelStat(nil, channelID, success)
}

func (cm *ChannelManager) healthLoop() {
	ticker := time.NewTicker(time.Minute * 5)
	for range ticker.C {
		// TODO: implement real health check
		cm.refresh()
	}
}
