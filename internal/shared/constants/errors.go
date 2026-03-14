package constants

import "errors"

var (
	ErrNotFound       = errors.New("resource not found")
	ErrUnauthorised   = errors.New("unauthorised")
	ErrForbidden      = errors.New("forbidden")
	ErrValidation     = errors.New("validation failed")
	ErrConflict       = errors.New("resource already exists")
	ErrPlanLimit      = errors.New("plan limit reached — upgrade your plan")
	ErrOrgSuspended   = errors.New("organisation is suspended")
	ErrAccountLocked  = errors.New("account locked — too many failed attempts")
	ErrTokenExpired   = errors.New("token has expired")
	ErrTokenInvalid   = errors.New("token is invalid")
	ErrMaintenance    = errors.New("platform is under maintenance")
	ErrInternalServer = errors.New("internal server error")
)
