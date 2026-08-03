package strategyevidence

import (
	"bytes"
	"errors"
	"testing"
	"time"

	marketclock "github.com/JungHoonGhae/tossinvest-cli/internal/clock"
)

func TestNewEnvelopeCanonicalizesAndCopiesPayload(t *testing.T) {
	t.Parallel()
	header := validHeader(marketclock.MarketUS, KindUSParticipation, "ev-1", "rev-1")
	payloadA := []byte(`{"volume": 12, "price": {"minor": "100"}}`)
	payloadB := []byte("{\n  \"price\": {\"minor\": \"100\"}, \"volume\": 12\n}")

	a, err := NewEnvelope(header, payloadA)
	if err != nil {
		t.Fatalf("NewEnvelope A: %v", err)
	}
	b, err := NewEnvelope(header, payloadB)
	if err != nil {
		t.Fatalf("NewEnvelope B: %v", err)
	}
	if a.PayloadDigest() != b.PayloadDigest() {
		t.Fatalf("canonical digest differs: %s != %s", a.PayloadDigest(), b.PayloadDigest())
	}
	payloadA[2] = 'X'
	got := a.CanonicalPayload()
	got[0] = 'X'
	if bytes.Equal(a.CanonicalPayload(), got) {
		t.Fatal("CanonicalPayload returned mutable storage")
	}
	if a.Header().Market != marketclock.MarketUS {
		t.Fatalf("market = %q", a.Header().Market)
	}
}

func TestNewEnvelopeRejectsCrossMarketKindAndIncompleteIdentity(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		mutate func(*Header)
		code   RefusalCode
	}{
		{"cross-market kind", func(h *Header) { h.Kind = KindKRNetFlow }, RefusalUnsupportedEvidence},
		{"missing issuer", func(h *Header) { h.IssuerIdentity = "" }, RefusalIdentityMismatch},
		{"missing source availability", func(h *Header) { h.SourceAvailableAt = time.Time{} }, RefusalSourceUnavailable},
		{"missing currency", func(h *Header) { h.Currency = "" }, RefusalCurrencyUnresolved},
		{"backwards observed clock", func(h *Header) { h.ObservedAt = h.SourceAvailableAt.Add(-time.Nanosecond) }, RefusalTimestampInvalid},
		{"cross-market authority", func(h *Header) { h.Authority = AuthorityKRX }, RefusalIdentityMismatch},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := validHeader(marketclock.MarketUS, KindUSParticipation, "ev", "rev")
			tt.mutate(&h)
			_, err := NewEnvelope(h, []byte(`{"value":1}`))
			var refusal *ValidationError
			if !errors.As(err, &refusal) || refusal.Code != tt.code {
				t.Fatalf("want %s, got %v", tt.code, err)
			}
		})
	}
}

func TestNewEnvelopeRequiresTopLevelJSONObject(t *testing.T) {
	t.Parallel()
	for _, payload := range []string{`[]`, `"scalar"`, `1`, `true`, `null`} {
		payload := payload
		t.Run(payload, func(t *testing.T) {
			t.Parallel()
			_, err := NewEnvelope(
				validHeader(marketclock.MarketUS, KindUSParticipation, "ev", "rev"),
				[]byte(payload),
			)
			var refusal *ValidationError
			if !errors.As(err, &refusal) || refusal.Code != RefusalPayloadInvalid {
				t.Fatalf("want %s for %s, got %v", RefusalPayloadInvalid, payload, err)
			}
		})
	}
}

func validHeader(market marketclock.Market, kind EvidenceKind, id, revision string) Header {
	sourceAvailable := time.Date(2026, 8, 3, 13, 0, 0, 0, time.UTC)
	return Header{
		EvidenceID:           id,
		Market:               market,
		Symbol:               "aapl",
		IssuerIdentity:       "issuer-aapl",
		IssuerMappingVersion: "issuer-map-v1",
		Kind:                 kind,
		SchemaVersion:        "v1",
		Authority:            AuthoritySEC,
		SourceRecordID:       "record-1",
		RevisionIdentity:     revision,
		MarketEffectiveDate:  "2026-08-03",
		SourceEventAt:        sourceAvailable.Add(-time.Hour),
		SourceAvailableAt:    sourceAvailable,
		ObservedAt:           sourceAvailable.Add(time.Minute),
		IngestedAt:           sourceAvailable.Add(2 * time.Minute),
		Currency:             "USD",
		Unit:                 "minor",
		Availability:         AvailabilityAvailable,
		Confidence:           ConfidenceVerified,
	}
}
