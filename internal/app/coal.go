package app

import (
	"context"
	"fmt"
	"time"

	"github.com/lacsar712/dyevat/internal/clock"
	"github.com/lacsar712/dyevat/internal/model"
)

func (a *App) advanceClock(d time.Duration) {
	if mc, ok := a.clk.(*clock.ManualClock); ok {
		mc.Advance(d)
		time.Sleep(time.Millisecond)
	} else {
		time.Sleep(d)
	}
}

func (a *App) bindLiquorLoop(holder string, ctx context.Context) context.Context {
	a.mu.Lock()
	if cancel, ok := a.liquorLoopCancels[holder]; ok {
		cancel()
	}
	child, cancel := context.WithCancel(ctx)
	a.liquorLoopCancels[holder] = cancel
	a.mu.Unlock()
	return child
}

func (a *App) cancelLiquorLoop(holder string) {
	a.mu.Lock()
	if cancel, ok := a.liquorLoopCancels[holder]; ok {
		cancel()
		delete(a.liquorLoopCancels, holder)
	}
	a.mu.Unlock()
}

func (a *App) cancelAllLiquorLoops() {
	a.mu.Lock()
	for holder, cancel := range a.liquorLoopCancels {
		cancel()
		delete(a.liquorLoopCancels, holder)
	}
	a.mu.Unlock()
}

func (a *App) CoalFeedTPH() float64 {
	return a.Snapshot().Liquorbank.LiquorFlowTPH
}

func (a *App) RunLiquorRamp(ctx context.Context, holder string, targetTPH float64) error {
	loopCtx := a.bindLiquorLoop(holder, ctx)
	defer a.cancelLiquorLoop(holder)
	for {
		if err := loopCtx.Err(); err != nil {
			return fmt.Errorf("%w", model.ErrContextDone)
		}
		snap := a.Snapshot()
		current := snap.Liquorbank.LiquorFlowTPH
		if current >= targetTPH {
			return nil
		}
		comb := snap.Liquorbank
		comb.LiquorFlowTPH = current + 1.0
		_ = a.store.UpdateLiquorbank(a.cfg.UnitID, comb)
		a.telemetry.RecordCoalFeed(comb.LiquorFlowTPH)
		a.advanceClock(100 * time.Millisecond)
	}
}

func (a *App) RunCoalFeed(ctx context.Context, holder string, steps int) error {
	loopCtx := a.bindLiquorLoop(holder, ctx)
	defer a.cancelLiquorLoop(holder)
	for i := 0; steps <= 0 || i < steps; i++ {
		if err := loopCtx.Err(); err != nil {
			return fmt.Errorf("%w", model.ErrContextDone)
		}
		snap := a.Snapshot()
		comb := snap.Liquorbank
		comb.LiquorFlowTPH += 0.5
		_ = a.store.UpdateLiquorbank(a.cfg.UnitID, comb)
		a.telemetry.RecordCoalFeed(comb.LiquorFlowTPH)
		a.advanceClock(100 * time.Millisecond)
	}
	return nil
}
