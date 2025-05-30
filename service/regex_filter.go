package service

import (
	"regexp"
	"sync"
)

// RegexRule represents a filter rule.
type RegexRule struct {
	Pattern *regexp.Regexp
	Group   string
	Raw     string
}

type RegexFilter struct {
	mu    sync.RWMutex
	rules []RegexRule
}

func NewRegexFilter() *RegexFilter {
	return &RegexFilter{}
}

func (f *RegexFilter) AddRule(group, expr string) error {
	r, err := regexp.Compile(expr)
	if err != nil {
		return err
	}
	f.mu.Lock()
	f.rules = append(f.rules, RegexRule{Pattern: r, Group: group, Raw: expr})
	f.mu.Unlock()
	return nil
}

// Check returns true if text matches any rule.
func (f *RegexFilter) Check(text string) (bool, string) {
	f.mu.RLock()
	defer f.mu.RUnlock()
	for _, rule := range f.rules {
		if rule.Pattern.MatchString(text) {
			return true, rule.Group
		}
	}
	return false, ""
}
