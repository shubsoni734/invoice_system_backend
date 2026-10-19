package access

import "testing"

func TestNormalizeDropsUnknownPagesAndUnsupportedActions(t *testing.T) {
	grants := Normalize([]Grant{
		{Resource: "unknown", CanRead: true},
		{Resource: "dashboard", CanCreate: true, CanRead: true, CanDelete: true},
		{Resource: "customers", CanCreate: true, CanUpdate: true},
	})
	if len(grants) != len(Pages) {
		t.Fatalf("expected %d pages, got %d", len(Pages), len(grants))
	}
	if grants[0].Resource != "dashboard" || grants[0].CanCreate || !grants[0].CanRead || grants[0].CanDelete {
		t.Fatalf("dashboard grant = %+v", grants[0])
	}
	if !Allows(grants, "customers", ActionCreate) || Allows(grants, "customers", ActionRead) {
		t.Fatal("customer flags were not preserved")
	}
	if Allows(grants, "dashboard", ActionCreate) || Allows(grants, "missing", ActionRead) {
		t.Fatal("unsupported actions must stay denied")
	}
}

func TestMergeAllowsAPageWhenAnyAccessAllowsIt(t *testing.T) {
	merged := Merge([]Grant{
		{Resource: "customers", CanRead: true},
		{Resource: "customers", CanCreate: true},
		{Resource: "invoices", CanRead: true},
	})
	if !Allows(merged, "customers", ActionRead) || !Allows(merged, "customers", ActionCreate) {
		t.Fatal("customer actions should be combined")
	}
	if !Allows(merged, "invoices", ActionRead) || Allows(merged, "invoices", ActionDelete) {
		t.Fatal("invoice read should stay limited to the linked access")
	}
}

func TestFullGrantsCoverEveryPage(t *testing.T) {
	grants := FullGrants()
	if !Allows(grants, "invoices", ActionDelete) || !Allows(grants, "reports", ActionRead) {
		t.Fatal("full access must allow supported actions")
	}
	if Allows(grants, "reports", ActionDelete) {
		t.Fatal("reports have no delete action")
	}
}
