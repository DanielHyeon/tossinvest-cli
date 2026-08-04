package execgw

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
)

// StrategyEntryGateAuthority proves that the exact Gateway-owned entry gate
// refreshed its durable reconciliation projection and admitted this scope.
type StrategyEntryGateAuthority struct {
	generation uint64
	digest     string
}

func (a StrategyEntryGateAuthority) Generation() uint64 { return a.generation }
func (a StrategyEntryGateAuthority) Digest() string     { return a.digest }

func (g *Gateway) ObserveStrategyEntryGate(ctx context.Context, market, symbol string) (StrategyEntryGateAuthority, error) {
	market, symbol = strings.ToLower(strings.TrimSpace(market)), strings.ToUpper(strings.TrimSpace(symbol))
	if g == nil || g.entry == nil || ctx == nil || ctx.Err() != nil || (market != "kr" && market != "us") || symbol == "" {
		return StrategyEntryGateAuthority{}, errors.New("execgw: strategy entry-gate authority unavailable")
	}
	if refusal := g.entry.CheckEntryFor(market, symbol); refusal != nil {
		return StrategyEntryGateAuthority{}, refusal
	}
	g.entry.mu.Lock()
	generation := g.entry.revision
	g.entry.mu.Unlock()
	return sealStrategyEntryGateAuthority(g.accountRef, market, symbol, generation), nil
}

func sealStrategyEntryGateAuthority(accountRef, market, symbol string, generation uint64) StrategyEntryGateAuthority {
	digest := sha256.Sum256([]byte(strings.Join([]string{"TossOS/strategy-entry-gate/v2", accountRef, market, symbol,
		fmt.Sprintf("%d", generation)}, "\x00")))
	return StrategyEntryGateAuthority{generation: generation, digest: "sha256:" + hex.EncodeToString(digest[:])}
}

// withStrategyEntryGateAuthority performs the final admission and broker-call
// interlock. Projection refresh happens before the mutex; the exact revision
// and all in-memory blocks are then held stable until fn returns.
func (g *Gateway) withStrategyEntryGateAuthority(ctx context.Context, expected StrategyEntryGateAuthority,
	market, symbol string, fn func() error,
) error {
	if g == nil || g.entry == nil || fn == nil || expected.generation == 0 || expected.digest == "" {
		return errors.New("execgw: strategy entry-gate authority unavailable")
	}
	market, symbol = strings.ToLower(strings.TrimSpace(market)), strings.ToUpper(strings.TrimSpace(symbol))
	g.entry.mu.Lock()
	refresh := g.entry.authorityRefresh
	g.entry.mu.Unlock()
	if refresh != nil {
		if err := refresh(); err != nil {
			return fmt.Errorf("execgw: durable reconciliation refresh failed: %w", err)
		}
	}
	g.entry.mu.Lock()
	defer g.entry.mu.Unlock()
	current := sealStrategyEntryGateAuthority(g.accountRef, market, symbol, g.entry.revision)
	if current != expected {
		return errors.New("execgw: strategy entry-gate authority changed")
	}
	if rejected := g.entry.checkAccountEntryLocked(); rejected != nil {
		return rejected
	}
	if rejected := g.entry.checkSymbolEntryLocked(market, symbol); rejected != nil {
		return rejected
	}
	return fn()
}
