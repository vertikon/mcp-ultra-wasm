// internal/features/flags.go
package features

import (
	"math"
	"time"

	"github.com/cespare/xxhash/v2"
)

type FlagType string

const (
	Boolean    FlagType = "boolean"
	Percentage FlagType = "percentage"
	Variant    FlagType = "variant"
	Gradual    FlagType = "gradual"
)

type Flag struct {
	Key       string
	Type      FlagType
	Enabled   bool
	Metadata  map[string]string
	Variants  map[string]float64 // name->weight (0..1)
	StartDate *time.Time
	EndDate   *time.Time
}

type EvalContext struct {
	UserID     string
	Attributes map[string]any
	Timestamp  time.Time
}

type Manager interface {
	Evaluate(key string, ctx EvalContext) any
}

type InMemoryManager struct {
	flags map[string]Flag
}

func NewInMemoryManager() *InMemoryManager {
	// bootstrap some defaults
	now := time.Now()
	end := now.Add(24 * time.Hour)
	return &InMemoryManager{
		flags: map[string]Flag{
			"new_checkout":   {Key: "new_checkout", Type: Variant, Variants: map[string]float64{"control": 0.5, "treatment": 0.5}},
			"enable_cache":   {Key: "enable_cache", Type: Boolean, Enabled: true},
			"rollout_search": {Key: "rollout_search", Type: Gradual, StartDate: &now, EndDate: &end},
			"beta_ui":        {Key: "beta_ui", Type: Percentage, Metadata: map[string]string{"percentage": "25"}},
		},
	}
}

func (m *InMemoryManager) Evaluate(key string, ctx EvalContext) any {
	f, ok := m.flags[key]
	if !ok {
		return false
	}
	switch f.Type {
	case Boolean:
		return f.Enabled
	case Percentage:
		pct := parsePercent(f.Metadata["percentage"]) // 0..100
		bucket := xxhash.Sum64String(ctx.UserID+f.Key) % 100
		return bucket < uint64(pct)
	case Variant:
		// deterministic pick by weights
		norm := float64(xxhash.Sum64String(ctx.UserID+f.Key)%10000) / 10000.0
		acc := 0.0
		for name, w := range f.Variants {
			acc += w
			if norm <= acc {
				return name
			}
		}
		return "control"
	case Gradual:
		if f.StartDate == nil || f.EndDate == nil {
			return false
		}
		now := time.Now()
		if now.Before(*f.StartDate) {
			return false
		}
		if now.After(*f.EndDate) {
			return true
		}
		perc := math.Max(0, math.Min(100, 100*now.Sub(*f.StartDate).Seconds()/f.EndDate.Sub(*f.StartDate).Seconds()))
		bucket := xxhash.Sum64String(ctx.UserID+f.Key) % 100
		return float64(bucket) < perc
	default:
		return false
	}
}

func parsePercent(s string) float64 {
	var v float64
	for _, r := range s {
		if r < '0' || r > '9' {
			continue
		}
		v = v*10 + float64(r-'0')
	}
	if v < 0 {
		v = 0
	}
	if v > 100 {
		v = 100
	}
	return v
}
