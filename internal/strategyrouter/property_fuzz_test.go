package strategyrouter

import (
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestSharedQuotaNeverExceedsPhysicalAllowanceProperty(t *testing.T) {
	rng := rand.New(rand.NewSource(7001))
	for iteration := 0; iteration < 2000; iteration++ {
		remaining := uint64(rng.Intn(100) + 1)
		reserve := uint64(rng.Intn(int(remaining) + 1))
		cycleCap := uint64(rng.Intn(100) + 1)
		absoluteCap := uint64(rng.Intn(100) + 1)
		key := PhysicalQuotaKey{Endpoint: "property", ResetGeneration: "reset-" + strconv.Itoa(iteration)}
		authority := newQuotaAuthority(func() time.Time { return routerNow })
		if err := authority.Install(mustQuotaSnapshot(t, key, remaining, reserve, cycleCap, absoluteCap)); err != nil {
			t.Fatal(err)
		}
		for request := 0; request < 200; request++ {
			market := []Market{MarketKR, MarketUS}[request%2]
			horizon := []Horizon{HorizonShort, HorizonWeekly}[(request/2)%2]
			authority.Acquire(validAcquire(key, market, horizon, "request-"+strconv.Itoa(request)))
		}
		allowance := remaining - reserve
		if cycleCap < allowance {
			allowance = cycleCap
		}
		if absoluteCap < allowance {
			allowance = absoluteCap
		}
		status := authority.Status(key)
		if status.Issued > allowance || status.SafetyReserve != reserve {
			t.Fatalf("iteration=%d status=%+v allowance=%d", iteration, status, allowance)
		}
	}
}

func FuzzOwnerKeyNeverIncludesHorizon(f *testing.F) {
	for _, seed := range []struct {
		account string
		symbol  string
		gen     uint64
	}{{"acct", "AAPL", 1}, {"acct-2", " 005930 ", 99}, {"a", "msft", ^uint64(0)}} {
		f.Add(seed.account, seed.symbol, seed.gen)
	}
	f.Fuzz(func(t *testing.T, account, symbol string, generation uint64) {
		short, errShort := NewOwnerKey(account, MarketUS, symbol, generation)
		weekly, errWeekly := NewOwnerKey(account, MarketUS, symbol, generation)
		if (errShort == nil) != (errWeekly == nil) {
			t.Fatal("horizon-independent construction diverged")
		}
		if errShort == nil && short != weekly {
			t.Fatalf("same owner scope diverged: short=%+v weekly=%+v", short, weekly)
		}
	})
}

func FuzzLegacyMigrationRetryConverges(f *testing.F) {
	for _, seed := range []struct{ enabled, verified, combined bool }{{false, false, false}, {true, true, false}, {true, false, false}, {true, true, true}} {
		f.Add(seed.enabled, seed.verified, seed.combined)
	}
	f.Fuzz(func(t *testing.T, enabled, verified, combined bool) {
		legacy := LegacyState{Enabled: enabled, Disabled: !enabled, SelectedMarket: MarketKR, Verified: verified, Combined: combined}
		if enabled && verified && !combined {
			legacy.Record = mustLegacyRecord(t, MarketKR)
		}
		first := MigrateLegacy(NewSchedulerState(), legacy, "migration-fuzz")
		retry := MigrateLegacy(first.State, legacy, "migration-fuzz")
		if !retry.Duplicate || !ValidSchedulerState(retry.State) {
			t.Fatalf("migration retry did not converge: first=%+v retry=%+v", first, retry)
		}
	})
}
