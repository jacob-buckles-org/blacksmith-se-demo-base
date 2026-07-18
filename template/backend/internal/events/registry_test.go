package events

import (
	"strings"
	"testing"
	"time"
)

func TestRegistryCoversTaxonomy(t *testing.T) {
	names := Names()
	// 14 domains x 10 lifecycle actions.
	if len(names) != 140 {
		t.Fatalf("expected 140 registered event types, got %d", len(names))
	}
	for _, name := range names {
		if !strings.Contains(name, ".") {
			t.Errorf("event name %q is not domain.action form", name)
		}
	}
}

func TestNewReturnsValidatingEvents(t *testing.T) {
	event, err := New("ingest.completed")
	if err != nil {
		t.Fatal(err)
	}

	if err := event.Validate(); err == nil {
		t.Error("zero event must fail validation (missing tenant)")
	}

	ingest, ok := event.(*IngestCompleted)
	if !ok {
		t.Fatalf("expected *IngestCompleted, got %T", event)
	}
	ingest.TenantID = "tenant-1"
	ingest.OccurredAt = time.Now().Add(-time.Minute)
	ingest.StatusCode = 200
	if err := ingest.Validate(); err != nil {
		t.Errorf("populated event should validate: %v", err)
	}

	fields := ingest.Fields()
	if fields["event"] != "ingest.completed" {
		t.Errorf("unexpected event field: %v", fields["event"])
	}
}

func TestNewRejectsUnknownTypes(t *testing.T) {
	if _, err := New("nope.never"); err == nil {
		t.Error("expected error for unknown event type")
	}
}

func TestValidateRejectsBadValues(t *testing.T) {
	event := &QueryFailed{
		TenantID:   "tenant-1",
		OccurredAt: time.Now().Add(-time.Minute),
		StatusCode: 42,
	}
	if err := event.Validate(); err == nil {
		t.Error("expected implausible status code to fail validation")
	}

	event.StatusCode = 500
	event.DurationMs = -1
	if err := event.Validate(); err == nil {
		t.Error("expected negative duration to fail validation")
	}
}
