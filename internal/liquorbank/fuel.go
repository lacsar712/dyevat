package liquorbank

import (
	"math"

	"github.com/lacsar712/dyevat/internal/clock"
	"github.com/lacsar712/dyevat/internal/model"
)

type LiquorRegulator struct {
	clk clock.ProcessClock
}

func NewLiquorRegulator(clk clock.ProcessClock) *LiquorRegulator {
	return &LiquorRegulator{clk: clk}
}

func (f *LiquorRegulator) IgnitionRate(settings model.PlantSettings) float64 {
	return settings.LiquorFlowTPH * 0.08
}

func (f *LiquorRegulator) ComputeForLoad(settings model.PlantSettings, loadPct float64) float64 {
	loadPct = math.Max(0, math.Min(1, loadPct))
	return settings.LiquorFlowTPH * loadPct
}

func (f *LiquorRegulator) Ramp(current, target, maxStep float64) float64 {
	delta := target - current
	if math.Abs(delta) <= maxStep {
		return target
	}
	if delta > 0 {
		return current + maxStep
	}
	return current - maxStep
}

func (f *LiquorRegulator) BtuPerHour(flowTPH float64) float64 {
	return flowTPH * 19_500_000
}

func (f *LiquorRegulator) HeatInputMW(flowTPH float64) float64 {
	return flowTPH * 11.6
}

func (f *LiquorRegulator) ValidatePermissive(settings model.PlantSettings, recipeOK, soakOK bool) error {
	if !soakOK {
		return model.ErrSoakIncomplete
	}
	if !recipeOK {
		return model.ErrRecipeLevelTrip
	}
	if settings.LiquorFlowTPH <= 0 {
		return model.ErrLiquorPermissive
	}
	return nil
}

func (f *LiquorRegulator) MinFlow(settings model.PlantSettings) float64 {
	return settings.LiquorFlowTPH * 0.2
}

func (f *LiquorRegulator) MaxFlow(settings model.PlantSettings) float64 {
	return settings.LiquorFlowTPH * 1.1
}
