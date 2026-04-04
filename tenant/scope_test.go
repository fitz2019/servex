package tenant

import (
	"testing"
)

func TestIDFromContext(t *testing.T) {
	ctx := WithTenant(t.Context(), &testTenant{id: "t1", enabled: true})
	if got := IDFromContext(ctx); got != "t1" {
		t.Fatalf("IDFromContext = %q, want %q", got, "t1")
	}
}

func TestIDFromContext_Empty(t *testing.T) {
	if got := IDFromContext(t.Context()); got != "" {
		t.Fatalf("IDFromContext = %q, want empty", got)
	}
}

func TestWhereClause(t *testing.T) {
	ctx := WithTenant(t.Context(), &testTenant{id: "t1", enabled: true})

	clause, args := WhereClause(ctx)
	if clause != "tenant_id = ?" {
		t.Fatalf("clause = %q, want %q", clause, "tenant_id = ?")
	}
	if len(args) != 1 || args[0] != "t1" {
		t.Fatalf("args = %v, want [t1]", args)
	}
}

func TestWhereClause_CustomColumn(t *testing.T) {
	ctx := WithTenant(t.Context(), &testTenant{id: "t1", enabled: true})

	clause, args := WhereClause(ctx, "t.tenant_id")
	if clause != "t.tenant_id = ?" {
		t.Fatalf("clause = %q, want %q", clause, "t.tenant_id = ?")
	}
	if len(args) != 1 || args[0] != "t1" {
		t.Fatalf("args = %v, want [t1]", args)
	}
}

func TestWhereClause_NoTenant(t *testing.T) {
	clause, args := WhereClause(t.Context())
	if clause != "" {
		t.Fatalf("clause = %q, want empty", clause)
	}
	if args != nil {
		t.Fatalf("args = %v, want nil", args)
	}
}
