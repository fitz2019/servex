package endpoint

import (
	"context"
	"errors"
	"testing"
)

func TestNop(t *testing.T) {
	resp, err := Nop(t.Context(), "anything")
	if err != nil {
		t.Fatalf("Nop returned error: %v", err)
	}
	if resp == nil {
		t.Fatal("Nop returned nil response")
	}
}

func TestNopMiddleware(t *testing.T) {
	called := false
	inner := func(ctx context.Context, req any) (any, error) {
		called = true
		return "ok", nil
	}

	wrapped := NopMiddleware(inner)
	resp, err := wrapped(t.Context(), "req")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !called {
		t.Fatal("inner endpoint was not called")
	}
	if resp != "ok" {
		t.Fatalf("unexpected response: %v", resp)
	}
}

func TestChain(t *testing.T) {
	var order []string

	makeMW := func(name string) Middleware {
		return func(next Endpoint) Endpoint {
			return func(ctx context.Context, req any) (any, error) {
				order = append(order, name+"-before")
				resp, err := next(ctx, req)
				order = append(order, name+"-after")
				return resp, err
			}
		}
	}

	chained := Chain(makeMW("A"), makeMW("B"), makeMW("C"))
	ep := chained(func(ctx context.Context, req any) (any, error) {
		order = append(order, "endpoint")
		return "done", nil
	})

	resp, err := ep(t.Context(), nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp != "done" {
		t.Fatalf("unexpected response: %v", resp)
	}

	expected := []string{"A-before", "B-before", "C-before", "endpoint", "C-after", "B-after", "A-after"}
	if len(order) != len(expected) {
		t.Fatalf("expected %d calls, got %d: %v", len(expected), len(order), order)
	}
	for i, v := range expected {
		if order[i] != v {
			t.Fatalf("order[%d] = %q, want %q", i, order[i], v)
		}
	}
}

func TestChainErrorPropagation(t *testing.T) {
	errTest := errors.New("test error")
	chained := Chain(NopMiddleware)
	ep := chained(func(ctx context.Context, req any) (any, error) {
		return nil, errTest
	})

	_, err := ep(t.Context(), nil)
	if !errors.Is(err, errTest) {
		t.Fatalf("expected errTest, got %v", err)
	}
}
