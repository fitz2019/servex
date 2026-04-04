// webhook/dispatcher_test.go
package webhook

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestDispatcher_Dispatch(t *testing.T) {
	var receivedBody []byte
	var receivedSig string
	var receivedType string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedBody, _ = io.ReadAll(r.Body)
		receivedSig = r.Header.Get("X-Webhook-Signature")
		receivedType = r.Header.Get("X-Webhook-Event")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	d := NewDispatcher(WithTimeout(5 * time.Second))
	defer d.Close()

	sub := &Subscription{ID: "s1", URL: server.URL, Secret: "my-secret"}
	event := &Event{ID: "e1", Type: "order.created", Payload: []byte(`{"id":1}`), Timestamp: time.Now()}

	err := d.Dispatch(t.Context(), sub, event)
	if err != nil {
		t.Fatal(err)
	}

	if string(receivedBody) != `{"id":1}` {
		t.Errorf("body = %s", receivedBody)
	}
	if receivedSig == "" {
		t.Error("signature header should be set")
	}
	if receivedType != "order.created" {
		t.Errorf("event type = %s", receivedType)
	}
}

func TestDispatcher_Dispatch_NilSubscription(t *testing.T) {
	d := NewDispatcher()
	err := d.Dispatch(t.Context(), nil, &Event{})
	if !errors.Is(err, ErrNilSubscription) {
		t.Errorf("got %v, want ErrNilSubscription", err)
	}
}

func TestDispatcher_Dispatch_NilEvent(t *testing.T) {
	d := NewDispatcher()
	err := d.Dispatch(t.Context(), &Subscription{URL: "http://x"}, nil)
	if !errors.Is(err, ErrNilEvent) {
		t.Errorf("got %v, want ErrNilEvent", err)
	}
}

func TestDispatcher_Dispatch_EmptyURL(t *testing.T) {
	d := NewDispatcher()
	err := d.Dispatch(t.Context(), &Subscription{}, &Event{})
	if !errors.Is(err, ErrEmptyURL) {
		t.Errorf("got %v, want ErrEmptyURL", err)
	}
}

func TestDispatcher_Dispatch_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	d := NewDispatcher()
	sub := &Subscription{URL: server.URL, Secret: "s"}
	event := &Event{Payload: []byte("x")}

	err := d.Dispatch(t.Context(), sub, event)
	if err == nil {
		t.Error("expected error for 500 response")
	}
}

func TestDispatcher_ImplementsInterface(t *testing.T) {
	var _ Dispatcher = (*dispatcher)(nil)
}
