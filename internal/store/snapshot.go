package store

import "github.com/lacsar712/dyevat/internal/model"

type RecipeSnapshotView struct {
	UnitID   string
	Recipe     model.RecipeReading
	Alarms   []model.AlarmEvent
	Revision uint64
}

func CloneRecipeSnapshot(s model.PlantSnapshot) RecipeSnapshotView {
	out := RecipeSnapshotView{
		UnitID:   s.UnitID,
		Recipe:     s.Recipe,
		Revision: s.Revision,
	}
	out.Alarms = make([]model.AlarmEvent, len(s.Alarms))
	copy(out.Alarms, s.Alarms)
	return out
}
