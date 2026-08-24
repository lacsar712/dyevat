package model

import "time"

const (
	DefaultLeaseTTL        = 30 * time.Second
	SoakWindow            = 5 * time.Minute
	IgnitionDelayWindow    = 15 * time.Second
	RecipeSwellSettleWindow  = 45 * time.Second
	LiquorbankWarmupWindow = 2 * time.Minute
	FeedwaterRampWindow    = 30 * time.Second
	MaxRecipeLevelPercent    = 95.0
	MinRecipeLevelPercent    = 15.0
	TripRecipeLowPercent     = 10.0
	TripRecipeHighPercent    = 98.0
	NormalSteamPressurePSI = 1800.0
	MaxSteamPressurePSI    = 2000.0
	MinVatlockO2Percent    = 2.5
	MaxVatlockO2Percent    = 6.0
	DefaultJournalCapacity = 512
)
