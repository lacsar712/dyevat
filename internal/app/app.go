package app

import (
	"context"
	"fmt"
	"sync"

	"github.com/lacsar712/dyevat/internal/vatline"
	"github.com/lacsar712/dyevat/internal/clock"
	"github.com/lacsar712/dyevat/internal/liquorbank"
	"github.com/lacsar712/dyevat/internal/config"
	"github.com/lacsar712/dyevat/internal/recipe"
	"github.com/lacsar712/dyevat/internal/fsm"
	"github.com/lacsar712/dyevat/internal/interlock"
	"github.com/lacsar712/dyevat/internal/model"
	"github.com/lacsar712/dyevat/internal/store"
)

type App struct {
	cfg           config.Config
	clk           clock.ProcessClock
	store         *store.PlantStore
	journal       *store.Journal
	fsm           *fsm.VatlineFSM
	vatline        *vatline.Controller
	liquorbank    *liquorbank.Coordinator
	recipe          *recipe.Coordinator
	interlock     *interlock.Interlock
	permissives   *interlock.PermissiveSet
	coordLock     *interlock.CoordinationLock
	scheduler     *clock.Scheduler
	soakWindow   *clock.SoakWindow
	warmupWindow  *clock.LiquorbankWarmupWindow
	telemetry     *Telemetry
	tickCancels    map[string]context.CancelFunc
	liquorLoopCancels map[string]context.CancelFunc
	mu             sync.RWMutex
}

func New(cfg config.Config, clk clock.ProcessClock) *App {
	return &App{
		cfg:          cfg,
		clk:          clk,
		store:        store.NewPlantStore(),
		journal:      store.NewJournal(cfg.JournalPath, cfg.JournalCapacity),
		fsm:          fsm.NewVatlineFSM(cfg.UnitID),
		vatline:       vatline.NewController(clk),
		liquorbank:   liquorbank.NewCoordinator(clk),
		recipe:         recipe.NewCoordinator(clk),
		interlock:    interlock.NewInterlock(cfg.LeaseTTL),
		permissives:  interlock.NewPermissiveSet(),
		coordLock:    interlock.NewCoordinationLock(),
		scheduler:    clock.NewScheduler(clk),
		soakWindow:  clock.NewSoakWindow(clk),
		warmupWindow: clock.NewLiquorbankWarmupWindow(clk),
		telemetry:    NewTelemetry(cfg.UnitID),
		tickCancels:     make(map[string]context.CancelFunc),
		liquorLoopCancels: make(map[string]context.CancelFunc),
	}
}

func (a *App) Snapshot() model.PlantSnapshot {
	snap, err := a.store.Require(a.cfg.UnitID)
	if err != nil {
		return model.DefaultSnapshot(a.cfg.UnitID)
	}
	return snap
}

func (a *App) Config() config.Config              { return a.cfg }
func (a *App) Clock() clock.ProcessClock          { return a.clk }
func (a *App) FSM() *fsm.VatlineFSM                { return a.fsm }
func (a *App) UnitID() string                     { return a.cfg.UnitID }
func (a *App) Store() *store.PlantStore           { return a.store }
func (a *App) Interlock() *interlock.Interlock    { return a.interlock }
func (a *App) Telemetry() TelemetrySnapshot       { return a.telemetry.Snapshot() }
func (a *App) Journal() *store.Journal            { return a.journal }

func (a *App) journalEvent(ev, payload string) {
	_, _ = a.journal.Append(a.cfg.UnitID, ev, payload)
}

func (a *App) syncState(state model.PlantState) {
	_ = a.store.UpdateState(a.cfg.UnitID, state)
}

func (a *App) isFiring(state model.PlantState) bool {
	return state == model.StateFiring || state == model.StateLoadFollow || state == model.StateRamp
}

func (a *App) refreshPermissives(snap model.PlantSnapshot) {
	a.permissives.SetRecipe(a.recipe.Level().WithinLimits(snap.Recipe.LevelPercent))
	a.permissives.SetPressure(a.vatline.Pressure().WithinTripLimits(snap.Vatline.SteamPressurePSI, a.isFiring(snap.State)))
	a.permissives.SetLiquorbank(a.liquorbank.Burner().ShadeStable(snap.Liquorbank))
	a.permissives.SetLiquor(snap.Liquorbank.LiquorFlowTPH > 0 || snap.State == model.StateSoak)
	a.permissives.SetIgnition(snap.Liquorbank.BurnerPhase == model.BurnerStable || snap.Liquorbank.BurnerPhase == model.BurnerIgnition)
	a.fsm.SetLiquorPermissive(a.permissives.LiquorOK())
	a.fsm.SetSoakComplete(a.soakWindow.Ready(snap.Liquorbank.SoakStartedAt))
}

func (a *App) tickLabel() string {
	return fmt.Sprintf("%s-tick", a.cfg.UnitID)
}
