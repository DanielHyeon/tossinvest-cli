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
	"github.com/JungHoonGhae/tossinvest-cli/internal/strategyrouter"
	_ "modernc.org/sqlite"
)

const (
	productionRiskPolicySchema       = "strategy-risk-bucket-policy:v1"
	productionRiskPolicyDomain       = "TossOS/strategy-risk-bucket-policy/ed25519/v1"
	productionRiskPolicyAlgorithm    = "Ed25519"
	productionRiskPolicyMaximumBytes = 1 << 20
	productionRiskSnapshotWindow     = 5 * time.Second
)

// 원장 판독 SQL 은 상수 하나씩 — 판독 전 prepare(a127 D7)와 실행이 같은 문자열을 씀(둘째 철자 금지). 판독 SQL 의 내용은 a127 이 바꾸지 않음(D5).
const (
	// productionRiskScopeLatchSQL 은 범위(계좌 · 시장 · 종목)의 scope latch 수.
	productionRiskScopeLatchSQL = `SELECT count(*) FROM risk_bucket_scope_latches WHERE account_ref=? AND market=? AND symbol=?`
	// productionRiskUsageSQL 은 계좌 bucket 의 예약 행과 떠남 판정의 원장 사실(a126 D1) — readProductionRiskUsage 의 유일한 질의.
	productionRiskUsageSQL = `SELECT r.reservation_id,r.policy_version,r.held_minor,r.filled_minor,r.state,
		r.risk_overage_latched,r.unknown_actual_latched,COALESCE(s.snapshot_id,''),COALESCE(p.record_digest,''),
		CASE WHEN rc.released_at IS NULL THEN 0 ELSE 1 END,COALESCE(rc.released_at,''),COALESCE(ow.released_at,''),
		CASE WHEN EXISTS(SELECT 1 FROM risk_bucket_scope_latches l WHERE l.account_ref=r.account_ref AND l.market=r.market AND l.symbol=r.symbol
			AND l.prospective_generation=r.owner_prospective_generation) THEN 1 ELSE 0 END,
		CASE WHEN d.decision_id IS NOT NULL AND d.account_ref=r.account_ref AND d.market=r.market AND d.symbol=r.symbol
			AND d.owner_prospective_generation=r.owner_prospective_generation THEN 1 ELSE 0 END,
		CASE WHEN EXISTS(SELECT 1 FROM risk_bucket_owner_release_receipts x WHERE x.account_ref=d.account_ref AND x.market=d.market AND x.symbol=d.symbol
			AND x.prospective_generation=d.owner_prospective_generation) THEN 1 ELSE 0 END
		FROM risk_bucket_reservations r
		LEFT JOIN risk_bucket_snapshots s ON s.snapshot_id=r.snapshot_id AND s.bucket_dimension=r.bucket_dimension AND s.bucket_value=r.bucket_value AND s.policy_version=r.policy_version
		LEFT JOIN risk_bucket_policies p ON p.bucket_dimension=r.bucket_dimension AND p.bucket_value=r.bucket_value AND p.policy_version=r.policy_version
		LEFT JOIN risk_bucket_owner_release_receipts rc ON rc.account_ref=r.account_ref AND rc.market=r.market AND rc.symbol=r.symbol AND rc.prospective_generation=r.owner_prospective_generation
		LEFT JOIN risk_bucket_owners ow ON ow.account_ref=r.account_ref AND ow.market=r.market AND ow.symbol=r.symbol AND ow.prospective_generation=r.owner_prospective_generation
		LEFT JOIN risk_bucket_final_decisions d ON d.decision_id=r.decision_id
		WHERE r.account_ref=? AND r.bucket_dimension=? AND r.bucket_value=? ORDER BY r.reservation_id`
)

var ErrProductionRiskSnapshotUnavailable = errors.New("risk bucket: production snapshot authority unavailable")

// ErrProductionRiskScopeRefused 는 실패가 **그 소유자 범위에만** 해당한다는 신원임(a112 5.2.2.2 리뷰 수리 — J4 = (A)).
// 서명 정책에 그 종목의 섹터 매핑이 없거나, 그 범위에 scope latch 가 있는 두 경우만 이것을 감쌈. 그 밖의 실패(원장 읽기 · 행 손상 ·
// 스키마 · 매니페스트 digest · ctx)는 결함이고 이 신원을 갖지 않음 — 엔진은 이 신원이 있을 때만 그 범위를 건너뛰고 같은 주기의 다른 범위로
// 감. 판정은 바뀌지 않음: 어느 쪽이든 이 함수는 거절함(신원만 운반).
var ErrProductionRiskScopeRefused = errors.New("risk bucket: owner scope refused by the signed policy or a scope latch")

// ProductionRiskSnapshotConfig contains only read paths and externally managed
// trust pins. It exposes no policy writer, signer, toggle or execution handle.
type ProductionRiskSnapshotConfig struct {
	ConfigDir, JournalPath       string
	Market                       Market
	AccountID, AccountCurrency   string
	ManifestDigest, TrustedKeyID string
	TrustedKey                   ed25519.PublicKey
	ObservedAt                   time.Time
	// JournalSchemaVersion 은 이 빌드가 이해하는 원장 스키마 버전(호출자가 `journal.SchemaVersion` 을 넣음 — journal 이 이 패키지를 import 하므로
	// 여기서는 읽을 수 없음, a127 D2). 원장 `user_version` 이 이 값과 정확히 같을 때만 판독함(a127 D1).
	JournalSchemaVersion int
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
	// 떠남 판정의 원장 사실(a126 D1) — SQL 이 싣고 aggregateProductionRiskUsage 만 판정함.
	// Receipted: 예약 행의 owner 키(r.*)로 해제 영수증이 있음. DecisionReceipted: 결정 사본의 owner 키(d.*)로 영수증이 있음.
	// ReceiptReleasedAt · OwnerReleasedAt: 영수증 · owner 행의 released_at(없으면 ""). ScopeLatched: 그 owner 키의 scope latch 가 하나라도 있음.
	// OwnerKeyMatches: 예약 행의 owner 키가 결정 사본과 같음.
	Receipted, DecisionReceipted, ScopeLatched, OwnerKeyMatches int
	ReceiptReleasedAt, OwnerReleasedAt                          string
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
	// a127 D2: 원장 스키마 주입 누락(0 이하)은 정책 결속 · 원장 열기 **앞**에서 거절 — 주입 누락은 결함이고, 받아 주면 판독 조건이 사라짐.
	if config.JournalSchemaVersion <= 0 {
		return RiskSnapshotAuthorityBundle{}, fmt.Errorf("%w: journal schema version not injected", ErrProductionRiskSnapshotUnavailable)
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
	// 원인의 신원을 `%w` 로 보존함(문구는 `%v` 와 같음) — 범위 국소 거절(ErrProductionRiskScopeRefused)과 결함을 호출자가 타입으로 가르게.
	scope, reserve, limits, err := bindProductionRiskInputs(config, manifest.productionRiskPolicyBody, input)
	if err != nil {
		return RiskSnapshotAuthorityBundle{}, fmt.Errorf("%w: %w", ErrProductionRiskSnapshotUnavailable, err)
	}
	entries, err := loadProductionRiskEntries(ctx, config, manifest.productionRiskPolicyBody, scope, reserve, limits)
	if err != nil {
		return RiskSnapshotAuthorityBundle{}, fmt.Errorf("%w: %w", ErrProductionRiskSnapshotUnavailable, err)
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
	// a112 6.1(Manager 판정 (C)): risk_id 는 **한 전략군(family)을 함의**한다 — 한 risk_id 를 서로 다른 family 의 레인이 공유하면 family 버킷
	// 경계가 무너지므로 정책 전체를 거절한다. family 는 strategyrouter 정본 표에서 유도하고, 해소되지 않는 레인(이 빌드 · 이 시장 밖)도 거절한다.
	riskFamilies := map[string]strategyrouter.Family{}
	for _, value := range body.Strategies {
		key := value.LaneID + "\x00" + value.LaneVersion + "\x00" + string(value.Horizon)
		if strategyKeys[key] || !canonicalIdentity(value.LaneID) || !canonicalIdentity(value.LaneVersion) ||
			(value.Horizon != HorizonShort && value.Horizon != HorizonMedium) || !canonicalIdentity(value.RiskID) ||
			!canonicalIdentity(value.RiskVersion) {
			return false
		}
		family, known := strategyrouter.ProductionLaneFamily(strategyrouter.Market(body.Market), value.LaneID)
		if prior, seen := riskFamilies[value.RiskID]; !known || seen && prior != family {
			return false
		}
		riskFamilies[value.RiskID] = family
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
		// 그 종목이 서명 정책 밖 — 범위 국소 거절(신원 운반, 판정 불변).
		return RiskSnapshotScope{}, ReservePolicy{}, nil, fmt.Errorf("%w: symbol sector mapping unavailable", ErrProductionRiskScopeRefused)
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
	// a127 D7: 버전 확인 · scope latch · 다섯 사용량 판독을 읽기 전용 트랜잭션 하나에서 — 확인과 판독 사이에 다른 프로세스(예: engine lock 없이
	// 원장을 여는 flatten)의 마이그레이션이 끼어 서로 다른 커밋을 보는 창을 닫음. 판정 · 오류 신원은 그대로.
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	// a127 D1: 원장 스키마는 주입된 현재 버전(engine 이 넣는 journal.SchemaVersion)과 **정확히** 같아야 함 — 같은 프로세스가 방금 연 · 마이그레이션한
	// 원장이 생산의 불변식이고, 더 옛 원장(트리거 · 제약이 그 버전 이전)과 더 새 원장(이 빌드가 모르는 의미)은 둘 다 거절. 방향을 문구로 가름.
	var version int
	if err := tx.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil {
		return nil, fmt.Errorf("risk bucket: journal schema unreadable: %w", err)
	}
	if version > config.JournalSchemaVersion {
		return nil, fmt.Errorf("risk bucket: journal schema %d is newer than this build's %d", version, config.JournalSchemaVersion)
	}
	if version < config.JournalSchemaVersion {
		return nil, fmt.Errorf("risk bucket: journal schema %d is older than this build's %d — the ledger is not migrated", version, config.JournalSchemaVersion)
	}
	// a127 D7: 원장 데이터 질의 전부를 첫 판독 전에 prepare — 사용량 질의는 scope latch 가 0 일 때만 실행되므로(아래 범위 국소 거절), prepare 가
	// 없으면 사용량 전용 열이 없는 원장의 결함이 latch 가 선 범위에서 범위 국소 거절로 재표식됨. 열 부재는 여기서 결함으로 먼저 나옴.
	for _, statement := range []string{productionRiskScopeLatchSQL, productionRiskUsageSQL} {
		prepared, err := tx.PrepareContext(ctx, statement)
		if err != nil {
			return nil, fmt.Errorf("risk bucket: journal read set unavailable: %w", err)
		}
		prepared.Close()
	}
	var scopeLatches int
	// 조회 결함과 latch 존재를 가름(편집 전에는 한 오류로 합쳐 있었음) — 둘 다 거절(판정 불변), latch 만 범위 국소 신원.
	if err := tx.QueryRowContext(ctx, productionRiskScopeLatchSQL, scope.AccountID, string(scope.Market), scope.Symbol).Scan(&scopeLatches); err != nil {
		return nil, fmt.Errorf("risk bucket: scope latch unreadable: %w", err)
	}
	if scopeLatches != 0 {
		return nil, fmt.Errorf("%w: scope latch present", ErrProductionRiskScopeRefused)
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
		usage, err := ReadJournalBucketUsage(ctx, tx, scope.AccountID, dimension, values[dimension])
		if err != nil {
			return nil, err
		}
		// 생산 snapshot 은 latch 된 사용량 위에 서지 않음(수리 전과 같은 거절·같은 문구).
		if usage.Latched {
			return nil, errors.New("risk bucket: invalid or latched journal usage")
		}
		filled, held, rowDigest := usage.FilledMinor, usage.HeldMinor, usage.RowDigest
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

// UsageQueryer 는 *sql.DB 와 *sql.Tx 가 함께 만족하는 읽기 면임 — 같은 사용량 함수를 생산 snapshot reader(읽기 전용 DB)와
// journal admission(쓰기 트랜잭션 안)이 함께 부르게 함.
type UsageQueryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
}

// ErrJournalUsageInvalid 는 원장 사용량 행이 계약을 벗어났다는(값·상태·결속이 잘못됐다는) 답임. 저장소 읽기 실패와 가르기
// 위한 타입 — 체결 경로는 이것을 의미 오류(체결은 보존, latch)로, 읽기 실패는 저장 오류(트랜잭션 되돌림)로 다룸(a066 5.7).
var ErrJournalUsageInvalid = errors.New("risk bucket: invalid or latched journal usage")

// JournalBucketUsage 는 계좌 하나의 bucket(dimension, value) 원장 사용량임.
type JournalBucketUsage struct {
	FilledMinor, HeldMinor, RowDigest string
	// Latched 는 합에 든 예약 중 RISK_OVERAGE/UNKNOWN_ACTUAL_RISK latch 가 있다는 뜻. 합은 latch 와 무관하게 셈.
	Latched bool
	// OverageLatched · UnknownLatched 는 Latched 의 원인을 가름 — 거절이 어떤 latch 인지 이름으로 말하게 함(a066 6.5).
	OverageLatched, UnknownLatched bool
}

// ReadJournalBucketUsage 는 원장 사용량(held+filled)의 **유일한** 계산임(a066 5.6.1 F1). 생산 snapshot reader
// (loadProductionRiskEntries)와 journal admission 의 stale 대조가 둘 다 이 함수를 부름 — 두 쪽의 합이 갈라질 수 없음.
func ReadJournalBucketUsage(ctx context.Context, q UsageQueryer, account string, dimension Dimension, value string) (JournalBucketUsage, error) {
	rows, err := readProductionRiskUsage(ctx, q, account, dimension, value)
	if err != nil {
		return JournalBucketUsage{}, err
	}
	return aggregateProductionRiskUsage(rows)
}

func readProductionRiskUsage(ctx context.Context, db UsageQueryer, account string, dimension Dimension, value string) ([]productionRiskUsageRow, error) {
	// 떠남 판정의 사실(a126 D1)을 행마다 함께 읽음 — 영수증 · owner released_at 은 예약 행의 owner 키(r.*)로, 결정 사본(d.*)의 영수증과
	// 두 사본의 일치는 따로. 규칙 적용은 aggregateProductionRiskUsage 한 곳(판정이 둘이면 서로의 시험을 통과시킴).
	rows, err := db.QueryContext(ctx, productionRiskUsageSQL, account, string(dimension), value)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var result []productionRiskUsageRow
	for rows.Next() {
		var row productionRiskUsageRow
		if err := rows.Scan(&row.ReservationID, &row.PolicyVersion, &row.HeldMinor, &row.FilledMinor, &row.State,
			&row.OverageLatched, &row.UnknownLatched, &row.SnapshotID, &row.PolicyRecordDigest,
			&row.Receipted, &row.ReceiptReleasedAt, &row.OwnerReleasedAt, &row.ScopeLatched, &row.OwnerKeyMatches, &row.DecisionReceipted); err != nil {
			return nil, err
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func aggregateProductionRiskUsage(rows []productionRiskUsageRow) (JournalBucketUsage, error) {
	filled, held := new(big.Int), new(big.Int)
	latched, overage, unknown := false, false, false
	parts := make([]string, 0, len(rows)*9)
	for _, row := range rows {
		rowFilled, filledOK := new(big.Int).SetString(row.FilledMinor, 10)
		rowHeld, heldOK := new(big.Int).SetString(row.HeldMinor, 10)
		if !filledOK || !heldOK || rowFilled.Sign() < 0 || rowHeld.Sign() < 0 || rowFilled.BitLen() > 256 || rowHeld.BitLen() > 256 ||
			(row.State != "HELD" && row.State != "FILLED" && row.State != "RELEASED") || (row.State == "RELEASED" && rowHeld.Sign() != 0) ||
			row.SnapshotID == "" || row.PolicyRecordDigest == "" ||
			!canonicalIdentity(row.ReservationID) || !canonicalIdentity(row.PolicyVersion) {
			return JournalBucketUsage{}, ErrJournalUsageInvalid
		}
		// a126 D1 · D2: 영수증이 있는 owner 의 행은 떠남 여부와 무관하게(scope latch 로 되돌려진 행 포함) 먼저 손상 · 불일치를 거절함 —
		// 해제는 bucket_held=0 을 요구했으므로 HELD · held≠0 은 손상, 영수증과 owner released_at 이 어긋나거나 예약 행의 owner 키가 결정
		// 사본과 다르면 떠남 판정과 소유 판정이 서로 다른 owner 를 보게 됨. 모르는 것은 "안 떠남" 이 아니라 판독 불가(fail-closed).
		if (row.Receipted != 0 || row.DecisionReceipted != 0) && row.OwnerKeyMatches == 0 {
			return JournalBucketUsage{}, fmt.Errorf("%w: reservation %s owner key diverges from its decision", ErrJournalUsageInvalid, row.ReservationID)
		}
		if row.Receipted != 0 && (row.OwnerReleasedAt == "" || row.OwnerReleasedAt != row.ReceiptReleasedAt || row.State == "HELD" || rowHeld.Sign() != 0) {
			return JournalBucketUsage{}, fmt.Errorf("%w: reservation %s of a released owner is not settled", ErrJournalUsageInvalid, row.ReservationID)
		}
		// 떠남 = 영수증 ∧ scope latch 없음(영수증 뒤 ORPHAN_FILL 등 scope latch 가 서면 되돌림 — D4). 떠난 행은 **합에서만** 빠짐.
		departed := row.Receipted != 0 && row.ScopeLatched == 0
		// latch 는 합에서 빼지 않고 호출자에게 알림 — 생산 snapshot 은 거절하고, admission 대조는 합만 씀. 떠난 행의 플래그도 셈(Q3).
		latched = latched || row.OverageLatched != 0 || row.UnknownLatched != 0
		overage = overage || row.OverageLatched != 0
		unknown = unknown || row.UnknownLatched != 0
		if !departed {
			filled.Add(filled, rowFilled)
			held.Add(held, rowHeld)
		}
		if filled.BitLen() > 256 || held.BitLen() > 256 {
			return JournalBucketUsage{}, fmt.Errorf("%w: journal usage overflow", ErrJournalUsageInvalid)
		}
		parts = append(parts, row.ReservationID, row.PolicyVersion, row.HeldMinor, row.FilledMinor, row.State,
			fmt.Sprint(row.OverageLatched), fmt.Sprint(row.UnknownLatched), row.SnapshotID, row.PolicyRecordDigest)
	}
	return JournalBucketUsage{FilledMinor: filled.String(), HeldMinor: held.String(),
		RowDigest: productionRiskDigest([]byte(strings.Join(parts, "\x00"))), Latched: latched, OverageLatched: overage, UnknownLatched: unknown}, nil
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
