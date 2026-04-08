package pbjson

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"google.golang.org/protobuf/types/known/wrapperspb"
)

func TestMarshalUnmarshal(t *testing.T) {
	msg := wrapperspb.String("hello")

	data, err := Marshal(msg)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	got := &wrapperspb.StringValue{}
	if err := Unmarshal(data, got); err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}
	if got.GetValue() != "hello" {
		t.Fatalf("got %q, want %q", got.GetValue(), "hello")
	}
}

func TestMarshalString(t *testing.T) {
	msg := wrapperspb.Int32(42)

	s, err := MarshalString(msg)
	if err != nil {
		t.Fatalf("MarshalString error: %v", err)
	}
	if !strings.Contains(s, "42") {
		t.Fatalf("expected string to contain 42, got %q", s)
	}
}

func TestEncodeResponse(t *testing.T) {
	msg := wrapperspb.String("test")
	w := httptest.NewRecorder()

	err := EncodeResponse(t.Context(), w, msg)
	if err != nil {
		t.Fatalf("EncodeResponse error: %v", err)
	}

	if ct := w.Header().Get("Content-Type"); ct != "application/json; charset=utf-8" {
		t.Fatalf("Content-Type = %q", ct)
	}
	if !strings.Contains(w.Body.String(), "test") {
		t.Fatalf("body does not contain expected value: %s", w.Body.String())
	}
}

func TestEncodeResponseNotProto(t *testing.T) {
	w := httptest.NewRecorder()
	err := EncodeResponse(t.Context(), w, "not a proto message")
	if err != ErrNotProtoMessage {
		t.Fatalf("expected ErrNotProtoMessage, got %v", err)
	}
}

func TestDecodeRequest(t *testing.T) {
	// protobuf JSON for StringValue is just a quoted string.
	body := `"decoded"`
	r := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))

	msg := &wrapperspb.StringValue{}
	if err := DecodeRequest(r, msg); err != nil {
		t.Fatalf("DecodeRequest error: %v", err)
	}
	if msg.GetValue() != "decoded" {
		t.Fatalf("got %q, want %q", msg.GetValue(), "decoded")
	}
}
