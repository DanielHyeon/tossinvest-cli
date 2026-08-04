package execgw

import (
	"context"
	"errors"
	"strings"

	"github.com/JungHoonGhae/tossinvest-cli/internal/protection"
)

// StrategyProtectionAuthority is a read-only projection of one sealed a071
// checkpoint. Its private fields prevent scalar status from becoming WIRED.
type StrategyProtectionAuthority struct {
	market     string
	generation uint64
	identity   string
}

func (a StrategyProtectionAuthority) Market() string     { return a.market }
func (a StrategyProtectionAuthority) Generation() uint64 { return a.generation }
func (a StrategyProtectionAuthority) Digest() string {
	if a.generation == 0 || len(a.identity) != 64 {
		return ""
	}
	return "sha256:" + a.identity
}

// ObserveStrategyProtection performs the same sealed provider check used by
// submission without preparing an intent or calling the broker. Submission
// still checks twice and rejects snapshot drift.
func (g *Gateway) ObserveStrategyProtection(ctx context.Context, market string, quantity uint64) (StrategyProtectionAuthority, error) {
	market = strings.ToLower(strings.TrimSpace(market))
	if g == nil || g.protectionReadiness == nil || ctx == nil || (market != "kr" && market != "us") || quantity == 0 {
		return StrategyProtectionAuthority{}, errors.New("execgw: strategy protection authority unavailable")
	}
	checkpoint, refusal := g.protectionReadiness.Check(ctx, protection.ReadinessRequest{
		Market: market, OrderType: "LIMIT", Quantity: quantity,
	}, g.clk.Now(), protection.Checkpoint{})
	if refusal != nil || !checkpoint.Valid() {
		if refusal != nil {
			return StrategyProtectionAuthority{}, refusal
		}
		return StrategyProtectionAuthority{}, errors.New("execgw: invalid strategy protection checkpoint")
	}
	return StrategyProtectionAuthority{market: checkpoint.Market(), generation: checkpoint.Generation(), identity: checkpoint.Identity()}, nil
}
