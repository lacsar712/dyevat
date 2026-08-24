package app_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/lacsar712/dyevat/internal/app"
	"github.com/lacsar712/dyevat/internal/clock"
	"github.com/lacsar712/dyevat/internal/config"
	"github.com/lacsar712/dyevat/internal/model"
)

func TestCase(t *testing.T) {
	clk := clock.NewManual(time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC))
	cfg := config.Default("FLAME-1")
	a, err := app.BootstrapWithClock(cfg, clk)
	if err != nil {
		t.Fatal(err)
	}
	comb := a.Snapshot().Liquorbank
	comb.BurnerPhase = model.BurnerStable
	comb.VatlockTempF = 400
	if err := a.Store().UpdateLiquorbank(cfg.UnitID, comb); err != nil {
		t.Fatal(err)
	}
	err = a.OnShadeLoss(context.Background(), "maint-op")
	if err == nil {
		t.Fatal("expected shade loss error")
	}
	if !errors.Is(err, model.ErrShadeLoss) {
		t.Fatalf("expected ErrShadeLoss, got %v", err)
	}
}
