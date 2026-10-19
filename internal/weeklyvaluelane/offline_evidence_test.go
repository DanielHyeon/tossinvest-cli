package weeklyvaluelane

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"
	"time"

	evidence "github.com/JungHoonGhae/tossinvest-cli/internal/campaignstateevidence"
)

func TestOfflineEvidencePreservesWeekly(t *testing.T) {
	for _, c := range []struct {
		market         Market
		lane, currency string
	}{{MarketKR, KRWeeklyLaneID, "KRW"}, {MarketUS, USWeeklyLaneID, "USD"}} {
		p := mustPlan(t, c.market, c.lane, c.currency, c.currency, nil, 14, "100")
		r := NewRiskState(p)
		r.filledMinor = "110"
		r.heldMinor = "4"
		r.latches[LatchUnknownActualRisk] = true
		r.latches[LatchCampaignRiskOverage] = false
		r.fills["unknown"] = AppliedFill{Known: true, Fingerprint: "f"}
		r.seal = riskStateSeal(r)
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
		delete(r.fills, "unknown")
		if !e.Risk.WeeklyFills["unknown"].KnownIdentity || !e.Risk.Latches[string(LatchUnknownActualRisk)] || e.Risk.CancelStatus != "not_applicable" || e.Risk.ReservationStatus != "not_captured" || e.Plan.Digest != p.Digest() || e.Risk.Filled != "110" {
			t.Fatal("snapshot lost")
		}
	}
}

func TestOfflineEvidenceCopiesWeeklyFX(t *testing.T) {
	f := validFX()
	p := mustPlan(t, MarketUS, USWeeklyLaneID, "KRW", "USD", &f, 14, "100")
	r := NewRiskState(p)
	e, err := CaptureOfflineEvidence(p, r, time.Date(2026, 10, 19, 0, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	if e.Plan.FX == nil || e.Plan.FX.Seal == "" || e.Plan.FX.FreshUntil != f.freshUntil.Format(time.RFC3339Nano) {
		t.Fatal("FX not preserved")
	}
	old := e.Plan.FX.Rate
	p.fx.rateQuoteToAccount = "99"
	if e.Plan.FX.Rate != old {
		t.Fatal("FX aliases source")
	}
}

func TestOfflineEvidenceFromNativeFill(t *testing.T) {
	fx := validFX()
	p := mustPlan(t, MarketUS, USWeeklyLaneID, "KRW", "USD", &fx, 14, "100")
	r := mustRiskState(t, p, "0", "60")
	event := validFill(p, "fill-1")
	event.FX = &fx
	r, result := ApplyFillRisk(r, p, event)
	if !result.Applied {
		t.Fatal(result)
	}
	e, err := CaptureOfflineEvidence(p, r, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if e.Risk.Filled != "60" || e.Risk.Held != "0" || !e.Risk.WeeklyFills["fill-1"].KnownIdentity || e.Risk.ReservationStatus != "not_captured" {
		t.Fatalf("lost native risk %+v", e.Risk)
	}
}

func TestOfflineCeilingsMatchNativeAllocator(t *testing.T) {
	p := mustPlan(t, MarketKR, KRWeeklyLaneID, "KRW", "KRW", nil, 14, "100")
	e, err := CaptureOfflineEvidence(p, NewRiskState(p), time.Now())
	if err != nil {
		t.Fatal(err)
	}
	for _, q := range []uint64{1, 2, 6, 7, 8, ^uint64(0)} {
		a := AllocateSeven(q)
		e.Plan.Quantity = q
		e.Plan.Ceilings = append([]uint64(nil), a[:]...)
		if _, err := evidence.Encode(e); err != nil {
			t.Fatalf("%d: %v", q, err)
		}
	}
	e.Plan.Quantity = 14
	e.Plan.Ceilings = []uint64{14, 0, 0, 0, 0, 0, 0}
	if _, err := evidence.Encode(e); err == nil {
		t.Fatal("redistribution accepted")
	}
}
