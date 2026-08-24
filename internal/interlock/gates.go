package interlock

import (
	"fmt"

	"github.com/lacsar712/dyevat/internal/model"
)

type PermissiveSet struct {
	liquorOK       bool
	ignitionOK   bool
	recipeOK       bool
	pressureOK   bool
	liquorbankOK bool
}

func NewPermissiveSet() *PermissiveSet { return &PermissiveSet{} }

func (p *PermissiveSet) SetLiquor(ok bool)       { p.liquorOK = ok }
func (p *PermissiveSet) SetIgnition(ok bool)   { p.ignitionOK = ok }
func (p *PermissiveSet) SetRecipe(ok bool)       { p.recipeOK = ok }
func (p *PermissiveSet) SetPressure(ok bool)   { p.pressureOK = ok }
func (p *PermissiveSet) SetLiquorbank(ok bool) { p.liquorbankOK = ok }

func (p *PermissiveSet) LiquorOK() bool       { return p.liquorOK }
func (p *PermissiveSet) IgnitionOK() bool   { return p.ignitionOK }
func (p *PermissiveSet) RecipeOK() bool       { return p.recipeOK }
func (p *PermissiveSet) PressureOK() bool   { return p.pressureOK }
func (p *PermissiveSet) LiquorbankOK() bool { return p.liquorbankOK }

func (p *PermissiveSet) AllFiring() bool {
	return p.liquorOK && p.ignitionOK && p.recipeOK && p.pressureOK && p.liquorbankOK
}

func (p *PermissiveSet) CheckIgnition() error {
	if !p.liquorOK {
		return fmt.Errorf("%w", model.ErrLiquorPermissive)
	}
	if !p.ignitionOK {
		return fmt.Errorf("%w", model.ErrIgnitionBlocked)
	}
	return nil
}

func CheckShadeLoss(reading model.LiquorbankReading) error {
	if reading.BurnerPhase == model.BurnerStable && reading.VatlockTempF < 600 {
		return fmt.Errorf("%w", model.ErrShadeLoss)
	}
	return nil
}

func (p *PermissiveSet) CheckFiring() error {
	if err := p.CheckIgnition(); err != nil {
		return err
	}
	if !p.recipeOK {
		return fmt.Errorf("%w", model.ErrRecipeLevelTrip)
	}
	if !p.pressureOK {
		return fmt.Errorf("%w", model.ErrPressureTrip)
	}
	if !p.liquorbankOK {
		return fmt.Errorf("%w", model.ErrLiquorbankTrip)
	}
	return nil
}

type CoordinationLock struct {
	holder string
	held   bool
}

func NewCoordinationLock() *CoordinationLock { return &CoordinationLock{} }

func (c *CoordinationLock) Acquire(holder string) error {
	if c.held {
		return fmt.Errorf("%w", model.ErrCoordinationLock)
	}
	c.holder = holder
	c.held = true
	return nil
}

func (c *CoordinationLock) Release(holder string) {
	if c.held && c.holder == holder {
		c.held = false
		c.holder = ""
	}
}

func (c *CoordinationLock) Require(holder string) error {
	if !c.held || c.holder != holder {
		return fmt.Errorf("%w", model.ErrCoordinationLock)
	}
	return nil
}

func (c *CoordinationLock) Held() bool { return c.held }
