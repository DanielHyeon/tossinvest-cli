//go:build tossos_testseams

package officialfx

import (
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
)

// EvidenceForAuthorityTest mints opaque evidence only in explicitly tagged
// test binaries. Production builds contain no scalar-to-authority constructor.
func EvidenceForAuthorityTest(quoteCurrency, accountCurrency, rate, haircut string, observedAt, freshUntil time.Time) (Evidence, error) {
	if quoteCurrency == accountCurrency {
		snapshot, err := newIdentitySnapshot(quoteCurrency, "authority-test-identity", sha256Identity("authority-test-identity"), observedAt, freshUntil)
		if err != nil {
			return Evidence{}, err
		}
		return Identity(snapshot)
	}
	policy, err := newHaircutPolicy("authority-test-haircut", "authority-test-v1", haircut, observedAt, freshUntil)
	if err != nil {
		return Evidence{}, err
	}
	return sealOfficial(domain.ExchangeRate{BaseCurrency: quoteCurrency, QuoteCurrency: accountCurrency,
		Code: quoteCurrency + "/" + accountCurrency, RateRaw: rate, MidRateRaw: rate,
		ValidFromRaw: observedAt.UTC().Format(time.RFC3339Nano), ValidUntilRaw: freshUntil.UTC().Format(time.RFC3339Nano)},
		quoteCurrency, accountCurrency, policy)
}
