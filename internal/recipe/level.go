package recipe

import (
	"math"

	"github.com/lacsar712/dyevat/internal/clock"
	"github.com/lacsar712/dyevat/internal/model"
)

type LevelController struct {
	clk clock.ProcessClock
}

func NewLevelController(clk clock.ProcessClock) *LevelController {
	return &LevelController{clk: clk}
}

func (l *LevelController) Compute(snap model.PlantSnapshot, firing bool) (float64, model.RecipeCondition) {
	level := snap.Recipe.LevelPercent
	if !firing {
		return level, model.RecipeNormal
	}
	balance := snap.Recipe.FeedwaterTPH - snap.Recipe.SteamFlowTPH
	level += balance * 0.01
	level = math.Max(model.MinRecipeLevelPercent, math.Min(model.MaxRecipeLevelPercent, level))
	cond := l.classify(level, snap)
	return level, cond
}

func (l *LevelController) classify(level float64, snap model.PlantSnapshot) model.RecipeCondition {
	setpoint := snap.Settings.RecipeLevelSetpoint
	if level > setpoint+15 {
		return model.RecipeSwell
	}
	if level < setpoint-15 {
		return model.RecipeShrink
	}
	if snap.Vatline.SteamPressurePSI > snap.Settings.TargetSteamPSI*0.9 && level > setpoint+5 {
		return model.RecipeCarry
	}
	return model.RecipeNormal
}

func (l *LevelController) RecommendFeedwater(snap model.PlantSnapshot, firing bool) float64 {
	if !firing {
		return 0
	}
	err := snap.Settings.RecipeLevelSetpoint - snap.Recipe.LevelPercent
	return snap.Settings.FeedwaterFlowTPH + err*3
}

func (l *LevelController) WithinLimits(level float64) bool {
	return level >= model.MinRecipeLevelPercent && level <= model.MaxRecipeLevelPercent
}

func (l *LevelController) TripLow(level float64) bool  { return level < model.TripRecipeLowPercent }
func (l *LevelController) TripHigh(level float64) bool { return level > model.TripRecipeHighPercent }

func (l *LevelController) LevelError(snap model.PlantSnapshot) float64 {
	return snap.Recipe.LevelPercent - snap.Settings.RecipeLevelSetpoint
}

func (l *LevelController) ThreeElementBias(snap model.PlantSnapshot) float64 {
	steam := snap.Recipe.SteamFlowTPH
	feed := snap.Recipe.FeedwaterTPH
	levelErr := l.LevelError(snap)
	return feed + (steam-feed)*0.5 + levelErr*2
}
