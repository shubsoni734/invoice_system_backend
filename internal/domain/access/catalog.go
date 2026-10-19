package access

// Page is one screen an organisation role can be allowed to use.
type Page struct {
	Resource string
	Label    string
	Actions  []string
}

const (
	ActionCreate = "create"
	ActionRead   = "read"
	ActionUpdate = "update"
	ActionDelete = "delete"
)

// Pages is the catalog of organisation screens and the actions each one supports.
var Pages = []Page{
	{Resource: "dashboard", Label: "Dashboard", Actions: []string{ActionRead}},
	{Resource: "users", Label: "Users", Actions: []string{ActionCreate, ActionRead, ActionUpdate, ActionDelete}},
	{Resource: "roles", Label: "Roles", Actions: []string{ActionCreate, ActionRead, ActionUpdate, ActionDelete}},
	{Resource: "access", Label: "Access", Actions: []string{ActionCreate, ActionRead, ActionUpdate, ActionDelete}},
	{Resource: "customers", Label: "Customers", Actions: []string{ActionCreate, ActionRead, ActionUpdate, ActionDelete}},
	{Resource: "services", Label: "Services", Actions: []string{ActionCreate, ActionRead, ActionUpdate, ActionDelete}},
	{Resource: "invoices", Label: "Invoices", Actions: []string{ActionCreate, ActionRead, ActionUpdate, ActionDelete}},
	{Resource: "payments", Label: "Payments", Actions: []string{ActionCreate, ActionRead, ActionUpdate, ActionDelete}},
	{Resource: "expenses", Label: "Expenses", Actions: []string{ActionCreate, ActionRead, ActionUpdate, ActionDelete}},
	{Resource: "documents", Label: "Documents", Actions: []string{ActionCreate, ActionRead, ActionUpdate, ActionDelete}},
	{Resource: "reports", Label: "Reports", Actions: []string{ActionRead}},
	{Resource: "gst", Label: "GST R1", Actions: []string{ActionRead}},
	{Resource: "whatsapp", Label: "WhatsApp", Actions: []string{ActionCreate, ActionRead, ActionUpdate}},
	{Resource: "settings", Label: "Settings", Actions: []string{ActionRead, ActionUpdate}},
}

// Merge combines several named accesses. A page is allowed when any linked access allows it.
func Merge(grants []Grant) []Grant {
	byResource := map[string]Grant{}
	for _, grant := range grants {
		current := byResource[grant.Resource]
		current.Resource = grant.Resource
		current.CanCreate = current.CanCreate || grant.CanCreate
		current.CanRead = current.CanRead || grant.CanRead
		current.CanUpdate = current.CanUpdate || grant.CanUpdate
		current.CanDelete = current.CanDelete || grant.CanDelete
		byResource[grant.Resource] = current
	}
	flat := make([]Grant, 0, len(byResource))
	for _, grant := range byResource {
		flat = append(flat, grant)
	}
	return Normalize(flat)
}

// Grant is the access stored for one role and one page.
type Grant struct {
	Resource  string `json:"resource"`
	CanCreate bool   `json:"can_create"`
	CanRead   bool   `json:"can_read"`
	CanUpdate bool   `json:"can_update"`
	CanDelete bool   `json:"can_delete"`
}

func pageByResource(resource string) (Page, bool) {
	for _, page := range Pages {
		if page.Resource == resource {
			return page, true
		}
	}
	return Page{}, false
}

func pageAllows(resource, action string) bool {
	page, ok := pageByResource(resource)
	if !ok {
		return false
	}
	for _, allowed := range page.Actions {
		if allowed == action {
			return true
		}
	}
	return false
}

// Normalize keeps only known pages and turns off actions that page does not support.
func Normalize(grants []Grant) []Grant {
	byResource := map[string]Grant{}
	for _, grant := range grants {
		if _, ok := pageByResource(grant.Resource); ok {
			byResource[grant.Resource] = grant
		}
	}
	out := make([]Grant, 0, len(Pages))
	for _, page := range Pages {
		grant := byResource[page.Resource]
		grant.Resource = page.Resource
		if !pageAllows(page.Resource, ActionCreate) {
			grant.CanCreate = false
		}
		if !pageAllows(page.Resource, ActionRead) {
			grant.CanRead = false
		}
		if !pageAllows(page.Resource, ActionUpdate) {
			grant.CanUpdate = false
		}
		if !pageAllows(page.Resource, ActionDelete) {
			grant.CanDelete = false
		}
		out = append(out, grant)
	}
	return out
}

// FullGrants marks every supported action as allowed.
func FullGrants() []Grant {
	grants := make([]Grant, 0, len(Pages))
	for _, page := range Pages {
		grant := Grant{Resource: page.Resource}
		for _, action := range page.Actions {
			setAction(&grant, action, true)
		}
		grants = append(grants, grant)
	}
	return grants
}

// Allows reports whether a grant permits an action on a known page.
func Allows(grants []Grant, resource, action string) bool {
	if !pageAllows(resource, action) {
		return false
	}
	for _, grant := range grants {
		if grant.Resource != resource {
			continue
		}
		switch action {
		case ActionCreate:
			return grant.CanCreate
		case ActionRead:
			return grant.CanRead
		case ActionUpdate:
			return grant.CanUpdate
		case ActionDelete:
			return grant.CanDelete
		}
	}
	return false
}

func setAction(grant *Grant, action string, value bool) {
	switch action {
	case ActionCreate:
		grant.CanCreate = value
	case ActionRead:
		grant.CanRead = value
	case ActionUpdate:
		grant.CanUpdate = value
	case ActionDelete:
		grant.CanDelete = value
	}
}
