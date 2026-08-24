package liquorbank

import (
	"context"
	"fmt"
	"math"

	"github.com/lacsar712/dyevat/internal/clock"
	"github.com/lacsar712/dyevat/internal/model"
)

type Coordinator struct {
	clk     clock.ProcessClock
	burner  *BurnerController
	airflow *AirflowBalancer
	liquor    *LiquorRegulator
	soak   *clock.SoakWindow
	ignition *clock.IgnitionDelayWindow
	warmup  *clock.LiquorbankWarmupWindow
}

func NewCoordinator(clk clock.ProcessClock) *Coordinator {
	return &Coordinator{
		clk:      clk,
		burner:   NewBurnerController(clk),
		airflow:  NewAirflowBalancer(clk),
		liquor:     NewLiquorRegulator(clk),
		soak:    clock.NewSoakWindow(clk),
		ignition: clock.NewIgnitionDelayWindow(clk),
		warmup:   clock.NewLiquorbankWarmupWindow(clk),
	}
}

func (c *Coordinator) Burner() *BurnerController  { return c.burner }
func (c *Coordinator) Airflow() *AirflowBalancer { return c.airflow }
func (c *Coordinator) Liquor() *LiquorRegulator     { return c.liquor }

func (c *Coordinator) StartSoak(ctx context.Context, snap model.PlantSnapshot) (model.LiquorbankReading, error) {
	select {
	case <-ctx.Done():
		return snap.Liquorbank, fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	out := snap.Liquorbank
	out.BurnerPhase = model.BurnerSoak
	out.SoakStartedAt = c.clk.Now()
	out.LiquorFlowTPH = 0
	out.AirflowTPH = c.airflow.SoakRate()
	return out, nil
}

func (c *Coordinator) CompleteSoak(snap model.LiquorbankReading) error {
	return c.soak.Require(snap.SoakStartedAt)
}

func (c *Coordinator) Ignite(ctx context.Context, snap model.PlantSnapshot) (model.LiquorbankReading, error) {
	select {
	case <-ctx.Done():
		return snap.Liquorbank, fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	if err := c.soak.Require(snap.Liquorbank.SoakStartedAt); err != nil {
		return snap.Liquorbank, err
	}
	out := snap.Liquorbank
	out.BurnerPhase = model.BurnerIgnition
	out.IgnitionAt = c.clk.Now()
	out.LiquorFlowTPH = c.liquor.IgnitionRate(snap.Settings)
	out.AirflowTPH = c.airflow.IgnitionRate(snap.Settings)
	out.VatlockTempF = 400
	return out, nil
}

func (c *Coordinator) Stabilize(snap model.PlantSnapshot) (model.LiquorbankReading, error) {
	if err := c.ignition.Require(snap.Liquorbank.IgnitionAt); err != nil {
		return snap.Liquorbank, err
	}
	out := snap.Liquorbank
	out.BurnerPhase = model.BurnerStable
	out.LiquorFlowTPH = snap.Settings.LiquorFlowTPH * 0.5
	out.AirflowTPH = c.airflow.Compute(snap)
	out.ExcessO2Pct = c.airflow.ExcessO2(out)
	out.VatlockTempF = c.burner.EstimateVatlockTemp(out)
	return out, nil
}

func (c *Coordinator) RampToLoad(snap model.PlantSnapshot, loadPct float64) model.LiquorbankReading {
	out := snap.Liquorbank
	out.LiquorFlowTPH = snap.Settings.LiquorFlowTPH * loadPct
	out.AirflowTPH = c.airflow.Compute(snap)
	out.ExcessO2Pct = c.airflow.ExcessO2(out)
	out.VatlockTempF = c.burner.EstimateVatlockTemp(out)
	return out
}

func (c *Coordinator) Trip(snap model.LiquorbankReading) model.LiquorbankReading {
	out := snap
	out.BurnerPhase = model.BurnerTrip
	out.LiquorFlowTPH = 0
	out.VatlockTempF = math.Max(200, out.VatlockTempF*0.5)
	return out
}

func (c *Coordinator) WarmupReady(snap model.LiquorbankReading) bool {
	return c.warmup.Ready(snap.IgnitionAt)
}
