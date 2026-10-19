package reversallane

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
	"time"

	evidence "github.com/JungHoonGhae/tossinvest-cli/internal/campaignstateevidence"
)

func TestOfflineEvidencePreservesReversal(t *testing.T) {
	for _, c := range []struct {
		market         Market
		lane, currency string
	}{{MarketKR, KRReversalLaneID, "KRW"}, {MarketUS, USReversalLaneID, "USD"}} {
		p := mustPlan(t, c.market, c.lane, 14, "100", c.currency, c.currency, nil)
		r := NewRiskState(p)
		r.FilledMinor = "110"
		r.HeldMinor = "4"
		r.Latches[LatchUnknownActualRisk] = true
		r.Latches[LatchCampaignRiskOverage] = false
		r.Fills["unknown"] = AppliedFill{Applied: true, Fingerprint: "f"}
		r.Cancels["cancel"] = "c"
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
		delete(r.Fills, "unknown")
		r.Cancels["cancel"] = "new"
		if e.Risk.ReversalFills["unknown"].Fingerprint != "f" || e.Risk.ReversalCancels["cancel"] != "c" || e.Plan.Digest != p.Digest() || e.Risk.Filled != "110" {
			t.Fatal("snapshot lost")
		}
	}
}

func TestOfflineEvidenceCopiesReversalFX(t *testing.T) {
	f := validFX()
	p := mustPlan(t, MarketUS, USReversalLaneID, 14, "100", "KRW", "USD", &f)
	r := NewRiskState(p)
	e, err := CaptureOfflineEvidence(p, r, time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if e.Plan.FX == nil || e.Plan.FX.Seal != f.seal || e.Plan.FX.FreshUntil != f.FreshUntil {
		t.Fatal("FX not preserved")
	}
	old := e.Plan.FX.Rate
	p.request.FX.RateQuoteToAccount = "99"
	if e.Plan.FX.Rate != old {
		t.Fatal("FX aliases source")
	}
}

func TestOfflineEvidencePreservesNativeWhitespaceAmounts(t *testing.T) {
	req := validPlanRequest(MarketKR, KRReversalLaneID, 14, " 100 ", "KRW", "KRW", nil)
	p, err := BuildCampaignPlan(req)
	if err != nil {
		t.Fatal(err)
	}
	r := NewRiskState(p)
	r.FilledMinor = " 12 "
	e, err := CaptureOfflineEvidence(p, r, time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if e.Plan.Budget != " 100 " || e.Risk.Filled != " 12 " {
		t.Fatal("native amount normalized")
	}
}

func TestOfflineEvidenceFromNativeUnknownFillAndCancel(t *testing.T) {
	p := mustPlan(t, MarketKR, KRReversalLaneID, 14, "100", "KRW", "KRW", nil)
	r := NewRiskState(p)
	r.HeldMinor = "20"
	event := validFillEvent(p, "")
	r, result := ApplyFillRisk(r, p, event)
	if !result.Applied {
		t.Fatal(result)
	}
	r, result = ApplyCancelRisk(r, CancelRiskEvent{CancelID: "", CampaignID: p.CampaignID(), LegOrdinal: 1, OrderRef: "order-1", ReleaseHeldMinor: "10", ObservedAt: event.ObservedAt, SourceDigest: "cancel-source-1"})
	if !result.Applied {
		t.Fatal(result)
	}
	e, err := CaptureOfflineEvidence(p, r, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if len(e.Risk.ReversalFills) != 1 || len(e.Risk.ReversalCancels) != 1 || !e.Risk.Latches[string(LatchUnknownActualRisk)] || e.Risk.Held != "20" {
		t.Fatalf("lost unknown native risk %+v", e.Risk)
	}
}

func TestOfflineCeilingsMatchNativeAllocator(t *testing.T) {
	p := mustPlan(t, MarketKR, KRReversalLaneID, 14, "100", "KRW", "KRW", nil)
	e, err := CaptureOfflineEvidence(p, NewRiskState(p), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	for _, q := range []uint64{1, 2, 13, 14, 15, ^uint64(0)} {
		a := AllocateTwoFourEight(q)
		e.Plan.Quantity = q
		e.Plan.Ceilings = append([]uint64(nil), a[:]...)
		if _, err := evidence.Encode(e); err != nil {
			t.Fatalf("%d: %v", q, err)
		}
	}
	e.Plan.Quantity = 14
	e.Plan.Ceilings = []uint64{14, 0, 0}
	if _, err := evidence.Encode(e); err == nil {
		t.Fatal("redistribution accepted")
	}
}
