package constants

// Context keys
const (
	CtxUserID               = "user_id"
	CtxOrgID                = "organisation_id"
	CtxUserRole             = "user_role"
	CtxSuperAdminID         = "super_admin_id"
	CtxSuperAdminRole       = "super_admin_role"
	CtxRequestID            = "request_id"
	CtxOrgPlan              = "org_plan"
	CtxIsImpersonating      = "is_impersonating"
	CtxImpersonationSession = "impersonation_session_id"
)

// SuperAdmin roles
const (
	RoleSuperAdmin = "superadmin"
	RoleSupport    = "support"
	RoleFinance    = "finance"
	RoleReadonly   = "readonly"
)

// Organisation roles
const (
	RoleOrgAdmin   = "admin"
	RoleOrgManager = "manager"
	RoleOrgViewer  = "viewer"
)
