package liquorbank

import (
	"math"

	"github.com/lacsar712/dyevat/internal/clock"
	"github.com/lacsar712/dyevat/internal/model"
)

type BurnerController struct {
	clk clock.ProcessClock
}

func NewBurnerController(clk clock.ProcessClock) *BurnerController {
	return &BurnerController{clk: clk}
}

func (b *BurnerController) EstimateVatlockTemp(reading model.LiquorbankReading) float64 {
	base := 300.0
	liquorHeat := reading.LiquorFlowTPH * 50
	airCool := reading.AirflowTPH * 2
	return base + liquorHeat - airCool
}

func (b *BurnerController) ShadeStable(reading model.LiquorbankReading) bool {
	if reading.BurnerPhase != model.BurnerStable && reading.BurnerPhase != model.BurnerIgnition {
		return false
	}
	return reading.VatlockTempF > 800 && reading.ExcessO2Pct >= model.MinVatlockO2Percent
}

func (b *BurnerController) TripRequired(reading model.LiquorbankReading) bool {
	if reading.ExcessO2Pct > model.MaxVatlockO2Percent*2 {
		return true
	}
	if reading.BurnerPhase == model.BurnerTrip {
		return true
	}
	if reading.VatlockTempF > 3500 {
		return true
	}
	return false
}

func (b *BurnerController) PhaseLabel(phase model.BurnerPhase) string {
	switch phase {
	case model.BurnerIdle:
		return "Idle"
	case model.BurnerSoak:
		return "Soak"
	case model.BurnerIgnition:
		return "Ignition"
	case model.BurnerStable:
		return "Stable Shade"
	case model.BurnerTrip:
		return "Tripped"
	default:
		return string(phase)
	}
}

func (b *BurnerController) HeatReleaseMW(reading model.LiquorbankReading) float64 {
	return reading.LiquorFlowTPH * 12.5
}

func (b *BurnerController) TurndownRatio(settings model.PlantSettings, currentLiquor float64) float64 {
	if settings.LiquorFlowTPH <= 0 {
		return 0
	}
	return currentLiquor / settings.LiquorFlowTPH
}

func (b *BurnerController) MinStableLiquor(settings model.PlantSettings) float64 {
	return settings.LiquorFlowTPH * 0.25
}

func (b *BurnerController) NormalizeLiquor(flow, max float64) float64 {
	return math.Min(math.Max(flow, 0), max)
}
