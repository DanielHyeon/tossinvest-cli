package httpapi

import "github.com/JungHoonGhae/tossinvest-cli/internal/operatorview"

func ExitLineReferenceFrom(v operatorview.ExitLineReferenceView) *ExitLineReference {
	if !v.Present() {
		return nil
	}
	return &ExitLineReference{
		Kind: string(v.Kind), Label: v.Label, EffectiveKnown: v.EffectiveKnown,
		Market: v.Market, Currency: v.Currency, EntryPrice: v.EntryPrice,
		InitialStop: v.InitialStop, Baseline: v.Baseline, HighWater: v.HighWater,
		StopPercent: v.StopPercent, Basis: v.Basis, Reason: v.Reason,
	}
}
