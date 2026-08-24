package api

import (
	"errors"

	"github.com/lacsar712/dyevat/internal/model"
)

func classifyRecipeError(err error) (string, bool) {
	if errors.Is(err, model.ErrRecipeLevelLow) {
		return "recipe_level_low", true
	}
	return "", false
}
