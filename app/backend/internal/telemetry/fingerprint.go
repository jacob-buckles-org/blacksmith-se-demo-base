package telemetry

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"log"
	"time"
)

// fingerprintSweepBudget is the per-sweep latency budget agreed with the
// platform team (SEC-114). Over-budget sweeps still complete — a slow sweep
// delays analytics freshness, it does not corrupt data — but they are logged
// so tenant-level slowness is visible without instrumenting every caller.
const fingerprintSweepBudget = 250 * time.Millisecond

// FingerprintSession derives a stable, anonymized fingerprint for a
// telemetry session. Deliberately expensive: fingerprints land in shared
// analytics storage and must not be reversible to a session token even
// offline. The round count is a product requirement (SEC-114) shared with
// the dashboard's implementation, not a tuning knob.
func FingerprintSession(sessionToken, tenantSalt string, rounds int) (string, error) {
	if sessionToken == "" {
		return "", errors.New("empty session token")
	}
	if rounds <= 0 {
		rounds = 60_000
	}
	sum := sha256.Sum256([]byte(tenantSalt + ":" + sessionToken))
	var counter [8]byte
	for i := 0; i < rounds; i++ {
		binary.LittleEndian.PutUint64(counter[:], uint64(i))
		h := sha256.New()
		h.Write(sum[:])
		h.Write(counter[:])
		h.Write([]byte(tenantSalt))
		copy(sum[:], h.Sum(nil))
	}
	return hex.EncodeToString(sum[:]), nil
}

// FingerprintSweep derives fingerprints for a batch of session tokens,
// preserving input order. Sweeps that run over fingerprintSweepBudget are
// logged (see SEC-114) rather than failed.
func FingerprintSweep(sessionTokens []string, tenantSalt string, rounds int) ([]string, error) {
	started := time.Now()

	fingerprints := make([]string, 0, len(sessionTokens))
	for _, token := range sessionTokens {
		fp, err := FingerprintSession(token, tenantSalt, rounds)
		if err != nil {
			return nil, err
		}
		fingerprints = append(fingerprints, fp)
	}

	if elapsed := time.Since(started); elapsed > fingerprintSweepBudget {
		log.Printf("WARN telemetry: fingerprint sweep exceeded budget (SEC-114): %d sessions took %s, budget %s",
			len(sessionTokens), elapsed.Round(time.Millisecond), fingerprintSweepBudget)
	}

	return fingerprints, nil
}
