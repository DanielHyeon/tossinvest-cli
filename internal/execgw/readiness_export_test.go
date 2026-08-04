package execgw

import (
	"context"
	"sync"
)

// UseReadinessAdapterForTest bypasses the legacy package-test WIRED default so
// a test can exercise the production market-scoped adapter. TESTS ONLY.
func (o *Options) UseReadinessAdapterForTest() { o.forceReadinessAdapterForTest = true }

// SetMarketProtectionForTest installs a market/check-sequence fixture used to
// prove KR/US isolation and last-moment drift refusal. TESTS ONLY; no symbol in
// a built binary can mint this path.
func (o *Options) SetMarketProtectionForTest(fixture func(market string, check int) (bool, string)) {
	var mu sync.Mutex
	counts := map[string]int{}
	o.protectionCheckForTest = func(_ context.Context, market string, previous protectionCheckpoint) (protectionCheckpoint, *RejectedError) {
		mu.Lock()
		defer mu.Unlock()
		counts[market]++
		allowed, identity := fixture(market, counts[market])
		if !allowed {
			return protectionCheckpoint{}, reject(ReasonProtectionNotWired, "test fixture refused protection for %s: %s", market, identity)
		}
		next := protectionCheckpoint{testIdentity: identity}
		if previous.testIdentity != "" && previous.testIdentity != next.testIdentity {
			return protectionCheckpoint{}, reject(ReasonProtectionNotWired, "test fixture protection snapshot drifted for %s", market)
		}
		return next, nil
	}
}
