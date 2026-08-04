//go:build tossos_testseams

package risk

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/costs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/officialfx"
)

// TestOnlySealAccountBaseFX is available only in explicitly tagged test
// binaries. Normal production builds expose no raw-value mint; they must call
// BindAccountBaseFX with opaque officialfx evidence.
func TestOnlySealAccountBaseFX(at time.Time, market costs.Market, policy Policy, rate, haircut string) (AccountBaseFX, error) {
	if at.IsZero() {
		return AccountBaseFX{}, fmt.Errorf("risk test seam: evaluation instant is zero")
	}
	if err := policy.Validate(); err != nil {
		return AccountBaseFX{}, fmt.Errorf("risk test seam: policy: %w", err)
	}
	quote, err := currencyOf(market)
	if err != nil {
		return AccountBaseFX{}, err
	}
	base := policy.LimitCurrency()
	source, version := officialfx.OfficialSource, officialfx.OfficialVersion
	if quote == base {
		source, version = officialfx.IdentitySource, officialfx.IdentityVersion
	}
	digestInput := strings.Join([]string{"tossos-testseam", quote, base, rate, haircut, at.UTC().Format(time.RFC3339Nano)}, "\x00")
	digestSum := sha256.Sum256([]byte(digestInput))
	fx := AccountBaseFX{
		quoteCurrency: quote, accountCurrency: base, rate: rate, haircut: haircut,
		source: source, version: version, digest: "sha256:" + hex.EncodeToString(digestSum[:]),
		observedAt: at.UTC().Add(-time.Minute), freshUntil: at.UTC().Add(time.Minute), evaluatedAt: at.UTC(),
	}
	fx.seal = sealAccountBaseFX(fx)
	if _, err := fx.multiplierAt(at.UTC(), quote, base); err != nil {
		return AccountBaseFX{}, err
	}
	return fx, nil
}
