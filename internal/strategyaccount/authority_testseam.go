//go:build tossos_testseams

package strategyaccount

import (
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/risk"
)

func AuthorityForTest(market Market, quoteCurrency string, account risk.AccountState, observedAt, freshUntil time.Time, generation uint64, manifestDigest string) Authority {
	return Authority{account: account, openExposure: account.OpenExposure, market: market, quoteCurrency: quoteCurrency,
		observedAt: observedAt.UTC(), freshUntil: freshUntil.UTC(), generation: generation, manifestDigest: manifestDigest, identity: manifestDigest}
}
