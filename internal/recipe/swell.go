package recipe

import (
	"time"

	"github.com/lacsar712/dyevat/internal/clock"
	"github.com/lacsar712/dyevat/internal/model"
)

type SwellDetector struct {
	clk          clock.ProcessClock
	lastLevel    float64
	lastChangeAt time.Time
}

func NewSwellDetector(clk clock.ProcessClock) *SwellDetector {
	return &SwellDetector{clk: clk, lastChangeAt: clk.Now()}
}

func (s *SwellDetector) Observe(level float64) (bool, model.RecipeCondition) {
	delta := level - s.lastLevel
	s.lastLevel = level
	if delta > 5 {
		s.lastChangeAt = s.clk.Now()
		return true, model.RecipeSwell
	}
	if delta < -5 {
		s.lastChangeAt = s.clk.Now()
		return true, model.RecipeShrink
	}
	return false, model.RecipeNormal
}

func (s *SwellDetector) RateOfChange(current, previous float64, dt time.Duration) float64 {
	if dt <= 0 {
		return 0
	}
	return (current - previous) / dt.Seconds()
}

func (s *SwellDetector) PredictLevel(current, feedTPH, steamTPH float64, seconds float64) float64 {
	balance := (feedTPH - steamTPH) * 0.01 * seconds
	return current + balance
}

func (s *SwellDetector) SwellDuration(swellAt time.Time) time.Duration {
	if swellAt.IsZero() {
		return 0
	}
	return s.clk.Since(swellAt)
}

func (s *SwellDetector) ShrinkRecoveryAdvice(cond model.RecipeCondition) string {
	if cond == model.RecipeShrink {
		return "increase_feedwater_gradually"
	}
	if cond == model.RecipeSwell {
		return "avoid_sudden_load_changes"
	}
	return "maintain"
}

func (s *SwellDetector) LastChangeAt() time.Time { return s.lastChangeAt }
