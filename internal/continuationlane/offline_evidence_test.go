package continuationlane

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
	"time"

	evidence "github.com/JungHoonGhae/tossinvest-cli/internal/campaignstateevidence"
)

func TestOfflineEvidencePreservesContinuation(t *testing.T) {
	for _, c := range []struct {
		market         Market
		lane, currency string
	}{{MarketKR, KRContinuationLaneID, "KRW"}, {MarketUS, USContinuationLaneID, "USD"}} {
		p := mustPlan(t, c.market, c.lane, c.currency, c.currency, nil, 14, "100")
		r := NewRiskState(p)
		r.FilledMinor = "110"
		r.HeldMinor = "4"
		r.Latches[LatchUnknownActualRisk] = true
		r.Latches[LatchCampaignRiskOverage] = false
		r.Fills["unknown"] = FillRiskRecord{Applied: true, Known: false, Fingerprint: "f"}
		r.Cancels["cancel"] = CancelRiskRecord{Applied: true, Fingerprint: "c"}
		e, err := CaptureOfflineEvidence(p, r, time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC))
		if err != nil {
			t.Fatal(err)
		}
		b, err := evidence.Encode(e)
		if err != nil {
			t.Fatal(err)
		}
		path := filepath.Join(t.TempDir(), "state.json")
		if err := os.WriteFile(path, b, 0600); err != nil {
			t.Fatal(err)
		}
		reopened, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(b)
		decoded, err := evidence.Decode(reopened, hex.EncodeToString(sum[:]), e.Plan.Scope)
		if err != nil || decoded.Plan.Digest != e.Plan.Digest || decoded.Risk.Filled != "110" {
			t.Fatalf("reopen: %+v %v", decoded, err)
		}
		r.Latches[LatchUnknownActualRisk] = false
		delete(r.Fills, "unknown")
		if !e.Risk.Latches[string(LatchUnknownActualRisk)] || e.Risk.ContinuationFills["unknown"].Fingerprint != "f" || e.Plan.Digest != p.RiskBudgetDigest || e.Risk.Filled != "110" || len(b) == 0 {
			t.Fatal("snapshot lost")
		}
	}
}

func TestOfflineEvidenceCopiesContinuationFX(t *testing.T) {
	f := validUSFX()
	p := mustPlan(t, MarketUS, USContinuationLaneID, "KRW", "USD", &f, 14, "100")
	r := NewRiskState(p)
	e, err := CaptureOfflineEvidence(p, r, time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if e.Plan.FX == nil || e.Plan.FX.Seal == "" || e.Plan.FX.QuoteID != f.QuoteID || e.Plan.FX.Official == nil || !*e.Plan.FX.Official {
		t.Fatal("FX not preserved")
	}
	old := e.Plan.FX.Rate
	p.FX.RateQuoteToAccount = "99"
	if e.Plan.FX.Rate != old {
		t.Fatal("FX aliases source")
	}
}

func TestOfflineEvidenceFromNativeFillAndCancel(t *testing.T) {
	p := mustPlan(t, MarketKR, KRContinuationLaneID, "KRW", "KRW", nil, 14, "100")
	r := NewRiskState(p)
	r.HeldMinor = "60"
	r, result := ApplyFillRisk(r, p, scopedFillRiskEvent(p, FillRiskEvent{FillID: "fill-1", Quantity: 2, TransferredReservationMinor: "40", EntryPriceMinor: "100", EffectiveStopMinor: "90", EntryFeesMinor: "3", EstimatedExitFeesLeviesMinor: "2"}))
	if !result.Applied {
		t.Fatal(result)
	}
	r, result = ApplyCancelRisk(r, p, scopedCancelRiskEvent(p, CancelRiskEvent{CancelID: "cancel-1", ReleaseHeldMinor: "20"}))
	if !result.Applied {
		t.Fatal(result)
	}
	e, err := CaptureOfflineEvidence(p, r, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if e.Risk.Filled != "40" || e.Risk.Held != "0" || !e.Risk.ContinuationFills["fill-1"].Applied || !e.Risk.ContinuationCancels["cancel-1"].Applied {
		t.Fatalf("lost native risk %+v", e.Risk)
	}
	p.PositionGeneration = -1
	p.RiskBudgetDigest = planDigest(p)
	r.RiskBudgetDigest = p.RiskBudgetDigest
	if e, err := CaptureOfflineEvidence(p, r, time.Now()); err == nil || e.Schema != "" {
		t.Fatal("negative generation accepted")
	}
}

func TestOfflineEvidenceRejectsUnencodableNativeText(t *testing.T) {
	p := mustPlan(t, MarketKR, KRContinuationLaneID, "KRW", "KRW", nil, 14, "100")
	r := NewRiskState(p)
	r.Fills[string([]byte{0xff})] = FillRiskRecord{Applied: true, Fingerprint: "record"}
	if e, err := CaptureOfflineEvidence(p, r, time.Now()); err == nil || e.Schema != "" {
		t.Fatal("invalid UTF-8 native map key accepted")
	}
	delete(r.Fills, string([]byte{0xff}))
	r.Fills["f"] = FillRiskRecord{Applied: true, Fingerprint: string([]byte{0xff})}
	if e, err := CaptureOfflineEvidence(p, r, time.Now()); err == nil || e.Schema != "" {
		t.Fatal("invalid UTF-8 native field accepted")
	}
}

func TestOfflineCeilingsMatchNativeAllocator(t *testing.T) {
	p := mustPlan(t, MarketKR, KRContinuationLaneID, "KRW", "KRW", nil, 14, "100")
	e, err := CaptureOfflineEvidence(p, NewRiskState(p), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	for _, q := range []uint64{1, 2, 13, 14, 15, ^uint64(0)} {
		a := AllocateEightFourTwo(q)
		e.Plan.Quantity = q
		e.Plan.Ceilings = append([]uint64(nil), a[:]...)
		if _, err := evidence.Encode(e); err != nil {
			t.Fatalf("%d: %v", q, err)
		}
	}
}
