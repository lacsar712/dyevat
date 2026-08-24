package app

import (
	"fmt"

	"github.com/lacsar712/dyevat/internal/model"
)

func (a *App) CheckRecipeLevel(snap model.PlantSnapshot) error {
	if snap.Recipe.LevelPercent < model.MinRecipeLevelPercent {
		return fmt.Errorf("%w", model.ErrRecipeLevelLow)
	}
	return nil
}
