package app

import (
	"context"
	"fmt"

	"github.com/lacsar712/dyevat/internal/model"
)

func (a *App) WarmupStatus() (ready bool, detail string) {
	snap := a.Snapshot()
	if snap.Liquorbank.SoakStartedAt.IsZero() {
		return false, "soak not started"
	}
	if !a.soakWindow.Ready(snap.Liquorbank.SoakStartedAt) {
		return false, "soak window open"
	}
	if !snap.Liquorbank.IgnitionAt.IsZero() && !a.warmupWindow.Ready(snap.Liquorbank.IgnitionAt) {
		return false, "liquorbank warmup window open"
	}
	if !snap.Recipe.LastSwellAt.IsZero() {
		if err := a.recipe.RequireSettled(snap.Recipe); err != nil {
			return false, "recipe swell settling"
		}
	}
	return true, "ready"
}

func (a *App) WaitWarmup(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return fmt.Errorf("%w", model.ErrContextDone)
		default:
		}
		ready, _ := a.WarmupStatus()
		if ready {
			return nil
		}
	}
}

func (a *App) SoakRemaining() string {
	snap := a.Snapshot()
	if snap.Liquorbank.SoakStartedAt.IsZero() {
		return "not started"
	}
	if a.soakWindow.Ready(snap.Liquorbank.SoakStartedAt) {
		return "complete"
	}
	return "in progress"
}

func (a *App) LiquorbankWarmupRemaining() string {
	snap := a.Snapshot()
	if snap.Liquorbank.IgnitionAt.IsZero() {
		return "not ignited"
	}
	if a.warmupWindow.Ready(snap.Liquorbank.IgnitionAt) {
		return "complete"
	}
	return "in progress"
}
