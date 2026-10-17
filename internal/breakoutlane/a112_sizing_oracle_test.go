package breakoutlane

// a112 breakout 덮개 2차 — 태스크 2.4 · 2.4.1 의 속성 시험(Manager 승인 2026-10-01). 예제 시험은 고른 값만 잰다; 여기서는 결정적 난수로
// 넓은 입력(작은 값 · 2^63 근처 · 128 비트 곱 경로 · 넘침)을 만들어 math/big 신탁과 비교한다.
//   (1) 거절 판정 신탁: 골든 fx_and_sizing(rounding · worst_entry · risk_per_share · refuse)과 코드의 거절 **순서**를 big.Int 로 옮긴 것 —
//       순서는 코드에서 옮겼으므로(설계 선택) 이 신탁이 증명하는 것은 「넘침이 감싸지 않고 거절이 된다」와 「각 단계 값이 정확하다」다.
//   (2) 독립 상한: 수락된 결과마다 q_candidate = floor(min(budget_acct·num/(den·risk), notional_acct·num/(den·worst)))(정확한 유리수 바닥)이고
//       0 <= q_final <= q_candidate, q_final <= FinalCap. 이 식은 코드의 두 단계(바닥 환산 뒤 정수 나눗셈)와 **다른 경로**로 계산한다.
//   (3) 호가 판정 신탁: 골든 quote 의 spread · drift 식과 포함 경계(한계와 같으면 수락, 한 단위 넘으면 거절).
// risk 의 entry 는 골든 문장이 「entry-stop」 이지만 코드는 worst_entry(max(entry, ask)+slippage) 기준이다 — 더 큰 위험 = 더 작은 수량(보수),
// 기존 판정 `TestFinalRedTeamSizingUsesWorstEntryInRisk` 와 같은 해석을 신탁이 따른다.

import (
	"math/big"
	"math/rand/v2"
	"testing"
)

var a112Two64 = new(big.Int).Lsh(big.NewInt(1), 64)

func a112Big(v uint64) *big.Int { return new(big.Int).SetUint64(v) }
func a112Fits(v *big.Int) bool  { return v.Sign() >= 0 && v.Cmp(a112Two64) < 0 }

// a112MulDivFloor · Ceil 는 정확한 a·b/d 의 바닥 · 천장이다.
func a112MulDiv(a, b, d uint64, ceil bool) *big.Int {
	n := new(big.Int).Mul(a112Big(a), a112Big(b))
	q, r := new(big.Int).QuoRem(n, a112Big(d), new(big.Int))
	if ceil && r.Sign() != 0 {
		q.Add(q, big.NewInt(1))
	}
	return q
}

// a112Draw 는 크기 등급을 섞어 뽑는다 — 작은 값 · 중간 · 2^63 근처 · 최댓값 근처. 넘침 · 128 비트 경로는 큰 등급에서만 열린다.
func a112Draw(r *rand.Rand) uint64 {
	switch r.IntN(6) {
	case 0:
		return r.Uint64N(16)
	case 1:
		return r.Uint64N(10_000)
	case 2:
		return r.Uint64N(1 << 32)
	case 3:
		return 1<<63 + r.Uint64N(1<<62)
	case 4:
		return maxUint64 - r.Uint64N(1_000)
	default:
		return r.Uint64()
	}
}

type a112SizingCase struct {
	in       SizingInput
	ask      uint64
	num, den uint64
}

func (c a112SizingCase) oracle() (RefusalCode, uint64, uint64) {
	in := c.in
	if in.ProposedEntryMinor == 0 || c.ask == 0 || in.StopMinor == 0 || in.StopMinor >= in.ProposedEntryMinor {
		return RefusalNonProtectiveStop, 0, 0
	}
	if in.TargetMinor <= in.ProposedEntryMinor {
		return RefusalNonProtectiveTarget, 0, 0
	}
	roundTrip := a112MulDiv(in.RoundTripCostAccountMinor, c.num, c.den, true)
	if !a112Fits(roundTrip) {
		return RefusalSizingOverflow, 0, 0
	}
	base := max(in.ProposedEntryMinor, c.ask)
	worst := new(big.Int).Add(a112Big(base), a112Big(in.EntrySlippageMinor))
	if !a112Fits(worst) {
		return RefusalSizingOverflow, 0, 0
	}
	risk := new(big.Int).Sub(worst, a112Big(in.StopMinor))
	risk.Add(risk, a112Big(in.ExitSlippageMinor)).Add(risk, roundTrip)
	if !a112Fits(risk) {
		return RefusalSizingOverflow, 0, 0
	}
	costExit := new(big.Int).Add(worst, a112Big(in.ExitSlippageMinor))
	costExit.Add(costExit, roundTrip)
	if !a112Fits(costExit) {
		return RefusalSizingOverflow, 0, 0
	}
	target := a112Big(in.TargetMinor)
	if target.Cmp(costExit) <= 0 {
		return RefusalNonProtectiveTarget, 0, 0
	}
	if in.MinRiskRewardPPM > 0 {
		need := new(big.Int).Mul(risk, a112Big(in.MinRiskRewardPPM))
		need.Add(need, big.NewInt(v1PPMScale-1)).Quo(need, big.NewInt(v1PPMScale))
		if !a112Fits(need) || new(big.Int).Sub(target, costExit).Cmp(need) < 0 {
			return RefusalNonProtectiveTarget, 0, 0
		}
	}
	budget := a112MulDiv(in.RiskBudgetAccountMinor, c.num, c.den, false)
	if !a112Fits(budget) {
		return RefusalSizingOverflow, 0, 0
	}
	notional := a112MulDiv(in.NotionalCapAccountMinor, c.num, c.den, false)
	if !a112Fits(notional) {
		return RefusalSizingOverflow, 0, 0
	}
	candidate := new(big.Int).Quo(budget, risk)
	if n := new(big.Int).Quo(notional, worst); n.Cmp(candidate) < 0 {
		candidate = n
	}
	if candidate.Sign() == 0 {
		return RefusalZeroQuantity, 0, 0
	}
	final := min(candidate.Uint64(), in.FinalCap)
	if final == 0 {
		return RefusalZeroQuantity, 0, 0
	}
	return RefusalNone, candidate.Uint64(), final
}

// a112ExactCandidateFloor 는 코드와 다른 경로의 정확한 유리수 바닥이다: floor(min(B·num/(den·risk), N·num/(den·worst))).
func a112ExactCandidateFloor(c a112SizingCase, risk, worst uint64) *big.Int {
	byRisk := new(big.Rat).SetFrac(new(big.Int).Mul(a112Big(c.in.RiskBudgetAccountMinor), a112Big(c.num)), new(big.Int).Mul(a112Big(c.den), a112Big(risk)))
	byNotional := new(big.Rat).SetFrac(new(big.Int).Mul(a112Big(c.in.NotionalCapAccountMinor), a112Big(c.num)), new(big.Int).Mul(a112Big(c.den), a112Big(worst)))
	least := byRisk
	if byNotional.Cmp(byRisk) < 0 {
		least = byNotional
	}
	return new(big.Int).Quo(least.Num(), least.Denom())
}

func a112SizingSeal(t *testing.T, num, den uint64) FXSeal {
	t.Helper()
	i := FXSealInput{AccountCurrency: "KRW", InstrumentCurrency: "USD", Direction: FXAccountToInstrument, RateNum: num, RateDen: den, Scale: 2, AsOfMS: 1, FreshUntilMS: 10}
	i.Digest = FXSealDigest(i)
	f, err := NewFXSeal(i)
	if err != nil {
		t.Fatal(err)
	}
	return f
}

func TestSizingMatchesTheExactRationalOracle(t *testing.T) {
	r := rand.New(rand.NewPCG(0xa112, 0x24))
	const n = 20_000
	seen := map[RefusalCode]int{}
	wide := 0 // 128 비트 곱이 실제로 쓰인(hi != 0) 수락 사례
	for k := 0; k < n; k++ {
		c := a112SizingCase{ask: 1 + a112Draw(r)%maxUint64, num: 1 + a112Draw(r)%maxUint64, den: 1 + a112Draw(r)%maxUint64}
		if c.ask == 0 {
			c.ask = 1
		}
		entry := a112Draw(r)
		c.in = SizingInput{
			ProposedEntryMinor: entry, StopMinor: a112Draw(r), TargetMinor: a112Draw(r),
			EntrySlippageMinor: a112Draw(r) % 64, ExitSlippageMinor: a112Draw(r) % 64, RoundTripCostAccountMinor: a112Draw(r),
			RiskBudgetAccountMinor: a112Draw(r), NotionalCapAccountMinor: a112Draw(r), FinalCap: a112Draw(r), MinRiskRewardPPM: a112Draw(r) % 4_000_000,
		}
		if entry > 0 && r.IntN(2) == 0 { // stop 의 절반은 entry 아래로
			c.in.StopMinor %= entry
		}
		if r.IntN(3) == 0 { // 수락 경로가 충분히 열리게: 작은 환율 · 보호 stop · 위쪽 target
			c.num, c.den = 1+r.Uint64N(1<<20), 1+r.Uint64N(1<<20)
			c.in.StopMinor = 1 + r.Uint64N(max(entry, 2)-1)
			c.in.TargetMinor = entry + r.Uint64N(1<<40)
			c.in.RoundTripCostAccountMinor %= 1 << 20
		}
		quote := QuoteSeal{value: QuoteSealInput{BidMinor: 1, AskMinor: c.ask, LastMinor: c.ask, Currency: "USD"}}
		got := size(c.in, quote, a112SizingSeal(t, c.num, c.den), 5)
		want, candidate, final := c.oracle()
		seen[want]++
		if got.Refusal != want || got.CandidateQuantity != candidate || got.FinalQuantity != final {
			t.Fatalf("case %d %+v: size=%+v, oracle refusal=%q candidate=%d final=%d", k, c, got, want, candidate, final)
		}
		if want != RefusalNone {
			continue
		}
		exact := a112ExactCandidateFloor(c, got.RiskPerShareMinor, got.WorstEntryMinor)
		if !exact.IsUint64() || exact.Uint64() != got.CandidateQuantity || got.FinalQuantity > got.CandidateQuantity || got.FinalQuantity > c.in.FinalCap {
			t.Fatalf("case %d %+v: candidate=%d final=%d, exact rational floor=%s", k, c, got.CandidateQuantity, got.FinalQuantity, exact)
		}
		if hi, _ := a112MulHi(c.in.RiskBudgetAccountMinor, c.num); hi != 0 {
			wide++
		}
	}
	// 분포 대조: 모든 거절 갈래와 수락, 그리고 128 비트 곱 경로가 실제로 뽑혔다 — 안 그러면 신탁 비교가 그 갈래를 안 잰 것이다.
	for _, code := range []RefusalCode{RefusalNone, RefusalNonProtectiveStop, RefusalNonProtectiveTarget, RefusalSizingOverflow, RefusalZeroQuantity} {
		if seen[code] < 50 {
			t.Errorf("only %d cases reached %q out of %d (distribution %v)", seen[code], code, n, seen)
		}
	}
	t.Logf("distribution %v, 128-bit accepted %d", seen, wide)
	if wide < 50 {
		t.Errorf("only %d accepted cases used the 128-bit product path", wide)
	}
}

func a112MulHi(a, b uint64) (uint64, uint64) {
	p := new(big.Int).Mul(a112Big(a), a112Big(b))
	return new(big.Int).Rsh(p, 64).Uint64(), new(big.Int).And(p, new(big.Int).Sub(a112Two64, big.NewInt(1))).Uint64()
}

func TestMulDivMatchesTheExactQuotientOrRefuses(t *testing.T) {
	r := rand.New(rand.NewPCG(0xa112, 0x241))
	for k := 0; k < 50_000; k++ {
		a, b, d := a112Draw(r), a112Draw(r), a112Draw(r)
		floor, ceil := a112MulDiv(a, b, max(d, 1), false), a112MulDiv(a, b, max(d, 1), true)
		gotFloor, refusal := mulDivFloor(a, b, d)
		gotCeil, overflow := mulDivCeil(a, b, d)
		switch {
		case d == 0:
			if refusal != RefusalFXInvalidRate || !overflow {
				t.Fatalf("d=0: floor refusal=%q ceil overflow=%v", refusal, overflow)
			}
		default:
			if a112Fits(floor) != (refusal == RefusalNone) || refusal == RefusalNone && gotFloor != floor.Uint64() || refusal != RefusalNone && refusal != RefusalSizingOverflow {
				t.Fatalf("floor(%d·%d/%d)=%s: got %d %q", a, b, d, floor, gotFloor, refusal)
			}
			if a112Fits(ceil) == overflow || !overflow && gotCeil != ceil.Uint64() {
				t.Fatalf("ceil(%d·%d/%d)=%s: got %d overflow=%v", a, b, d, ceil, gotCeil, overflow)
			}
		}
	}
	// 천장이 정확히 2^64 가 되는 경계(바닥 = 최댓값, 나머지 있음) — q == maxUint64 갈래.
	if _, overflow := mulDivCeil(maxUint64, 3, 2); !overflow {
		t.Fatal("ceil(max·3/2) fits only if the q==max branch is skipped")
	}
	if q, overflow := mulDivCeil(maxUint64, 2, 2); overflow || q != maxUint64 {
		t.Fatalf("ceil(max·2/2)=%d overflow=%v, want max exactly", q, overflow)
	}
}

func TestQuoteVetoMatchesTheGoldenFormulasAtAndBeyondEachLimit(t *testing.T) {
	r := rand.New(rand.NewPCG(0xa112, 0x2411))
	million := big.NewInt(v1PPMScale)
	ceilDiv := func(n, d *big.Int) *big.Int {
		q, rem := new(big.Int).QuoRem(n, d, new(big.Int))
		if rem.Sign() != 0 {
			q.Add(q, big.NewInt(1))
		}
		return q
	}
	reached := map[RefusalCode]int{}
	for k := 0; k < 20_000; k++ {
		bid := 1 + a112Draw(r)%(maxUint64-1)
		ask := bid + a112Draw(r)%(maxUint64-bid+1)
		if r.IntN(2) == 0 { // 실제 호가 모양(좁은 스프레드)
			bid = 1 + r.Uint64N(1<<40)
			ask = bid + r.Uint64N(bid/50+1)
		}
		entry := 1 + a112Draw(r)%maxUint64
		if r.IntN(2) == 0 {
			entry = max(1, ask-r.Uint64N(ask/20+1)+r.Uint64N(ask/20+1))
		}
		spread := ceilDiv(new(big.Int).Mul(a112Big(ask-bid), million), new(big.Int).Rsh(new(big.Int).Add(a112Big(ask), a112Big(bid)), 1))
		drift := ceilDiv(new(big.Int).Mul(a112Big(absoluteDifference(ask, entry)), million), a112Big(entry))
		quote := QuoteSealInput{BidMinor: bid, AskMinor: ask, LastMinor: ask, Currency: "USD", SourceObservedAtMS: 5, ReceivedAtMS: 6}
		quote.Digest = QuoteSealDigest(quote)
		seal, err := NewQuoteSeal(quote)
		if err != nil {
			t.Fatal(err)
		}
		if !a112Fits(spread) {
			t.Fatalf("spread %s does not fit — the overflow guard would be reachable", spread)
		}
		limits := func(spreadLimit, driftLimit uint64) V1Config {
			in := V1ConfigInput{Version: "v1", TickMinor: 1, OpeningRangeMinutes: 15, BreakoutBufferPPM: 100_000, RetestTolerancePPM: 100_000, TimeoutKR: 8, TimeoutUS: 10,
				RVOLMinPPM: 1_500_000, UpperWickRangeMaxPPM: 350_000, MaxQuoteAgeMS: 5, MaxSpreadPPM: max(spreadLimit, 1), MaxEntryDriftPPM: max(driftLimit, 1)}
			in.Digest = V1ConfigDigest(in)
			c, err := NewV1Config(in)
			if err != nil {
				t.Fatal(err)
			}
			return c
		}
		if !a112Fits(drift) {
			if got := validateQuote(seal, 10, entry, limits(maxUint64, maxUint64)); got != RefusalSizingOverflow {
				t.Fatalf("drift %s overflows: got %q, want SIZING_OVERFLOW", drift, got)
			}
			reached[RefusalSizingOverflow]++
			continue
		}
		s, d := spread.Uint64(), drift.Uint64()
		// 한계와 같으면 수락(한계는 양수여야 하므로 0 인 값은 1 로 올린다 — 여전히 값 이상).
		if got := validateQuote(seal, 10, entry, limits(s, d)); got != RefusalNone {
			t.Fatalf("bid=%d ask=%d entry=%d spread=%d drift=%d at the limits: %q", bid, ask, entry, s, d, got)
		}
		reached[RefusalNone]++
		if s > 1 {
			if got := validateQuote(seal, 10, entry, limits(s-1, maxUint64)); got != RefusalSpreadTooWide {
				t.Fatalf("spread %d over limit %d: %q", s, s-1, got)
			}
			reached[RefusalSpreadTooWide]++
		}
		if d > 1 {
			if got := validateQuote(seal, 10, entry, limits(maxUint64, d-1)); got != RefusalEntryDriftExceeded {
				t.Fatalf("drift %d over limit %d (ask=%d entry=%d): %q", d, d-1, ask, entry, got)
			}
			reached[RefusalEntryDriftExceeded]++
		}
	}
	t.Logf("quote distribution %v", reached)
	for _, code := range []RefusalCode{RefusalNone, RefusalSpreadTooWide, RefusalEntryDriftExceeded, RefusalSizingOverflow} {
		if reached[code] < 50 {
			t.Errorf("%q reached %d times (distribution %v)", code, reached[code], reached)
		}
	}
	// 시각 순서(골든 timestamp_order: source <= received <= evaluated)와 나이 포함 경계 — 각 위반이 QUOTE_STALE.
	for _, tc := range []struct {
		name                       string
		source, received, evaluate uint64
		want                       RefusalCode
	}{
		{"age exactly the limit", 5, 6, 10, RefusalNone},
		{"age one over the limit", 4, 6, 10, RefusalQuoteStale},
		{"source after received", 7, 6, 10, RefusalQuoteStale},
		{"received after evaluated", 5, 11, 10, RefusalQuoteStale},
		{"all equal", 10, 10, 10, RefusalNone},
	} {
		q := QuoteSealInput{BidMinor: 100, AskMinor: 101, LastMinor: 101, Currency: "USD", SourceObservedAtMS: tc.source, ReceivedAtMS: tc.received}
		q.Digest = QuoteSealDigest(q)
		seal, err := NewQuoteSeal(q)
		if err != nil {
			t.Fatal(err)
		}
		in := V1ConfigInput{Version: "v1", TickMinor: 1, OpeningRangeMinutes: 15, BreakoutBufferPPM: 100_000, RetestTolerancePPM: 100_000, TimeoutKR: 8, TimeoutUS: 10,
			RVOLMinPPM: 1_500_000, UpperWickRangeMaxPPM: 350_000, MaxQuoteAgeMS: 5, MaxSpreadPPM: 100_000, MaxEntryDriftPPM: 100_000}
		in.Digest = V1ConfigDigest(in)
		c, _ := NewV1Config(in)
		if got := validateQuote(seal, tc.evaluate, 101, c); got != tc.want {
			t.Errorf("%s: %q, want %q", tc.name, got, tc.want)
		}
	}
}
