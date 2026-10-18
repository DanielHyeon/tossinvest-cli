package strategyengine

import (
	"math/big"
	"time"
)

type RRState string

const (
	RRKnown   RRState = "known"
	RRUnknown RRState = "unknown"
	RRInvalid RRState = "invalid"
)

// RRMetric은 판단 권한 없이 계산값과 그 기준을 함께 기록함.
type RRMetric struct {
	State                                                     RRState
	Ratio, Numerator, Denominator                             string
	NumeratorUnit, DenominatorUnit, PriceBasis, PolicyVersion string
	Provenance, Reason                                        string
}
type ParkerRRDiagnostics struct{ Space, Target, Realized RRMetric }

// ParkerRRFromDecision은 호출자가 제공한 결정 기록을 변경하지 않고 계산함.
func ParkerRRFromDecision(r DecisionRecord) ParkerRRDiagnostics {
	unit := ""
	if r.Currency != "" {
		unit = r.Currency + "/share"
	}
	basis := "unadjusted"
	if r.BarAdjusted {
		basis = "adjusted"
	}
	planned := RRMetric{State: RRUnknown, Reason: "missing planned input", NumeratorUnit: unit, DenominatorUnit: unit, PriceBasis: basis, PolicyVersion: r.LaneVersion, Provenance: "caller-supplied decision record; unverified"}
	result := ParkerRRDiagnostics{Space: planned, Target: planned, Realized: RRMetric{State: RRUnknown, Reason: "no realized evidence", PolicyVersion: r.LaneVersion}}
	if r.EntryPrice == "" || r.StopPrice == "" || r.Currency == "" || r.LaneVersion == "" {
		return result
	}
	entry, e := positive(r.EntryPrice)
	stop, s := positive(r.StopPrice)
	if !e || !s || entry.Cmp(stop) <= 0 || r.BarAdjusted || r.LaneID != LaneID || r.LaneVersion != LaneVersion {
		result.Space.State = RRInvalid
		result.Target.State = RRInvalid
		result.Space.Reason = "invalid planned input"
		result.Target.Reason = result.Space.Reason
		return result
	}
	risk := new(big.Rat).Sub(entry, stop)
	metric := func(raw string, space bool) RRMetric {
		out := planned
		out.Denominator = risk.RatString()
		out.Reason = "missing numerator"
		if raw == "" {
			return out
		}
		value, ok := positive(raw)
		if !ok {
			out.State = RRInvalid
			out.Reason = "invalid numerator"
			return out
		}
		numerator := new(big.Rat)
		if space {
			numerator.Quo(new(big.Rat).Mul(entry, value), big.NewRat(100, 1))
		} else {
			numerator.Sub(value, entry)
		}
		if numerator.Sign() <= 0 {
			out.State = RRInvalid
			out.Reason = "invalid numerator"
			return out
		}
		out.State = RRKnown
		out.Reason = ""
		out.Numerator = numerator.RatString()
		out.Ratio = new(big.Rat).Quo(numerator, risk).RatString()
		return out
	}
	result.Space = metric(r.LVNSpacePct, true)
	result.Target = metric(r.TargetPrice, false)
	return result
}

// RRFill은 실제 최초 체결과 보호 손절, 진입 시 환율 및 확정 손익 근거를 전달함.
type RRFill struct {
	AverageEntry, InitialStop, Quantity                string
	PriceCurrency, PriceBasis, StopCurrency, StopBasis string
	AccountCurrency, FXRate, FXFrom, FXTo, FXSource    string
	EntryAt, FXAt                                      time.Time
	GrossPnL, Costs, PnLCurrency, CostsCurrency        string
	FillSource, ExitSource, CostSource                 string
	EntryFinal, ExitFinal, CostsFinal                  bool
}

// ParkerRealizedRR은 실제 최초 계좌 위험과 확정 순손익만 사용함.
func ParkerRealizedRR(f RRFill) RRMetric {
	out := RRMetric{State: RRUnknown, Reason: "missing or unconfirmed realized evidence", Provenance: "caller-supplied, unverified", PriceBasis: f.PriceBasis, PolicyVersion: LaneVersion, NumeratorUnit: f.AccountCurrency, DenominatorUnit: f.AccountCurrency}
	if !f.EntryFinal || !f.ExitFinal || !f.CostsFinal || f.AverageEntry == "" || f.InitialStop == "" || f.Quantity == "" || f.PriceCurrency == "" || f.PriceBasis == "" || f.StopCurrency == "" || f.StopBasis == "" || f.AccountCurrency == "" || f.FXRate == "" || f.FXFrom == "" || f.FXTo == "" || f.FXSource == "" || f.FXAt.IsZero() || f.EntryAt.IsZero() || f.GrossPnL == "" || f.Costs == "" || f.PnLCurrency == "" || f.CostsCurrency == "" || f.FillSource == "" || f.ExitSource == "" || f.CostSource == "" {
		return out
	}
	entry, e := positive(f.AverageEntry)
	stop, s := positive(f.InitialStop)
	qty, q := positive(f.Quantity)
	fx, x := positive(f.FXRate)
	pnl, p := decimal(f.GrossPnL)
	cost, c := nonnegative(f.Costs)
	if !e || !s || !q || !x || !p || !c || entry.Cmp(stop) <= 0 || f.PriceBasis != "unadjusted" || f.StopBasis != f.PriceBasis || f.StopCurrency != f.PriceCurrency || f.FXFrom != f.PriceCurrency || f.FXTo != f.AccountCurrency || f.PnLCurrency != f.AccountCurrency || f.CostsCurrency != f.AccountCurrency || !f.FXAt.Equal(f.EntryAt) || (f.PriceCurrency == f.AccountCurrency && fx.Cmp(big.NewRat(1, 1)) != 0) {
		out.State = RRInvalid
		out.Reason = "invalid or incompatible realized evidence"
		return out
	}
	risk := new(big.Rat).Mul(new(big.Rat).Mul(new(big.Rat).Sub(entry, stop), qty), fx)
	net := new(big.Rat).Sub(pnl, cost)
	out.State = RRKnown
	out.Reason = ""
	out.Numerator = net.RatString()
	out.Denominator = risk.RatString()
	out.Ratio = new(big.Rat).Quo(net, risk).RatString()
	out.Provenance = "caller-supplied, unverified; fill=" + f.FillSource + "; exit=" + f.ExitSource + "; cost=" + f.CostSource + "; fx=" + f.FXSource + " at " + f.FXAt.UTC().Format(time.RFC3339Nano) + " (" + f.FXFrom + "/" + f.FXTo + ")"
	return out
}
