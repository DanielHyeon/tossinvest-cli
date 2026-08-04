package candidate

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"time"
)

const (
	productionThresholdJSONMaximumBytes     = 1 << 20
	productionThresholdEvidenceMaximumBytes = 16 << 20
)

var ErrProductionThresholdAuthorityUnavailable = errors.New("candidate: production threshold authority unavailable")

// ProductionThresholdAuthorityConfig contains only a closed market, an
// absolute directory and an externally supplied digest pin. It contains no
// writer, approval constructor, signing key, toggle or execution capability.
type ProductionThresholdAuthorityConfig struct {
	ConfigDir        string
	Market           string
	ActivationDigest string
}

// ProductionThresholdAuthority is an immutable-by-copy read result. Its
// ThresholdSet remains sealed by the candidate package's private validity bit.
type ProductionThresholdAuthority struct {
	valid            bool
	market           string
	activationDigest string
	set              ThresholdSet
}

func (a ProductionThresholdAuthority) Valid() bool                { return a.valid }
func (a ProductionThresholdAuthority) Market() string             { return a.market }
func (a ProductionThresholdAuthority) ActivationDigest() string   { return a.activationDigest }
func (a ProductionThresholdAuthority) SetDigest() string          { return a.set.SetDigest() }
func (a ProductionThresholdAuthority) EvidenceDigest() string     { return a.set.EvidenceDigest() }
func (a ProductionThresholdAuthority) Version() string            { return a.set.Version() }
func (a ProductionThresholdAuthority) ApprovedAt() time.Time      { return a.set.ApprovedAt() }
func (a ProductionThresholdAuthority) ThresholdSet() ThresholdSet { return a.set }

func ProductionThresholdSetFileName(market string) string {
	return productionThresholdFileName(market, "candidate-thresholds-%s.json")
}

func ProductionThresholdEvidenceFileName(market string) string {
	return productionThresholdFileName(market, "candidate-threshold-evidence-%s.bin")
}

func ProductionThresholdActivationFileName(market string) string {
	return productionThresholdFileName(market, "candidate-threshold-activation-%s.json")
}

func productionThresholdFileName(market, pattern string) string {
	switch market {
	case MarketKR, MarketUS:
		return fmt.Sprintf(pattern, market)
	default:
		return ""
	}
}

// LoadProductionThresholdAuthority consumes an approval made by a separate
// human-controlled process. It does not create, update or activate approval.
func LoadProductionThresholdAuthority(ctx context.Context, config ProductionThresholdAuthorityConfig,
	asOf time.Time, futureSkew time.Duration,
) (ProductionThresholdAuthority, error) {
	if ctx == nil || asOf.IsZero() || futureSkew < 0 {
		return ProductionThresholdAuthority{}, ErrProductionThresholdAuthorityUnavailable
	}
	if err := ctx.Err(); err != nil {
		return ProductionThresholdAuthority{}, err
	}
	config.ConfigDir = filepath.Clean(strings.TrimSpace(config.ConfigDir))
	config.Market = strings.ToUpper(strings.TrimSpace(config.Market))
	config.ActivationDigest = strings.TrimSpace(config.ActivationDigest)
	setName := ProductionThresholdSetFileName(config.Market)
	evidenceName := ProductionThresholdEvidenceFileName(config.Market)
	activationName := ProductionThresholdActivationFileName(config.Market)
	owner, ownerOK := productionThresholdOwnerUID()
	if !ownerOK || !filepath.IsAbs(config.ConfigDir) || setName == "" || evidenceName == "" || activationName == "" ||
		!canonicalProductionThresholdDigest(config.ActivationDigest) {
		return ProductionThresholdAuthority{}, ErrProductionThresholdAuthorityUnavailable
	}

	activationBytes, err := readProductionThresholdFile(filepath.Join(config.ConfigDir, activationName), owner, 0o400, productionThresholdJSONMaximumBytes)
	if err != nil || productionThresholdDigest(activationBytes) != config.ActivationDigest {
		return ProductionThresholdAuthority{}, ErrProductionThresholdAuthorityUnavailable
	}
	setBytes, err := readProductionThresholdFile(filepath.Join(config.ConfigDir, setName), owner, 0o400, productionThresholdJSONMaximumBytes)
	if err != nil {
		return ProductionThresholdAuthority{}, ErrProductionThresholdAuthorityUnavailable
	}
	evidenceBytes, err := readProductionThresholdFile(filepath.Join(config.ConfigDir, evidenceName), owner, 0o400, productionThresholdEvidenceMaximumBytes)
	if err != nil {
		return ProductionThresholdAuthority{}, ErrProductionThresholdAuthorityUnavailable
	}
	if err := ctx.Err(); err != nil {
		return ProductionThresholdAuthority{}, err
	}
	activation, err := LoadActivationRecord(strings.NewReader(string(activationBytes)))
	if err != nil {
		return ProductionThresholdAuthority{}, fmt.Errorf("%w: %v", ErrProductionThresholdAuthorityUnavailable, err)
	}
	set, err := LoadThresholdSet(strings.NewReader(string(setBytes)), evidenceBytes, activation,
		ThresholdScope{Market: config.Market, Session: SessionRegular}, asOf.UTC(), futureSkew)
	if err != nil {
		return ProductionThresholdAuthority{}, fmt.Errorf("%w: %v", ErrProductionThresholdAuthorityUnavailable, err)
	}
	if err := ctx.Err(); err != nil {
		return ProductionThresholdAuthority{}, err
	}
	return ProductionThresholdAuthority{valid: true, market: config.Market,
		activationDigest: config.ActivationDigest, set: set}, nil
}

func canonicalProductionThresholdDigest(value string) bool {
	if len(value) != len("sha256:")+sha256.Size*2 || !strings.HasPrefix(value, "sha256:") || value != strings.ToLower(value) {
		return false
	}
	decoded, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil && len(decoded) == sha256.Size
}

func productionThresholdDigest(data []byte) string {
	digest := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(digest[:])
}
