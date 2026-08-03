package reversallane

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"
)

func ValidateStructure(structure StructuralConfirmation, envelope CommonEnvelope, window time.Duration) RefusalCode {
	events := []struct {
		want StructuralEventKind
		got  StructuralEvent
	}{{EventSweep, structure.Sweep}, {EventBreak, structure.Break}, {EventReclaim, structure.Reclaim}}
	for _, item := range events {
		event := item.got
		if event.Kind != item.want || strings.TrimSpace(event.RecordID) == "" || strings.TrimSpace(event.Digest) == "" || event.At.IsZero() || event.FreshUntil.IsZero() {
			return RefusalStructuralMissing
		}
		if len(event.AccountRef) > maxIdentityBytes || len(event.Symbol) > maxIdentityBytes || len(event.EvidenceVersion) > maxIdentityBytes || len(event.RecordID) > maxIdentityBytes || len(event.Digest) > maxIdentityBytes {
			return RefusalStructuralMissing
		}
		if event.AccountRef != envelope.AccountRef || event.Market != envelope.Market || event.Symbol != envelope.Symbol || event.PositionGeneration != envelope.PositionGeneration || event.EvidenceVersion != envelope.SchemaVersion {
			return RefusalScopeMismatch
		}
		if event.At.After(event.FreshUntil) || envelope.EvaluatedAt.After(event.FreshUntil) {
			return RefusalStructuralStale
		}
	}
	if structure.Sweep.At.After(structure.Break.At) || structure.Break.At.After(structure.Reclaim.At) || structure.Reclaim.At.After(envelope.EvaluatedAt) {
		return RefusalStructuralOrder
	}
	if structure.Sweep.RecordID == structure.Break.RecordID || structure.Sweep.RecordID == structure.Reclaim.RecordID || structure.Break.RecordID == structure.Reclaim.RecordID ||
		structure.Sweep.Digest == structure.Break.Digest || structure.Sweep.Digest == structure.Reclaim.Digest || structure.Break.Digest == structure.Reclaim.Digest {
		return RefusalStructuralOrder
	}
	if window < 0 || envelope.EvaluatedAt.Sub(structure.Sweep.At) > window {
		return RefusalStructuralStale
	}
	return ""
}

func structuralDigest(structure StructuralConfirmation) string {
	hash := sha256.New()
	for _, event := range []StructuralEvent{structure.Sweep, structure.Break, structure.Reclaim} {
		_, _ = fmt.Fprintf(hash, "%s\x00%s\x00%s\x00%s\x00%d\x00%s\x00%s\x00%s\n", event.Kind, event.AccountRef, event.Market, event.Symbol, event.PositionGeneration, event.EvidenceVersion, event.RecordID, event.Digest)
		_, _ = hash.Write([]byte(event.At.UTC().Format(time.RFC3339Nano)))
		_, _ = hash.Write([]byte{0})
		_, _ = hash.Write([]byte(event.FreshUntil.UTC().Format(time.RFC3339Nano)))
	}
	return hex.EncodeToString(hash.Sum(nil))
}
