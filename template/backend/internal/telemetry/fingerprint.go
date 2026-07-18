package telemetry

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
)

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
