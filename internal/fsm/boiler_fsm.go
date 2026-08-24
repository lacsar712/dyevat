package fsm

import (
	"context"
	"fmt"
	"sync"

	"github.com/lacsar712/dyevat/internal/model"
)

type VatlineFSM struct {
	mu            sync.RWMutex
	state         model.PlantState
	liquorPermissive bool
	soakComplete  bool
	hooks          *HookChain
}

func NewVatlineFSM(unitID string) *VatlineFSM {
	_ = unitID
	return &VatlineFSM{state: model.StateColdStandby, hooks: NewHookChain()}
}

func (f *VatlineFSM) Hooks() *HookChain { return f.hooks }

func (f *VatlineFSM) State() model.PlantState {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.state
}

func (f *VatlineFSM) SetLiquorPermissive(ok bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.liquorPermissive = ok
}

func (f *VatlineFSM) SetSoakComplete(ok bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.soakComplete = ok
}

func (f *VatlineFSM) LiquorPermissive() bool {
	f.mu.RLock()
	defer f.mu.RUnlock()
	return f.liquorPermissive
}

func (f *VatlineFSM) Dispatch(ctx context.Context, event PlantEvent) (model.PlantState, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	select {
	case <-ctx.Done():
		return f.state, fmt.Errorf("%w", model.ErrContextDone)
	default:
	}
	if event == EvTrip {
		from := f.state
		if f.hooks != nil {
			if err := f.hooks.RunBefore(ctx, from, model.StateTrip, event); err != nil {
				return f.state, err
			}
		}
		f.state = model.StateTrip
		if f.hooks != nil {
			if err := f.hooks.RunAfter(ctx, from, model.StateTrip, event); err != nil {
				return f.state, err
			}
		}
		return f.state, nil
	}
	next, ok := NextState(f.state, event)
	if !ok {
		return f.state, fmt.Errorf("%s from %s: %w", event, f.state, ErrIllegalTransition)
	}
	if event == EvIgnite && !f.liquorPermissive {
		return f.state, fmt.Errorf("%w", model.ErrLiquorPermissive)
	}
	if event == EvSoakComplete && !f.soakComplete {
		return f.state, fmt.Errorf("%w", model.ErrSoakIncomplete)
	}
	from := f.state
	if f.hooks != nil {
		if err := f.hooks.RunBefore(ctx, from, next, event); err != nil {
			return f.state, err
		}
	}
	f.state = next
	if f.hooks != nil {
		if err := f.hooks.RunAfter(ctx, from, next, event); err != nil {
			return f.state, err
		}
	}
	return f.state, nil
}

func (f *VatlineFSM) ForceState(state model.PlantState) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.state = state
}
