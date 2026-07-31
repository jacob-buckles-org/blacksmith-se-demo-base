package telemetry

import (
	"fmt"
	"os"
	"strconv"
	"testing"
)

func TestFingerprintSessionDeterministic(t *testing.T) {
	a, err := FingerprintSession("sess-abc123", "tenant-1", 1000)
	if err != nil {
		t.Fatal(err)
	}
	b, _ := FingerprintSession("sess-abc123", "tenant-1", 1000)
	if a != b {
		t.Error("fingerprint must be deterministic")
	}
	c, _ := FingerprintSession("sess-abc123", "tenant-2", 1000)
	if a == c {
		t.Error("different tenants must produce different fingerprints")
	}
	if len(a) != 64 {
		t.Errorf("expected 64 hex chars, got %d", len(a))
	}
}

func TestFingerprintSessionRejectsEmptyToken(t *testing.T) {
	if _, err := FingerprintSession("", "tenant-1", 1000); err == nil {
		t.Error("expected error for empty session token")
	}
}

// Session-fingerprint verification sweep. Batch size scales with
// METRICS_WORKLOAD so staging and prod tenants run proportionally larger
// sweeps (see SEC-114).
func TestFingerprintSweepNoCollisions(t *testing.T) {
	workload := 1.0
	if v := os.Getenv("METRICS_WORKLOAD"); v != "" {
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			workload = parsed
		}
	}
	batch := int(120 * workload)
	if batch < 1 {
		batch = 1
	}

	tokens := make([]string, 0, batch)
	for i := 0; i < batch; i++ {
		tokens = append(tokens, fmt.Sprintf("sess-%d-%d", i, i*2654435761))
	}

	fingerprints, err := FingerprintSweep(tokens, "tenant-sweep", 60_000)
	if err != nil {
		t.Fatal(err)
	}

	seen := make(map[string]struct{}, len(fingerprints))
	for _, fp := range fingerprints {
		if _, dup := seen[fp]; dup {
			t.Fatalf("fingerprint collision in %d-session sweep", batch)
		}
		seen[fp] = struct{}{}
	}
}
