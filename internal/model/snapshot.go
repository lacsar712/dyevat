package model

import "time"

func CloneSnapshot(s PlantSnapshot) PlantSnapshot {
	out := s
	out.Alarms = append([]AlarmEvent(nil), s.Alarms...)
	return out
}

func DefaultSnapshot(unitID string) PlantSnapshot {
	now := time.Now()
	return PlantSnapshot{
		UnitID: unitID,
		State:  StateColdStandby,
		Settings: PlantSettings{
			Mode:              ModeBaseLoad,
			TargetMW:          150,
			TargetSteamPSI:    NormalSteamPressurePSI,
			RecipeLevelSetpoint: 55,
			FeedwaterFlowTPH:  400,
			LiquorFlowTPH:       35,
			ExcessO2Setpoint:  3.5,
		},
		Plant: PlantRef{UnitLabel: unitID, PlantCode: "STEAM-PLT"},
		Recipe: RecipeReading{
			LevelPercent: 50,
			Condition:    RecipeNormal,
			FeedwaterTPH: 0,
			SteamFlowTPH: 0,
		},
		Liquorbank: LiquorbankReading{
			BurnerPhase: BurnerIdle,
		},
		Vatline: VatlineReading{
			SteamPressurePSI: 0,
			SteamTempF:       70,
		},
		UpdatedAt: now,
	}
}

func (s PlantSnapshot) IsFiring() bool {
	return s.State == StateFiring || s.State == StateLoadFollow || s.State == StateRamp
}

func (s PlantSnapshot) RecipeWithinLimits() bool {
	return s.Recipe.LevelPercent >= MinRecipeLevelPercent && s.Recipe.LevelPercent <= MaxRecipeLevelPercent
}

func (s PlantSnapshot) PressureWithinLimits() bool {
	if !s.IsFiring() {
		return true
	}
	return s.Vatline.SteamPressurePSI <= MaxSteamPressurePSI
}
