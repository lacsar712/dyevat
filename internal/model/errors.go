package model

import "errors"

var (
	ErrContextDone      = errors.New("operation cancelled")
	ErrPlantNotFound    = errors.New("plant unit not found")
	ErrLeaseHeld        = errors.New("interlock lease held by another operator")
	ErrLeaseMissing     = errors.New("interlock lease missing or expired")
	ErrGateBlocked      = errors.New("safety gate blocked")
	ErrLiquorPermissive   = errors.New("liquor permissive not satisfied")
	ErrIgnitionBlocked  = errors.New("ignition sequence blocked")
	ErrRecipeLevelTrip    = errors.New("recipe level trip condition")
	ErrPressureTrip     = errors.New("steam pressure trip condition")
	ErrLiquorbankTrip   = errors.New("liquorbank trip condition")
	ErrIllegalState     = errors.New("illegal plant state transition")
	ErrSnapshotStale    = errors.New("snapshot revision stale")
	ErrWindowOpen       = errors.New("timing window still open")
	ErrSoakIncomplete  = errors.New("vatlock soak incomplete")
	ErrCoordinationLock = errors.New("coordination lock held")
	ErrRecipeLevelLow     = errors.New("recipe level below low limit")
	ErrShadeLoss        = errors.New("vatlock shade lost")
	ErrDrainLimit    = errors.New("drain valve at limit")
)
