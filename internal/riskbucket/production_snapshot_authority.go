package riskbucket

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/url"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/officialfx"
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyflow"
	_ "modernc.org/sqlite"
)

const (
	productionRiskPolicySchema       = "strategy-risk-bucket-policy:v1"
	productionRiskPolicyDomain       = "TossOS/strategy-risk-bucket-policy/ed25519/v1"
	productionRiskPolicyAlgorithm    = "Ed25519"
	productionRiskPolicyMaximumBytes = 1 << 20
	productionRiskSnapshotWindow     = 5 * time.Second
	productionRiskJournalSchema      = 27
)

var ErrProductionRiskSnapshotUnavailable = errors.New("risk bucket: production snapshot authority unavailable")

// ProductionRiskSnapshotConfig contains only read paths and externally managed
// trust pins. It exposes no policy writer, signer, toggle or execution handle.
type ProductionRiskSnapshotConfig struct {
	ConfigDir, JournalPath       string
	Market                       Market
	AccountID, AccountCurrency   string
	ManifestDigest, TrustedKeyID string
	TrustedKey                   ed25519.PublicKey
	ObservedAt                   time.Time
}

// ProductionRiskSnapshotInput carries opaque authorities. Result and FX remain
// independently sealed by their owning packages and are revalidated here.
type ProductionRiskSnapshotInput struct {
	Result strategyflow.Result
	FX     officialfx.Evidence
}

type productionRiskFeePolicy struct {
	FixedBaseMinor   string `json:"fixed_base_minor"`
	PerUnitBaseMinor string `json:"per_unit_base_minor"`
	MinimumBaseMinor string `json:"minimum_base_minor"`
	Version          string `json:"version"`
	Digest           string `json:"digest"`
}

type productionRiskStrategyPolicy struct {
	LaneID      string  `json:"lane_id"`
	LaneVersion string  `json:"lane_version"`
	Horizon     Horizon `json:"horizon"`
	RiskID      string  `json:"risk_id"`
	RiskVersion string  `json:"risk_version"`
	LimitMinor  string  `json:"limit_minor"`
}

type productionRiskSymbolPolicy struct {
	Symbol           string `json:"symbol"`
	Sector           string `json:"sector"`
	SectorLimitMinor string `json:"sector_limit_minor"`
	SymbolLimitMinor string `json:"symbol_limit_minor"`
}

type productionRiskPolicyBody struct {
	SchemaVersion      string                         `json:"schema_version"`
	Domain             string                         `json:"domain"`
	SignatureAlgorithm string                         `json:"signature_algorithm"`
	KeyID              string                         `json:"key_id"`
	Generation         uint64                         `json:"generation"`
	Market             Market                         `json:"market"`
	AccountID          string                         `json:"account_id"`
	AccountCurrency    string                         `json:"account_currency"`
	QuoteCurrency      string                         `json:"quote_currency"`
	PolicyVersion      string                         `json:"policy_version"`
	Approver           string                         `json:"approver"`
	ObservedAt         string                         `json:"observed_at"`
	FreshUntil         string                         `json:"fresh_until"`
	Revoked            bool                           `json:"revoked"`
	Fee                productionRiskFeePolicy        `json:"fee"`
	HorizonLimits      map[Horizon]string             `json:"horizon_limits"`
	MarketLimitMinor   string                         `json:"market_limit_minor"`
	Strategies         []productionRiskStrategyPolicy `json:"strategies"`
	Symbols            []productionRiskSymbolPolicy   `json:"symbols"`
}

type productionRiskPolicyManifest struct {
	productionRiskPolicyBody
	Signature string `json:"signature"`
}

type productionRiskUsageRow struct {
	ReservationID, PolicyVersion, HeldMinor, FilledMinor, State string
	SnapshotID, PolicyRecordDigest                              string
	OverageLatched, UnknownLatched                              int
}

type fixedProductionRiskSnapshotSource struct{ material riskSnapshotAuthorityMaterial }

func (source fixedProductionRiskSnapshotSource) loadRiskSnapshotAuthority(context.Context, RiskSnapshotScope) (riskSnapshotAuthorityMaterial, error) {
	return source.material, nil
}

func ProductionRiskPolicyFileName(market Market) string {
	switch market {
	case MarketKR:
		return "risk-bucket-policy-KR.json"
	case MarketUS:
		return "risk-bucket-policy-US.json"
	default:
		return ""
	}
}

// LoadProductionRiskSnapshotAuthority consumes one signed market policy and
// current read-only journal state. It never creates, migrates or writes either.
func LoadProductionRiskSnapshotAuthority(ctx context.Context, config ProductionRiskSnapshotConfig, input ProductionRiskSnapshotInput) (RiskSnapshotAuthorityBundle, error) {
	if ctx == nil || config.ObservedAt.IsZero() {
		return RiskSnapshotAuthorityBundle{}, ErrProductionRiskSnapshotUnavailable
	}
	if err := ctx.Err(); err != nil {
		return RiskSnapshotAuthorityBundle{}, err
	}
	config = canonicalProductionRiskConfig(config)
	owner, ownerOK := productionRiskOwnerUID()
	name := ProductionRiskPolicyFileName(config.Market)
	if !ownerOK || name == "" || !filepath.IsAbs(config.ConfigDir) || !filepath.IsAbs(config.JournalPath) ||
		!canonicalRiskDigest(config.ManifestDigest) || !canonicalIdentity(config.TrustedKeyID) || len(config.TrustedKey) != ed25519.PublicKeySize ||
		!canonicalIdentity(config.AccountID) || !canonicalCurrency(config.AccountCurrency) {
		return RiskSnapshotAuthorityBundle{}, ErrProductionRiskSnapshotUnavailable
	}
	data, err := readProductionRiskFile(filepath.Join(config.ConfigDir, name), owner, 0o400, productionRiskPolicyMaximumBytes)
	if err != nil || productionRiskDigest(data) != config.ManifestDigest {
		return RiskSnapshotAuthorityBundle{}, ErrProductionRiskSnapshotUnavailable
	}
	manifest, err := decodeProductionRiskPolicy(data)
	if err != nil || !verifyProductionRiskPolicy(manifest, config) {
		return RiskSnapshotAuthorityBundle{}, ErrProductionRiskSnapshotUnavailable
	}
	scope, reserve, limits, err := bindProductionRiskInputs(config, manifest.productionRiskPolicyBody, input)
	if err != nil {
		return RiskSnapshotAuthorityBundle{}, fmt.Errorf("%w: %v", ErrProductionRiskSnapshotUnavailable, err)
	}
	entries, err := loadProductionRiskEntries(ctx, config, manifest.productionRiskPolicyBody, scope, reserve, limits)
	if err != nil {
		return RiskSnapshotAuthorityBundle{}, fmt.Errorf("%w: %v", ErrProductionRiskSnapshotUnavailable, err)
	}
	material := riskSnapshotAuthorityMaterial{Scope: scope, Policy: reserve, Generation: manifest.Generation, Entries: entries}
	service := newRiskSnapshotAuthorityService(fixedProductionRiskSnapshotSource{material: material})
	return service.Load(ctx, scope)
}

func canonicalProductionRiskConfig(config ProductionRiskSnapshotConfig) ProductionRiskSnapshotConfig {
	config.ConfigDir = filepath.Clean(strings.TrimSpace(config.ConfigDir))
	config.JournalPath = filepath.Clean(strings.TrimSpace(config.JournalPath))
	config.AccountID = strings.TrimSpace(config.AccountID)
	config.AccountCurrency = strings.ToUpper(strings.TrimSpace(config.AccountCurrency))
	config.ManifestDigest = strings.TrimSpace(config.ManifestDigest)
	config.TrustedKeyID = strings.TrimSpace(config.TrustedKeyID)
	config.TrustedKey = append(ed25519.PublicKey(nil), config.TrustedKey...)
	config.ObservedAt = config.ObservedAt.UTC()
	return config
}

func decodeProductionRiskPolicy(data []byte) (productionRiskPolicyManifest, error) {
	if len(data) == 0 || len(data) > productionRiskPolicyMaximumBytes {
		return productionRiskPolicyManifest{}, ErrProductionRiskSnapshotUnavailable
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var manifest productionRiskPolicyManifest
	if err := decoder.Decode(&manifest); err != nil {
		return productionRiskPolicyManifest{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return productionRiskPolicyManifest{}, errors.New("risk bucket: trailing policy JSON")
	}
	canonical, err := json.Marshal(manifest)
	if err != nil || !bytes.Equal(canonical, data) {
		return productionRiskPolicyManifest{}, errors.New("risk bucket: non-canonical policy JSON")
	}
	return manifest, nil
}

func verifyProductionRiskPolicy(manifest productionRiskPolicyManifest, config ProductionRiskSnapshotConfig) bool {
	body := manifest.productionRiskPolicyBody
	if body.SchemaVersion != productionRiskPolicySchema || body.Domain != productionRiskPolicyDomain ||
		body.SignatureAlgorithm != productionRiskPolicyAlgorithm || body.KeyID != config.TrustedKeyID || body.Generation == 0 ||
		body.Market != config.Market || body.AccountID != config.AccountID || body.AccountCurrency != config.AccountCurrency ||
		body.QuoteCurrency != productionRiskQuoteCurrency(config.Market) || !canonicalIdentity(body.PolicyVersion) ||
		!canonicalIdentity(body.Approver) || body.Revoked || !validProductionRiskPolicyContents(body) {
		return false
	}
	observed, observedOK := canonicalProductionRiskTime(body.ObservedAt)
	freshUntil, freshOK := canonicalProductionRiskTime(body.FreshUntil)
	if !observedOK || !freshOK || observed.After(config.ObservedAt) || freshUntil.Before(config.ObservedAt) || observed.After(freshUntil) {
		return false
	}
	canonicalBody, err := json.Marshal(body)
	if err != nil {
		return false
	}
	signature, err := base64.StdEncoding.Strict().DecodeString(manifest.Signature)
	return err == nil && base64.StdEncoding.EncodeToString(signature) == manifest.Signature && len(signature) == ed25519.SignatureSize &&
		ed25519.Verify(config.TrustedKey, canonicalBody, signature)
}

func validProductionRiskPolicyContents(body productionRiskPolicyBody) bool {
	fee := body.Fee
	if !canonicalIdentity(fee.Version) || !canonicalRiskDigest(fee.Digest) {
		return false
	}
	for _, raw := range []string{fee.FixedBaseMinor, fee.PerUnitBaseMinor, fee.MinimumBaseMinor} {
		if _, err := parseDecimal(raw, true, 256); err != nil {
			return false
		}
	}
	for _, raw := range []string{body.MarketLimitMinor, body.HorizonLimits[HorizonShort], body.HorizonLimits[HorizonMedium]} {
		if _, err := parseMinor(raw, 256); err != nil {
			return false
		}
	}
	if len(body.HorizonLimits) != 2 || len(body.Strategies) == 0 || len(body.Symbols) == 0 {
		return false
	}
	strategyKeys := map[string]bool{}
	for _, value := range body.Strategies {
		key := value.LaneID + "\x00" + value.LaneVersion + "\x00" + string(value.Horizon)
		if strategyKeys[key] || !canonicalIdentity(value.LaneID) || !canonicalIdentity(value.LaneVersion) ||
			(value.Horizon != HorizonShort && value.Horizon != HorizonMedium) || !canonicalIdentity(value.RiskID) ||
			!canonicalIdentity(value.RiskVersion) {
			return false
		}
		if _, err := parseMinor(value.LimitMinor, 256); err != nil {
			return false
		}
		strategyKeys[key] = true
	}
	symbols := map[string]bool{}
	for _, value := range body.Symbols {
		if symbols[value.Symbol] || value.Symbol == "" || value.Symbol != strings.ToUpper(strings.TrimSpace(value.Symbol)) || !canonicalIdentity(value.Sector) {
			return false
		}
		if _, err := parseMinor(value.SectorLimitMinor, 256); err != nil {
			return false
		}
		if _, err := parseMinor(value.SymbolLimitMinor, 256); err != nil {
			return false
		}
		symbols[value.Symbol] = true
	}
	return true
}

func bindProductionRiskInputs(config ProductionRiskSnapshotConfig, body productionRiskPolicyBody, input ProductionRiskSnapshotInput) (RiskSnapshotScope, ReservePolicy, map[Dimension]string, error) {
	result, lineage, terms := input.Result, input.Result.Lineage, input.Result.ExecutionTerms
	if !result.ValidProposal() || result.Code != strategyflow.RefusalNone || result.Quantity == 0 || !lineage.Complete || !lineage.Valid() || !terms.Valid() ||
		lineage.AccountRef != config.AccountID || Market(lineage.Market) != config.Market || terms.AccountRef() != config.AccountID ||
		terms.Market() != lineage.Market || terms.Symbol() != lineage.Symbol || terms.Quantity() != result.Quantity ||
		terms.LineageIdentity() != lineage.Identity || lineage.Symbol == "" || lineage.Symbol != strings.ToUpper(strings.TrimSpace(lineage.Symbol)) {
		return RiskSnapshotScope{}, ReservePolicy{}, nil, errors.New("sealed strategy result mismatch")
	}
	horizon := Horizon(lineage.Horizon)
	if horizon != HorizonShort && horizon != HorizonMedium {
		return RiskSnapshotScope{}, ReservePolicy{}, nil, errors.New("unsupported horizon")
	}
	strategy, ok := exactProductionRiskStrategy(body.Strategies, lineage.LaneID, lineage.LaneVersion, horizon)
	if !ok {
		return RiskSnapshotScope{}, ReservePolicy{}, nil, errors.New("strategy risk mapping unavailable")
	}
	symbol, ok := exactProductionRiskSymbol(body.Symbols, lineage.Symbol)
	if !ok {
		return RiskSnapshotScope{}, ReservePolicy{}, nil, errors.New("symbol sector mapping unavailable")
	}
	entry := terms.Entry()
	entryObserved, entryOK := canonicalProductionRiskTime(entry.AsOf())
	entryMajor, entryMajorOK := entry.MajorDecimal()
	entryFresh := time.Unix(0, lineage.CandidateValidUntilNS).UTC()
	if !entryOK || !entryMajorOK || entry.Currency() != body.QuoteCurrency || entry.UnitVersion() != "minor-v1" ||
		entry.MinorScale() != productionRiskMinorScale(config.Market) || entry.Source() == "" || entry.Version() == "" || entry.Digest() == "" ||
		entryObserved.After(config.ObservedAt) || entryFresh.Before(config.ObservedAt) {
		return RiskSnapshotScope{}, ReservePolicy{}, nil, errors.New("worst executable price authority unavailable")
	}
	fx, err := input.FX.EvidenceAt(config.ObservedAt, body.QuoteCurrency, body.AccountCurrency)
	if err != nil {
		return RiskSnapshotScope{}, ReservePolicy{}, nil, err
	}
	policyObserved, _ := canonicalProductionRiskTime(body.ObservedAt)
	policyFresh, _ := canonicalProductionRiskTime(body.FreshUntil)
	observed := latestProductionRiskTime(policyObserved, entryObserved, fx.ObservedAt())
	fresh := earliestProductionRiskTime(policyFresh, entryFresh, fx.FreshUntil())
	if observed.After(config.ObservedAt) || fresh.Before(config.ObservedAt) {
		return RiskSnapshotScope{}, ReservePolicy{}, nil, errors.New("policy windows do not intersect")
	}
	versionDigest := productionRiskDigest([]byte(strings.Join([]string{config.ManifestDigest, lineage.Identity, terms.Identity(), input.FX.Digest(),
		config.ObservedAt.Format(time.RFC3339Nano)}, "\x00")))
	version := strategy.RiskVersion + ":" + strings.TrimPrefix(versionDigest, "sha256:")[:24]
	scope := RiskSnapshotScope{AccountID: config.AccountID, Market: config.Market, Horizon: horizon,
		StrategyRiskID: strategy.RiskID, StrategyRiskVersion: version, Sector: symbol.Sector, Symbol: lineage.Symbol,
		AccountCurrency: body.AccountCurrency, QuoteCurrency: body.QuoteCurrency, AsOf: config.ObservedAt}
	fee := body.Fee
	reserve := ReservePolicy{AccountCurrency: body.AccountCurrency, QuoteCurrency: body.QuoteCurrency, EvaluatedAt: config.ObservedAt, MaxDecimalBits: 256,
		Price: PriceEvidence{WorstExecutableQuote: entryMajor, Evidence: Evidence{Source: "tossos-official-limit-contract", Version: entry.Version(),
			Digest:   productionRiskDigest([]byte(strings.Join([]string{entry.Source(), entry.Version(), entry.Digest(), entry.PriceMinor(), terms.Identity()}, "\x00"))),
			Official: true, Frozen: true, ObservedAt: entryObserved, FreshUntil: entryFresh}},
		FX: FXEvidence{RateQuoteToBase: fx.RateQuoteToBase(), Haircut: fx.Haircut(), Evidence: Evidence{Source: fx.Source(), Version: fx.Version(),
			Digest: fx.Digest(), Official: true, Frozen: true, ObservedAt: fx.ObservedAt(), FreshUntil: fx.FreshUntil()}},
		Fee: FeePolicy{FixedBaseMinor: fee.FixedBaseMinor, PerUnitBaseMinor: fee.PerUnitBaseMinor, MinimumBaseMinor: fee.MinimumBaseMinor,
			Version: fee.Version, Digest: fee.Digest}}
	if _, _, _, _, _, _, err := validateReservePolicy(reserve); err != nil {
		return RiskSnapshotScope{}, ReservePolicy{}, nil, err
	}
	limits := map[Dimension]string{DimensionHorizon: body.HorizonLimits[horizon], DimensionMarket: body.MarketLimitMinor,
		DimensionStrategy: strategy.LimitMinor, DimensionSector: symbol.SectorLimitMinor, DimensionSymbol: symbol.SymbolLimitMinor}
	for _, dimension := range requiredDimensions {
		if _, err := parseMinor(limits[dimension], reserve.MaxDecimalBits); err != nil {
			return RiskSnapshotScope{}, ReservePolicy{}, nil, fmt.Errorf("invalid %s limit", dimension)
		}
	}
	return scope, reserve, limits, nil
}

func loadProductionRiskEntries(ctx context.Context, config ProductionRiskSnapshotConfig, body productionRiskPolicyBody, scope RiskSnapshotScope, reserve ReservePolicy, limits map[Dimension]string) ([]riskSnapshotAuthorityMaterialEntry, error) {
	owner, ok := productionRiskOwnerUID()
	if !ok {
		return nil, ErrProductionRiskSnapshotUnavailable
	}
	if err := validateProductionRiskJournalFile(config.JournalPath, owner); err != nil {
		return nil, err
	}
	query := url.Values{}
	query.Set("mode", "ro")
	query.Add("_pragma", "query_only(true)")
	query.Add("_pragma", "busy_timeout(5000)")
	dsn := url.URL{Scheme: "file", Path: config.JournalPath, RawQuery: query.Encode()}
	db, err := sql.Open("sqlite", dsn.String())
	if err != nil {
		return nil, err
	}
	defer db.Close()
	db.SetMaxOpenConns(1)
	if err := db.PingContext(ctx); err != nil {
		return nil, err
	}
	var version int
	if err := db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil || version != productionRiskJournalSchema {
		return nil, errors.New("risk bucket: exact journal schema unavailable")
	}
	var scopeLatches int
	if err := db.QueryRowContext(ctx, `SELECT count(*) FROM risk_bucket_scope_latches WHERE account_ref=? AND market=? AND symbol=?`,
		scope.AccountID, string(scope.Market), scope.Symbol).Scan(&scopeLatches); err != nil || scopeLatches != 0 {
		return nil, errors.New("risk bucket: scope latch present")
	}
	values := map[Dimension]string{DimensionHorizon: string(scope.Horizon), DimensionMarket: string(scope.Market),
		DimensionStrategy: scope.StrategyRiskID, DimensionSector: scope.Sector, DimensionSymbol: scope.Symbol}
	manifestObserved, _ := canonicalProductionRiskTime(body.ObservedAt)
	manifestFresh, _ := canonicalProductionRiskTime(body.FreshUntil)
	authorityObserved := latestProductionRiskTime(manifestObserved, reserve.Price.ObservedAt, reserve.FX.ObservedAt)
	authorityFresh := earliestProductionRiskTime(manifestFresh, reserve.Price.FreshUntil, reserve.FX.FreshUntil,
		scope.AsOf.Add(productionRiskSnapshotWindow))
	if authorityObserved.After(scope.AsOf) || authorityFresh.Before(scope.AsOf) {
		return nil, errors.New("risk bucket: authority window unavailable")
	}
	entries := make([]riskSnapshotAuthorityMaterialEntry, 0, len(requiredDimensions))
	for _, dimension := range requiredDimensions {
		rows, err := readProductionRiskUsage(ctx, db, scope.AccountID, dimension, values[dimension])
		if err != nil {
			return nil, err
		}
		filled, held, rowDigest, err := aggregateProductionRiskUsage(rows)
		if err != nil {
			return nil, err
		}
		key := BucketKey{Dimension: dimension, Value: values[dimension], PolicyVersion: scope.StrategyRiskVersion}
		policyDigest := productionRiskDigest([]byte(strings.Join([]string{config.ManifestDigest, body.PolicyVersion, string(dimension), values[dimension],
			scope.StrategyRiskVersion, reserve.Price.Digest, reserve.FX.Digest, reserve.Fee.Digest}, "\x00")))
		policyEvidence := Evidence{Source: RiskPolicyAuthoritySource, Version: key.PolicyVersion, Digest: policyDigest, Official: true, Frozen: true,
			ObservedAt: authorityObserved, FreshUntil: authorityFresh}
		policyProvenance, err := NewPolicyProvenance(key, policyEvidence)
		if err != nil {
			return nil, err
		}
		snapshotDigest := productionRiskDigest([]byte(strings.Join([]string{config.ManifestDigest, string(dimension), values[dimension], limits[dimension], filled, held,
			rowDigest, scope.AsOf.Format(time.RFC3339Nano)}, "\x00")))
		snapshotVersion := "risk-snapshot:v1:" + strings.TrimPrefix(snapshotDigest, "sha256:")
		binding := BucketSnapshotBinding{Key: key, LimitMinor: limits[dimension], FilledMinor: filled, HeldMinor: held, SnapshotVersion: snapshotVersion}
		snapshotEvidence := Evidence{Source: RiskSnapshotAuthoritySource, Version: snapshotVersion, Digest: snapshotDigest, Official: true, Frozen: true,
			ObservedAt: scope.AsOf, FreshUntil: authorityFresh}
		snapshotProvenance, err := NewSnapshotProvenance(binding, snapshotEvidence)
		if err != nil {
			return nil, err
		}
		bucket := BucketSnapshot{Key: key, LimitMinor: binding.LimitMinor, FilledMinor: filled, HeldMinor: held, SnapshotVersion: snapshotVersion,
			PolicyProvenance: policyProvenance, SnapshotProvenance: snapshotProvenance}
		reference := RiskSnapshotJournalReference{Key: key, SnapshotID: "journal-" + snapshotVersion, SnapshotDigest: snapshotDigest,
			SnapshotVersion: snapshotVersion, PolicyDigest: policyDigest, PolicyObservedAt: policyEvidence.ObservedAt,
			PolicyFreshUntil: policyEvidence.FreshUntil, SnapshotObservedAt: snapshotEvidence.ObservedAt, SnapshotFreshUntil: snapshotEvidence.FreshUntil}
		entry, err := newRiskSnapshotAuthorityMaterialEntry(bucket, reference)
		if err != nil {
			return nil, err
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func readProductionRiskUsage(ctx context.Context, db *sql.DB, account string, dimension Dimension, value string) ([]productionRiskUsageRow, error) {
	rows, err := db.QueryContext(ctx, `SELECT r.reservation_id,r.policy_version,r.held_minor,r.filled_minor,r.state,
		r.risk_overage_latched,r.unknown_actual_latched,COALESCE(s.snapshot_id,''),COALESCE(p.record_digest,'')
		FROM risk_bucket_reservations r
		LEFT JOIN risk_bucket_snapshots s ON s.snapshot_id=r.snapshot_id AND s.bucket_dimension=r.bucket_dimension AND s.bucket_value=r.bucket_value AND s.policy_version=r.policy_version
		LEFT JOIN risk_bucket_policies p ON p.bucket_dimension=r.bucket_dimension AND p.bucket_value=r.bucket_value AND p.policy_version=r.policy_version
		WHERE r.account_ref=? AND r.bucket_dimension=? AND r.bucket_value=? ORDER BY r.reservation_id`, account, string(dimension), value)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []productionRiskUsageRow
	for rows.Next() {
		var row productionRiskUsageRow
		if err := rows.Scan(&row.ReservationID, &row.PolicyVersion, &row.HeldMinor, &row.FilledMinor, &row.State,
			&row.OverageLatched, &row.UnknownLatched, &row.SnapshotID, &row.PolicyRecordDigest); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func aggregateProductionRiskUsage(rows []productionRiskUsageRow) (string, string, string, error) {
	filled, held := new(big.Int), new(big.Int)
	parts := make([]string, 0, len(rows)*9)
	for _, row := range rows {
		rowFilled, filledOK := new(big.Int).SetString(row.FilledMinor, 10)
		rowHeld, heldOK := new(big.Int).SetString(row.HeldMinor, 10)
		if !filledOK || !heldOK || rowFilled.Sign() < 0 || rowHeld.Sign() < 0 || rowFilled.BitLen() > 256 || rowHeld.BitLen() > 256 ||
			(row.State != "HELD" && row.State != "FILLED" && row.State != "RELEASED") || (row.State == "RELEASED" && rowHeld.Sign() != 0) ||
			row.OverageLatched != 0 || row.UnknownLatched != 0 || row.SnapshotID == "" || row.PolicyRecordDigest == "" ||
			!canonicalIdentity(row.ReservationID) || !canonicalIdentity(row.PolicyVersion) {
			return "", "", "", errors.New("risk bucket: invalid or latched journal usage")
		}
		filled.Add(filled, rowFilled)
		held.Add(held, rowHeld)
		if filled.BitLen() > 256 || held.BitLen() > 256 {
			return "", "", "", errors.New("risk bucket: journal usage overflow")
		}
		parts = append(parts, row.ReservationID, row.PolicyVersion, row.HeldMinor, row.FilledMinor, row.State,
			fmt.Sprint(row.OverageLatched), fmt.Sprint(row.UnknownLatched), row.SnapshotID, row.PolicyRecordDigest)
	}
	return filled.String(), held.String(), productionRiskDigest([]byte(strings.Join(parts, "\x00"))), nil
}

func exactProductionRiskStrategy(values []productionRiskStrategyPolicy, laneID, laneVersion string, horizon Horizon) (productionRiskStrategyPolicy, bool) {
	var found productionRiskStrategyPolicy
	count := 0
	for _, value := range values {
		if value.LaneID == laneID && value.LaneVersion == laneVersion && value.Horizon == horizon {
			found, count = value, count+1
		}
	}
	return found, count == 1 && canonicalIdentity(found.RiskID) && canonicalIdentity(found.RiskVersion)
}

func exactProductionRiskSymbol(values []productionRiskSymbolPolicy, symbol string) (productionRiskSymbolPolicy, bool) {
	var found productionRiskSymbolPolicy
	count := 0
	for _, value := range values {
		if value.Symbol == symbol {
			found, count = value, count+1
		}
	}
	return found, count == 1 && canonicalIdentity(found.Sector)
}

func canonicalProductionRiskTime(raw string) (time.Time, bool) {
	parsed, err := time.Parse(time.RFC3339Nano, raw)
	return parsed.UTC(), err == nil && !parsed.IsZero() && parsed.Location() == time.UTC && parsed.UTC().Format(time.RFC3339Nano) == raw
}

func productionRiskQuoteCurrency(market Market) string {
	if market == MarketKR {
		return "KRW"
	}
	if market == MarketUS {
		return "USD"
	}
	return ""
}

func productionRiskMinorScale(market Market) int {
	if market == MarketKR {
		return 0
	}
	return 2
}

func canonicalRiskDigest(value string) bool {
	if len(value) != len("sha256:")+sha256.Size*2 || !strings.HasPrefix(value, "sha256:") || value != strings.ToLower(value) {
		return false
	}
	decoded, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil && len(decoded) == sha256.Size
}

func productionRiskDigest(data []byte) string {
	digest := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func latestProductionRiskTime(values ...time.Time) time.Time {
	sort.Slice(values, func(i, j int) bool { return values[i].Before(values[j]) })
	return values[len(values)-1]
}

func earliestProductionRiskTime(values ...time.Time) time.Time {
	sort.Slice(values, func(i, j int) bool { return values[i].Before(values[j]) })
	return values[0]
}
