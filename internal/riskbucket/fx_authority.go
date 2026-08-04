package riskbucket

import (
	"fmt"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/officialfx"
)

// BindFXAuthority converts one opaque officialfx capability into the public
// arithmetic DTO only inside riskbucket. Caller-created FXEvidence is ignored;
// the returned policy is a value copy bound to the exact evaluation instant.
func BindFXAuthority(policy ReservePolicy, authority officialfx.Evidence, at time.Time) (ReservePolicy, error) {
	reserve, err := authority.EvidenceAt(at, policy.QuoteCurrency, policy.AccountCurrency)
	if err != nil {
		return ReservePolicy{}, refusal(RefusalCurrencyUnresolved, "fx_authority", err)
	}
	policy.EvaluatedAt = at
	policy.FX = FXEvidence{
		RateQuoteToBase: reserve.RateQuoteToBase(),
		Haircut:         reserve.Haircut(),
		Evidence: Evidence{
			Source: reserve.Source(), Version: reserve.Version(), Digest: reserve.Digest(),
			Official: true, Frozen: true, ObservedAt: reserve.ObservedAt(), FreshUntil: reserve.FreshUntil(),
		},
	}
	if _, _, _, _, _, _, err := validateReservePolicy(policy); err != nil {
		return ReservePolicy{}, fmt.Errorf("risk bucket: bind opaque FX authority: %w", err)
	}
	return policy, nil
}
