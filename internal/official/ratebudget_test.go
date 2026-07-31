package official

import (
	"net/http"
	"testing"
	"time"
)

var budgetNow = time.Date(2026, 7, 28, 9, 0, 0, 0, time.UTC)

func header(pairs ...string) http.Header {
	h := http.Header{}
	for i := 0; i+1 < len(pairs); i += 2 {
		h.Set(pairs[i], pairs[i+1])
	}
	return h
}

// TestAResponseWithNoRateHeadersIsNotZeroRemaining is the absent-versus-zero rule.
//
// Not every endpoint or environment sends these headers. A caller that read a
// missing header as "0 calls remaining" would back off from an endpoint that
// never asked it to — and the whole reason for reading them is to stop guessing.
func TestAResponseWithNoRateHeadersIsNotZeroRemaining(t *testing.T) {
	got := readRateBudget("/api/v1/rankings", header(), budgetNow)
	if got.Reported {
		t.Fatal("a response with no rate headers reports a budget")
	}
	if got.Exhausted() {
		t.Error("an unreported budget reads as exhausted")
	}
	if got.Tight(5) {
		t.Error("an unreported budget reads as tight")
	}
}

func TestTheReportedBudgetIsKept(t *testing.T) {
	got := readRateBudget("/api/v1/rankings", header(
		HeaderRateLimit, "10",
		HeaderRateRemaining, "3",
	), budgetNow)

	if !got.Reported {
		t.Fatal("a response with rate headers does not report a budget")
	}
	if got.Limit != 10 || got.Remaining != 3 {
		t.Errorf("budget = %d/%d, want 3/10", got.Remaining, got.Limit)
	}
	if got.Exhausted() {
		t.Error("3 remaining reads as exhausted")
	}
	if !got.Tight(3) || got.Tight(2) {
		t.Errorf("Tight is wrong at the boundary: Tight(3)=%v Tight(2)=%v",
			got.Tight(3), got.Tight(2))
	}
	if got.Path != "/api/v1/rankings" {
		t.Errorf("path = %q", got.Path)
	}
}

// TestTheResetHeaderIsReadBothWaysAndSaysWhich.
//
// Rate-limit resets are spelled as epoch seconds by some APIs and as seconds-from-
// now by others, and the Toss documentation says neither. Picking one silently is
// how a measurement becomes a new assumption, so the reading is labelled and the
// raw value is kept beside it.
func TestTheResetHeaderIsReadBothWaysAndSaysWhich(t *testing.T) {
	epoch := budgetNow.Add(30 * time.Second).Unix()

	for _, tc := range []struct {
		name     string
		raw      string
		wantKind ResetKind
		wantTime time.Time
	}{
		{"delta seconds", "30", ResetDelta, budgetNow.Add(30 * time.Second)},
		{"epoch seconds", itoa(epoch), ResetEpoch, time.Unix(epoch, 0).UTC()},
		{"nonsense", "soon", ResetUnparsed, time.Time{}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got := readRateBudget("/api/v1/rankings", header(
				HeaderRateRemaining, "3",
				HeaderRateReset, tc.raw,
			), budgetNow)

			if got.ResetKind != tc.wantKind {
				t.Errorf("ResetKind = %q, want %q", got.ResetKind, tc.wantKind)
			}
			if !got.Reset.Equal(tc.wantTime) {
				t.Errorf("Reset = %v, want %v", got.Reset, tc.wantTime)
			}
			if got.ResetRaw != tc.raw {
				t.Errorf("ResetRaw = %q, want the header verbatim %q", got.ResetRaw, tc.raw)
			}
		})
	}
}

// TestOnlyAResetHeaderStillCountsAsReported: a response that sends Reset alone has
// said something about the budget, and dropping it would be the same information
// loss this file exists to stop.
func TestOnlyAResetHeaderStillCountsAsReported(t *testing.T) {
	got := readRateBudget("/api/v1/prices", header(HeaderRateReset, "5"), budgetNow)
	if !got.Reported {
		t.Fatal("a lone Reset header is not counted as a report")
	}
	if got.ResetKind != ResetDelta {
		t.Errorf("ResetKind = %q, want %q", got.ResetKind, ResetDelta)
	}
}

// TestTheBudgetStoreIsPerPath keeps two endpoints from overwriting each other.
// The limits are per group and the group is not in the response, so the path is
// the finest distinction actually available — one shared slot would make the
// ranking's budget read whatever the last price call happened to say.
func TestTheBudgetStoreIsPerPath(t *testing.T) {
	var store rateBudgets
	store.record(readRateBudget("/api/v1/rankings", header(HeaderRateRemaining, "2"), budgetNow))
	store.record(readRateBudget("/api/v1/prices", header(HeaderRateRemaining, "9"), budgetNow))

	if got := store.get("/api/v1/rankings"); got.Remaining != 2 {
		t.Errorf("rankings remaining = %d, want 2", got.Remaining)
	}
	if got := store.get("/api/v1/prices"); got.Remaining != 9 {
		t.Errorf("prices remaining = %d, want 9", got.Remaining)
	}
	if got := store.get("/api/v1/never-called"); got.Reported {
		t.Error("a path that was never called reports a budget")
	}
	if all := store.all(); len(all) != 2 {
		t.Errorf("all() returned %d entries, want 2", len(all))
	}
}

func itoa(v int64) string {
	const digits = "0123456789"
	if v == 0 {
		return "0"
	}
	var buf []byte
	for v > 0 {
		buf = append([]byte{digits[v%10]}, buf...)
		v /= 10
	}
	return string(buf)
}

// TestTheBudgetKeyCollapsesObjectIdentifiers.
//
// Order cancels, order modifies and per-symbol candles all embed an id in the
// path. Keyed raw, a long-lived engine client accumulates one entry per order id
// for the life of the process — and, worse for the purpose this serves, the
// candle budget is split across every symbol, so the per-group allowance the
// server actually meters is never observable.
func TestTheBudgetKeyCollapsesObjectIdentifiers(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"/api/v1/orders/OsBakhtsu9afmp1j2/cancel", "/api/v1/orders/{id}/cancel"},
		{"/api/v1/orders/abc123/modify", "/api/v1/orders/{id}/modify"},
		{"/api/v1/market-indicators/KOSPI/candles", "/api/v1/market-indicators/{id}/candles"},
		{"/api/v1/rankings", "/api/v1/rankings"},
		{"/api/v1/prices", "/api/v1/prices"},
	} {
		if got := budgetKey(tc.in); got != tc.want {
			t.Errorf("budgetKey(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestTheBudgetMapDoesNotGrowWithOrderIDs is the same point as a bound.
func TestTheBudgetMapDoesNotGrowWithOrderIDs(t *testing.T) {
	var store rateBudgets
	for i := 0; i < 500; i++ {
		path := "/api/v1/orders/" + itoa(int64(i)) + "/cancel"
		store.record(readRateBudget(path, header(HeaderRateRemaining, "5"), budgetNow))
	}
	if n := len(store.all()); n != 1 {
		t.Fatalf("500 distinct order ids produced %d budget entries, want 1", n)
	}
}

// TestAnImplausibleResetIsNotPresentedAsADetermination.
//
// The heuristic used to answer confidently on inputs it could not actually read:
// epoch milliseconds — a very common encoding — became a reset in the year 58541,
// and a nanosecond delay became one in 2001. Both were labelled ResetEpoch. A
// caller waiting for a reset in 2001 does not wait at all; one waiting for 58541
// never stops.
func TestAnImplausibleResetIsNotPresentedAsADetermination(t *testing.T) {
	for _, raw := range []string{
		"1785229260000", // epoch milliseconds
		"1000000000",    // a 1s delay encoded in nanoseconds
		"-30",           // a reset in the past
		"999999999999999",
	} {
		got := readRateBudget("/api/v1/rankings", header(
			HeaderRateRemaining, "3", HeaderRateReset, raw,
		), budgetNow)
		if got.ResetKind != ResetUnparsed {
			t.Errorf("reset %q was read as %q giving %v; want unparsed",
				raw, got.ResetKind, got.Reset)
		}
		if got.ResetRaw != raw {
			t.Errorf("reset %q lost its raw value", raw)
		}
		if !got.Reported {
			t.Errorf("reset %q made the whole budget unreported", raw)
		}
	}
}

func TestParseRateBudgetResetUsesExactThresholdAndPlausibilityBoundaries(t *testing.T) {
	now := budgetNow
	epochThreshold := time.Unix(1_000_000_000, 0).UTC()
	tests := []struct {
		name      string
		raw       string
		observed  time.Time
		wantRaw   string
		wantReset time.Time
		wantKind  ResetKind
	}{
		{name: "canonical whitespace", raw: " 60 ", observed: now, wantRaw: "60", wantReset: now.Add(time.Minute), wantKind: ResetDelta},
		{name: "delta max ahead inclusive", raw: "86400", observed: now, wantRaw: "86400", wantReset: now.Add(24 * time.Hour), wantKind: ResetDelta},
		{name: "delta beyond max ahead", raw: "86401", observed: now, wantRaw: "86401", wantKind: ResetUnparsed},
		{name: "exact threshold is epoch", raw: "1000000000", observed: epochThreshold, wantRaw: "1000000000", wantReset: epochThreshold, wantKind: ResetEpoch},
		{name: "epoch max behind inclusive", raw: itoa(now.Add(-time.Minute).Unix()), observed: now, wantRaw: itoa(now.Add(-time.Minute).Unix()), wantReset: now.Add(-time.Minute), wantKind: ResetEpoch},
		{name: "epoch beyond max behind", raw: itoa(now.Add(-time.Minute - time.Second).Unix()), observed: now, wantRaw: itoa(now.Add(-time.Minute - time.Second).Unix()), wantKind: ResetUnparsed},
		{name: "epoch max ahead inclusive", raw: itoa(now.Add(24 * time.Hour).Unix()), observed: now, wantRaw: itoa(now.Add(24 * time.Hour).Unix()), wantReset: now.Add(24 * time.Hour), wantKind: ResetEpoch},
		{name: "epoch beyond max ahead", raw: itoa(now.Add(24*time.Hour + time.Second).Unix()), observed: now, wantRaw: itoa(now.Add(24*time.Hour + time.Second).Unix()), wantKind: ResetUnparsed},
		{name: "duration wrapping integer", raw: "36028797018963969", observed: now, wantRaw: "36028797018963969", wantKind: ResetUnparsed},
		{name: "zero observed at", raw: "60", wantRaw: "60", wantKind: ResetUnparsed},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			raw, reset, kind := ParseRateBudgetReset(tc.raw, tc.observed)
			if raw != tc.wantRaw || kind != tc.wantKind || !reset.Equal(tc.wantReset) {
				t.Fatalf("ParseRateBudgetReset(%q) = (%q, %s, %q), want (%q, %s, %q)",
					tc.raw, raw, reset, kind, tc.wantRaw, tc.wantReset, tc.wantKind)
			}
		})
	}
}

func TestAddResetDeltaRejectsDurationConversionOverflow(t *testing.T) {
	maxDurationSeconds := int64((time.Duration(1<<63 - 1)) / time.Second)
	if reset, ok := addResetDelta(budgetNow, maxDurationSeconds+1); ok || !reset.IsZero() {
		t.Fatalf("overflowing delta derived reset %s with ok=%v", reset, ok)
	}
}
