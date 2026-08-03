package strategyevidence

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"testing"
	"time"

	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func TestStoreStampsTrustedIngestionClockAndRejectsBackdating(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	trusted := time.Date(2026, 8, 3, 18, 0, 0, 123, time.UTC)
	store, err := Open(ctx, Options{Path: filepath.Join(t.TempDir(), "evidence.db"), Clock: marketclock.NewFake(trusted)})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	h := validHeader(marketclock.MarketUS, KindUSParticipation, "trusted-ingest", "rev-1")
	h.IngestedAt = h.ObservedAt // maliciously early but structurally valid
	result, err := store.Append(ctx, mustEnvelope(t, h, `{"value":1}`))
	if err != nil {
		t.Fatal(err)
	}
	if got := result.Evidence.Header().IngestedAt; !got.Equal(trusted) {
		t.Fatalf("ingested_at=%s want trusted %s", got, trusted)
	}
	before, err := store.SealSnapshot(ctx, snapshotQueryFor(result.Evidence, trusted.Add(-time.Nanosecond)))
	if err != nil {
		t.Fatal(err)
	}
	if len(before.Items) != 0 {
		t.Fatal("caller-backdated evidence crossed trusted ingestion cutoff")
	}
}

func TestStoreSameDigestWithChangedProvenanceIsConflict(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := openTestStore(t)
	h := validHeader(marketclock.MarketUS, KindUSParticipation, "stable", "rev-1")
	first := mustEnvelope(t, h, `{"value":1}`)
	if _, err := store.Append(ctx, first); err != nil {
		t.Fatal(err)
	}
	h.EvidenceID = "forged"
	h.SourceAvailableAt = h.SourceAvailableAt.Add(time.Second)
	h.ObservedAt = h.ObservedAt.Add(time.Second)
	if _, err := store.Append(ctx, mustEnvelope(t, h, `{"value":1}`)); !errors.Is(err, ErrRevisionConflict) {
		t.Fatalf("same digest with provenance drift error=%v", err)
	}
}

func TestSnapshotRequiresIssuerAndMappingScope(t *testing.T) {
	t.Parallel()
	ctx := context.Background()
	store := openTestStore(t)
	h := validHeader(marketclock.MarketUS, KindUSParticipation, "issuer", "rev-1")
	stored, err := store.Append(ctx, mustEnvelope(t, h, `{"value":1}`))
	if err != nil {
		t.Fatal(err)
	}
	query := snapshotQueryFor(stored.Evidence, stored.Evidence.Header().IngestedAt)
	query.IssuerIdentity = ""
	if _, err := store.SealSnapshot(ctx, query); err == nil {
		t.Fatal("snapshot without issuer identity was accepted")
	}
	query = snapshotQueryFor(stored.Evidence, stored.Evidence.Header().IngestedAt)
	query.IssuerMappingVersion = "mapping-v2"
	snapshot, err := store.SealSnapshot(ctx, query)
	if err != nil {
		t.Fatal(err)
	}
	if len(snapshot.Items) != 0 {
		t.Fatal("evidence from another mapping version leaked into snapshot")
	}
}

func TestCanonicalJSONNormalizesEquivalentNumbersAndRejectsDuplicateKeys(t *testing.T) {
	t.Parallel()
	h := validHeader(marketclock.MarketUS, KindUSParticipation, "numeric", "rev-1")
	var digest string
	for _, payload := range []string{`{"value":1}`, `{"value":1.0}`, `{"value":1e0}`} {
		evidence, err := NewEnvelope(h, []byte(payload))
		if err != nil {
			t.Fatal(err)
		}
		if digest != "" && digest != evidence.PayloadDigest() {
			t.Fatalf("numeric canonicalization differs for %s", payload)
		}
		digest = evidence.PayloadDigest()
	}
	zero, err := NewEnvelope(h, []byte(`{"value":-0}`))
	if err != nil {
		t.Fatal(err)
	}
	positive, err := NewEnvelope(h, []byte(`{"value":0}`))
	if err != nil || zero.PayloadDigest() != positive.PayloadDigest() {
		t.Fatalf("negative zero not canonical: %v", err)
	}
	for _, pair := range [][2]string{
		{`{"value":12.30}`, `{"value":123e-1}`},
		{`{"value":0.001}`, `{"value":1e-3}`},
		{`{"value":1000}`, `{"value":1e3}`},
		{`{"value":-0.000e999}`, `{"value":0}`},
	} {
		left, leftErr := NewEnvelope(h, []byte(pair[0]))
		right, rightErr := NewEnvelope(h, []byte(pair[1]))
		if leftErr != nil || rightErr != nil || left.PayloadDigest() != right.PayloadDigest() {
			t.Fatalf("equivalent decimals differ: %s / %s, errors=%v/%v", pair[0], pair[1], leftErr, rightErr)
		}
	}
	if _, err := NewEnvelope(h, []byte(`{"value":1,"value":2}`)); err == nil {
		t.Fatal("duplicate JSON key accepted")
	}
}

func TestCanonicalJSONPreservesArbitraryPrecisionDecimalIdentity(t *testing.T) {
	t.Parallel()
	h := validHeader(marketclock.MarketUS, KindUSParticipation, "long-decimal", "rev-1")
	prefix := strings.Repeat("1234567890", 70)
	first := mustEnvelope(t, h, `{"value":`+prefix+`1}`)
	second := mustEnvelope(t, h, `{"value":`+prefix+`2}`)
	if first.PayloadDigest() == second.PayloadDigest() {
		t.Fatal("distinct long-precision decimals collided")
	}
	equivalent := mustEnvelope(t, h, `{"value":`+prefix+`1.000e0}`)
	if first.PayloadDigest() != equivalent.PayloadDigest() {
		t.Fatal("exactly equivalent long-precision decimals did not canonicalize")
	}

	largeExponent := mustEnvelope(t, h, `{"value":1e1000000}`)
	sameLargeExponent := mustEnvelope(t, h, `{"value":10e999999}`)
	if largeExponent.PayloadDigest() != sameLargeExponent.PayloadDigest() {
		t.Fatal("equivalent bounded large exponents did not canonicalize")
	}
	neighbor := mustEnvelope(t, h, `{"value":1e999999}`)
	if largeExponent.PayloadDigest() == neighbor.PayloadDigest() {
		t.Fatal("distinct large exponents collided")
	}
}

func TestCanonicalJSONRejectsOversizedNumbersAndExponents(t *testing.T) {
	t.Parallel()
	h := validHeader(marketclock.MarketUS, KindUSParticipation, "number-bound", "rev-1")
	for _, payload := range []string{
		`{"value":` + strings.Repeat("1", maxCanonicalNumberBytes+1) + `}`,
		`{"value":1e1000001}`,
		`{"value":1e-1000001}`,
		`{"value":0e1000001}`,
	} {
		if _, err := NewEnvelope(h, []byte(payload)); err == nil {
			t.Fatalf("unbounded numeric input accepted: length=%d", len(payload))
		}
	}
}

func TestTypedPayloadRejectsWrongTypesAndSecretFields(t *testing.T) {
	t.Parallel()
	h := validHeader(marketclock.MarketUS, KindUSParticipation, "typed", "rev-1")
	for _, payload := range []string{`{"score_ppm":"high"}`, `{"api_key":"must-not-store"}`, `{"nested":{"access_token":"must-not-store"}}`} {
		if _, err := NewEnvelope(h, []byte(payload)); err == nil {
			t.Fatalf("unsafe payload accepted: %s", payload)
		}
	}
	provider := StaticCredential("credential-value")
	if got := fmt.Sprintf("%v", provider); strings.Contains(got, "credential-value") || got != "[REDACTED]" {
		t.Fatalf("credential provider formatting=%q", got)
	}
	if got := fmt.Sprintf("%#v", provider); strings.Contains(got, "credential-value") {
		t.Fatalf("credential provider Go formatting leaked: %q", got)
	}
}

func TestProjectionWithoutAuthorityPriorityFailsClosed(t *testing.T) {
	t.Parallel()
	h := validHeader(marketclock.MarketUS, KindUSParticipation, "empty-priority", "rev-1")
	evidence := mustEnvelope(t, h, `{"value":1}`)
	storedHeader := evidence.Header()
	snapshot := Snapshot{Market: storedHeader.Market, Symbol: storedHeader.Symbol, IssuerIdentity: storedHeader.IssuerIdentity, IssuerMappingVersion: storedHeader.IssuerMappingVersion, EvaluationAt: storedHeader.ObservedAt.Add(time.Minute), Items: []Envelope{evidence}}
	result, err := Project(snapshot, ProjectionPolicy{Version: "v1", Market: h.Market, Required: []Requirement{{Kind: h.Kind, Required: true, MaxAge: time.Hour}}})
	if err != nil {
		t.Fatal(err)
	}
	if result.Lane.Eligible || !hasRefusal(result.Lane.Refusals, RefusalEvidenceMissing) {
		t.Fatalf("empty authority policy failed open: %+v", result.Lane)
	}
}

func snapshotQueryFor(evidence Envelope, cutoff time.Time) SnapshotQuery {
	h := evidence.Header()
	return SnapshotQuery{Market: h.Market, Symbol: h.Symbol, IssuerIdentity: h.IssuerIdentity, IssuerMappingVersion: h.IssuerMappingVersion, EvaluationAt: h.SourceAvailableAt.Add(time.Hour), IngestionCutoff: cutoff}
}
