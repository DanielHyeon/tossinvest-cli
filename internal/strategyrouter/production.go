package strategyrouter

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
	"net/url"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/breakoutlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/continuationlane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reversallane"
	"github.com/JungHoonGhae/tossinvest-cli/internal/weeklyvaluelane"
	_ "modernc.org/sqlite"
)

const (
	// 스키마 이름이 v2 로 올라간 이유: 이 매니페스트는 이제 가족 신원과
	// 중재 점수·보정 봉인을 반드시 실어야 한다. 이름을 그대로 두면 세 가족짜리
	// 옛 매니페스트가 "그냥 형식이 안 맞는 파일"로만 거절되고, 왜 거절됐는지
	// 사람이 읽을 수 없다. 이름을 바꾸면 거절이 계약이 된다.
	productionRouteSchema       = "strategy-lane-authority:v2"
	productionRouteDomain       = "TossOS/strategy-router-lane-authority/ed25519/v1"
	productionRouteAlgorithm    = "Ed25519"
	productionRouteMaximumBytes = 1 << 20
	productionRouteJournalV     = 27
	productionRouteMaxOwners    = 10_000
	// 중재 점수는 0..1,000,000 의 정수 ppm 이다(design.md "score_ppm(0..1,000,000)").
	productionRouteScorePPMMax = 1_000_000
)

// 세 봉인은 서로 다른 도메인 문자열로 시작한다. 같은 재료를 해시해도
// 다른 봉인끼리 같은 값이 나올 수 없게 하려는 것이다.
const (
	productionRouteFamilySealDomain      = "TossOS/strategy-router-family-seal/sha256/v1"
	productionRouteScoringSealDomain     = "TossOS/strategy-router-scoring-seal/sha256/v1"
	productionRouteCalibrationSealDomain = "TossOS/strategy-router-calibration-seal/sha256/v1"
)

// Family 는 전략군 신원이다. 정확히 이 네 개뿐이고, 시장마다 네 개가 모두
// 한 번씩 있어야 한다(design.md "family enum 은 ... exact set 이다").
type Family string

const (
	FamilyContinuation   Family = "CONTINUATION"
	FamilyReversal       Family = "REVERSAL"
	FamilyWeeklyValue    Family = "WEEKLY_VALUE"
	FamilyBreakoutRetest Family = "BREAKOUT_RETEST"
)

var ErrProductionRouteUnavailable = errors.New("strategyrouter: production route authority unavailable")

// ProductionRouteConfig contains only immutable read paths, external trust
// pins, and the scheduler facts already verified privately by the engine.
type ProductionRouteConfig struct {
	ConfigDir, JournalPath                 string
	AccountRef, Symbol                     string
	Market                                 Market
	PositionGeneration                     uint64
	ManifestDigest, TrustedKeyID           string
	TrustedKey                             ed25519.PublicKey
	ObservedAt                             time.Time
	ActivationDigest, CalendarGeneration   string
	CalendarDigest, SchedulerConfigVersion string
	ActivationExpiresAt                    time.Time
}

// productionRouteCandidate 는 매니페스트가 서명한 한 가족의 후보다.
// 원시 int64 Score 는 사라졌다 — 단위가 다른 신호를 견주는 값이라 가족 간
// 선택 권한이 될 수 없기 때문이다. 대신 승인된 점수 버전에 묶인 ScorePPM 만 쓴다.
type productionRouteCandidate struct {
	Family         Family       `json:"family"`
	Horizon        Horizon      `json:"horizon"`
	LaneID         string       `json:"lane_id"`
	LaneVersion    string       `json:"lane_version"`
	ScorePPM       uint32       `json:"score_ppm"`
	Eligible       bool         `json:"eligible"`
	Desired        DesiredState `json:"desired"`
	Effective      DesiredState `json:"effective"`
	EvidenceDigest string       `json:"evidence_digest"`
	ConfigDigest   string       `json:"config_digest"`
}

type productionRouteScope struct {
	Symbol             string                     `json:"symbol"`
	PositionGeneration uint64                     `json:"position_generation"`
	OwnerRevision      uint64                     `json:"owner_revision"`
	Candidates         []productionRouteCandidate `json:"candidates"`
}

type productionRouteBody struct {
	SchemaVersion       string `json:"schema_version"`
	Domain              string `json:"domain"`
	SignatureAlgorithm  string `json:"signature_algorithm"`
	KeyID               string `json:"key_id"`
	Generation          uint64 `json:"generation"`
	AccountRef          string `json:"account_ref"`
	Market              Market `json:"market"`
	MarketRevision      uint64 `json:"market_revision"`
	ActivationDigest    string `json:"activation_digest"`
	ActivationExpiresAt string `json:"activation_expires_at"`
	CalendarGeneration  string `json:"calendar_generation"`
	CalendarDigest      string `json:"calendar_digest"`
	Timezone            string `json:"timezone"`
	SessionScope        string `json:"session_scope"`
	ConfigVersion       string `json:"config_version"`
	// 아래 두 값이 승인된 공통 채점 기준이다. 둘 중 하나라도 없으면
	// 서로 다른 전략군의 점수를 견줄 근거가 없으므로 매니페스트 전체를 거절한다.
	ArbitrationScoreVersion string                 `json:"arbitration_score_version"`
	CalibrationDigest       string                 `json:"calibration_digest"`
	Actor                   string                 `json:"actor"`
	ObservedAt              string                 `json:"observed_at"`
	FreshUntil              string                 `json:"fresh_until"`
	Revoked                 bool                   `json:"revoked"`
	Scopes                  []productionRouteScope `json:"scopes"`
}

type productionRouteManifest struct {
	productionRouteBody
	Signature string `json:"signature"`
}

// ProductionRouteCalibration 은 이 시장 매니페스트가 승인한 공통 채점 기준이다.
// 두 값이 함께 있어야 서로 다른 전략군의 ScorePPM 을 견줄 수 있다.
type ProductionRouteCalibration struct {
	ScoreVersion      string
	CalibrationDigest string
}

// ProductionRouteFamilyScore 는 한 가족의 봉인된 중재 점수다.
type ProductionRouteFamilyScore struct {
	Family      Family
	Horizon     Horizon
	LaneID      string
	LaneVersion string
	ScorePPM    uint32
}

// ProductionRouteSeals 는 서로 다른 것을 증명하는 세 봉인이다.
//
//	Family      — 어떤 가족 네 개가 어떤 레인/호라이즌에 묶였는가
//	Scoring     — 어떤 점수 버전 아래 어떤 ScorePPM 을 견주는가
//	Calibration — 그 점수 버전이 어떤 보정 다이제스트에 묶였는가
//
// 셋을 하나로 합치지 않는 이유는, 가족 구성이 그대로인데 보정만 바뀐 경우와
// 그 반대를 뒤에서 구별할 수 있어야 하기 때문이다.
type ProductionRouteSeals struct {
	Family      string
	Scoring     string
	Calibration string
}

// ProductionRouteAuthority exposes a sealed route request plus scalar
// provenance. It contains no writer, signer, activation, or transport handle.
type ProductionRouteAuthority struct {
	request                     RouteRequest
	manifestDigest, ownerDigest string
	calibration                 ProductionRouteCalibration
	scores                      []ProductionRouteFamilyScore
	seals                       ProductionRouteSeals
}

func (authority ProductionRouteAuthority) Request() RouteRequest  { return authority.request }
func (authority ProductionRouteAuthority) ManifestDigest() string { return authority.manifestDigest }
func (authority ProductionRouteAuthority) OwnerDigest() string    { return authority.ownerDigest }

// Calibration 은 승인된 공통 채점 기준을 돌려준다.
func (authority ProductionRouteAuthority) Calibration() ProductionRouteCalibration {
	return authority.calibration
}

// Seals 는 세 봉인을 돌려준다.
func (authority ProductionRouteAuthority) Seals() ProductionRouteSeals { return authority.seals }

// FamilyScores 는 가족 순서로 정렬된 점수 행을 복사해서 돌려준다.
// 복사본이라 밖에서 고쳐도 봉인된 원본은 바뀌지 않는다.
func (authority ProductionRouteAuthority) FamilyScores() []ProductionRouteFamilyScore {
	return append([]ProductionRouteFamilyScore(nil), authority.scores...)
}

// SealsValid 는 세 봉인이 지금 들고 있는 보정/점수 행과 정확히 맞는지 다시 계산해
// 확인한다. 봉인만 있고 재료가 바뀌었거나, 재료만 채우고 봉인을 다시 만들지 않은
// 권한 객체는 여기서 거절된다.
func (authority ProductionRouteAuthority) SealsValid() bool {
	if len(authority.scores) == 0 || !productionRouteIdentity(authority.calibration.ScoreVersion) ||
		!productionRouteIdentity(authority.calibration.CalibrationDigest) {
		return false
	}
	market := authority.request.Key.Market
	return authority.seals.Family == productionRouteFamilySeal(market, authority.scores) &&
		authority.seals.Scoring == productionRouteScoringSeal(authority.calibration, authority.scores) &&
		authority.seals.Calibration == productionRouteCalibrationSeal(market, authority.calibration)
}

// productionRouteFamilyScores 는 서명된 후보를 가족 순서로 봉인 재료로 바꾼다.
// 매니페스트에 적힌 순서가 아니라 가족 이름 순서를 쓰므로, 같은 내용이면
// 파일에 적힌 차례가 달라도 같은 봉인이 나온다.
func productionRouteFamilyScores(candidates []productionRouteCandidate) []ProductionRouteFamilyScore {
	values := make([]ProductionRouteFamilyScore, 0, len(candidates))
	for _, candidate := range candidates {
		values = append(values, ProductionRouteFamilyScore{Family: candidate.Family, Horizon: candidate.Horizon,
			LaneID: candidate.LaneID, LaneVersion: candidate.LaneVersion, ScorePPM: candidate.ScorePPM})
	}
	sort.Slice(values, func(i, j int) bool { return values[i].Family < values[j].Family })
	return values
}

// productionRouteFamilySeal 은 가족 구성의 신원이다. 가족 하나가 빠지거나
// 다른 레인에 붙으면 값이 달라진다.
func productionRouteFamilySeal(market Market, scores []ProductionRouteFamilyScore) string {
	h := sha256.New()
	writeString(h, productionRouteFamilySealDomain)
	writeString(h, string(market))
	writeUint64(h, uint64(len(scores)))
	for _, score := range scores {
		writeString(h, string(score.Family))
		writeString(h, string(score.Horizon))
		writeString(h, score.LaneID)
		writeString(h, score.LaneVersion)
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil))
}

// productionRouteScoringSeal 은 "어떤 버전 아래 어떤 점수를 견줬는가"의 신원이다.
func productionRouteScoringSeal(calibration ProductionRouteCalibration, scores []ProductionRouteFamilyScore) string {
	h := sha256.New()
	writeString(h, productionRouteScoringSealDomain)
	writeString(h, calibration.ScoreVersion)
	writeUint64(h, uint64(len(scores)))
	for _, score := range scores {
		writeString(h, string(score.Family))
		writeUint64(h, uint64(score.ScorePPM))
	}
	return "sha256:" + hex.EncodeToString(h.Sum(nil))
}

// productionRouteCalibrationSeal 은 점수 버전과 보정 다이제스트의 묶음 신원이다.
func productionRouteCalibrationSeal(market Market, calibration ProductionRouteCalibration) string {
	h := sha256.New()
	writeString(h, productionRouteCalibrationSealDomain)
	writeString(h, string(market))
	writeString(h, calibration.ScoreVersion)
	writeString(h, calibration.CalibrationDigest)
	return "sha256:" + hex.EncodeToString(h.Sum(nil))
}

// ProductionRouteTarget identifies one approved symbol scope requested from a
// single frozen market snapshot. PositionGeneration zero selects the unique
// signed generation for the symbol and refuses ambiguity.
type ProductionRouteTarget struct {
	Symbol             string
	PositionGeneration uint64
}

// ProductionRouteBatchAuthority is a sealed, immutable-by-copy set of symbol
// authorities reconstructed from one manifest read and one SQLite read
// transaction. Missing signed scopes are absent rather than synthesized.
type ProductionRouteBatchAuthority struct {
	values         map[string]ProductionRouteAuthority
	manifestDigest string
}

func (authority ProductionRouteBatchAuthority) Len() int { return len(authority.values) }

func (authority ProductionRouteBatchAuthority) For(symbol string) (ProductionRouteAuthority, bool) {
	value, ok := authority.values[strings.ToUpper(strings.TrimSpace(symbol))]
	return value, ok
}

func (authority ProductionRouteBatchAuthority) ManifestDigest() string {
	return authority.manifestDigest
}

func ProductionRouteFileName(market Market) string {
	if market == MarketKR {
		return "strategy-lane-authority-KR.json"
	}
	if market == MarketUS {
		return "strategy-lane-authority-US.json"
	}
	return ""
}

// LoadProductionRouteAuthority verifies one signed market manifest and
// reconstructs owner state from the existing journal through a read-only
// connection. It never creates, migrates, or writes either source.
func LoadProductionRouteAuthority(ctx context.Context, config ProductionRouteConfig) (ProductionRouteAuthority, error) {
	batch, err := LoadProductionRouteAuthorityBatch(ctx, config, []ProductionRouteTarget{{
		Symbol: config.Symbol, PositionGeneration: config.PositionGeneration,
	}})
	if err != nil {
		return ProductionRouteAuthority{}, err
	}
	authority, ok := batch.For(config.Symbol)
	if !ok {
		return ProductionRouteAuthority{}, ErrProductionRouteUnavailable
	}
	return authority, nil
}

// LoadProductionRouteAuthorityBatch verifies one signed market manifest and
// reconstructs every requested symbol inside one read-only SQLite transaction.
// It never creates, migrates, or writes either source. A requested symbol that
// is absent from the signed scope set is simply refused; any global manifest or
// journal-integrity failure refuses the entire market snapshot.
func LoadProductionRouteAuthorityBatch(ctx context.Context, config ProductionRouteConfig, targets []ProductionRouteTarget) (ProductionRouteBatchAuthority, error) {
	if ctx == nil || config.ObservedAt.IsZero() {
		return ProductionRouteBatchAuthority{}, ErrProductionRouteUnavailable
	}
	if err := ctx.Err(); err != nil {
		return ProductionRouteBatchAuthority{}, err
	}
	config = canonicalProductionRouteConfig(config)
	targets, ok := canonicalProductionRouteTargets(targets)
	ownerUID, ownerOK := productionRouteOwnerUID()
	name := ProductionRouteFileName(config.Market)
	if !ownerOK || name == "" || !filepath.IsAbs(config.ConfigDir) || !filepath.IsAbs(config.JournalPath) ||
		!ok || !validProductionRouteBaseConfig(config) || !productionRouteDigestValid(config.ManifestDigest) ||
		!productionRouteIdentity(config.TrustedKeyID) || len(config.TrustedKey) != ed25519.PublicKeySize ||
		!productionRouteIdentity(config.ActivationDigest) || !productionRouteIdentity(config.CalendarGeneration) ||
		!productionRouteIdentity(config.CalendarDigest) || !productionRouteIdentity(config.SchedulerConfigVersion) {
		return ProductionRouteBatchAuthority{}, ErrProductionRouteUnavailable
	}
	data, err := readProductionRouteFile(filepath.Join(config.ConfigDir, name), ownerUID, 0o400, productionRouteMaximumBytes)
	if err != nil || productionRouteDigest(data) != config.ManifestDigest {
		return ProductionRouteBatchAuthority{}, ErrProductionRouteUnavailable
	}
	manifest, err := decodeProductionRouteManifest(data)
	if err != nil || len(manifest.Scopes) == 0 {
		return ProductionRouteBatchAuthority{}, ErrProductionRouteUnavailable
	}
	verificationConfig := config
	verificationConfig.Symbol = manifest.Scopes[0].Symbol
	verificationConfig.PositionGeneration = manifest.Scopes[0].PositionGeneration
	if _, verified := verifyProductionRouteManifest(manifest, verificationConfig); !verified {
		return ProductionRouteBatchAuthority{}, ErrProductionRouteUnavailable
	}
	db, tx, err := openProductionRouteSnapshot(ctx, config.JournalPath, ownerUID)
	if err != nil {
		return ProductionRouteBatchAuthority{}, fmt.Errorf("%w: owner snapshot", ErrProductionRouteUnavailable)
	}
	defer db.Close()
	defer tx.Rollback()
	observed, _ := productionRouteTime(manifest.ObservedAt)
	fresh, _ := productionRouteTime(manifest.FreshUntil)
	activationExpires, _ := productionRouteTime(manifest.ActivationExpiresAt)
	record, err := newMarketRecord(marketRecordInput{Market: config.Market, Desired: StateOn, Effective: StateOn,
		Revision: manifest.MarketRevision, LockID: "production-route:" + string(config.Market) + ":" + strconv.FormatUint(manifest.Generation, 10),
		CalendarGeneration: manifest.CalendarGeneration, CalendarDigest: manifest.CalendarDigest, Timezone: manifest.Timezone,
		SessionScope: manifest.SessionScope, ActivationDigest: manifest.ActivationDigest, ActivationExpiresAt: activationExpires,
		ConfigVersion: manifest.ConfigVersion, UpdatedActor: manifest.Actor, UpdatedAt: observed, Runtime: RuntimeUnobserved})
	if err != nil || EvaluateMarketLifecycle(record, config.ObservedAt) != LifecycleReady {
		return ProductionRouteBatchAuthority{}, ErrProductionRouteUnavailable
	}
	values := make(map[string]ProductionRouteAuthority, len(targets))
	for _, target := range targets {
		if err := ctx.Err(); err != nil {
			return ProductionRouteBatchAuthority{}, err
		}
		scope, found := validProductionRouteScopes(manifest.Market, manifest.Scopes, target.Symbol, target.PositionGeneration)
		if !found {
			continue
		}
		key, err := NewOwnerKey(config.AccountRef, config.Market, target.Symbol, scope.PositionGeneration)
		if err != nil {
			return ProductionRouteBatchAuthority{}, ErrProductionRouteUnavailable
		}
		owners, revision, ownerDigest, err := loadProductionRouteOwnersFrom(ctx, tx, key)
		if err != nil || revision != scope.OwnerRevision {
			return ProductionRouteBatchAuthority{}, fmt.Errorf("%w: owner reconstruction", ErrProductionRouteUnavailable)
		}
		snapshot, err := newOwnerSnapshot(key, revision, ownerDigest, observed, fresh, owners)
		if err != nil {
			return ProductionRouteBatchAuthority{}, ErrProductionRouteUnavailable
		}
		candidates := make([]Candidate, 0, len(scope.Candidates))
		for _, value := range scope.Candidates {
			// Score 를 일부러 채우지 않는다. 원시 점수를 넘기지 않으면
			// 레인 선택은 평가 뒤 조정자의 보정 점수 말고는 할 방법이 없다.
			candidates = append(candidates, Candidate{Key: key, Horizon: value.Horizon, LaneID: value.LaneID, LaneVersion: value.LaneVersion,
				Eligible: value.Eligible, Desired: value.Desired, Effective: value.Effective,
				EvidenceDigest: value.EvidenceDigest, ConfigDigest: value.ConfigDigest})
		}
		request := RouteRequest{Key: key, ExpectedOwnerRevision: revision, ExpectedMarketRevision: record.Revision,
			EvaluatedAt: config.ObservedAt, Snapshot: snapshot, MarketRecord: record, Candidates: candidates}
		calibration := ProductionRouteCalibration{ScoreVersion: manifest.ArbitrationScoreVersion, CalibrationDigest: manifest.CalibrationDigest}
		scores := productionRouteFamilyScores(scope.Candidates)
		values[target.Symbol] = ProductionRouteAuthority{request: request, manifestDigest: config.ManifestDigest, ownerDigest: ownerDigest,
			calibration: calibration, scores: scores, seals: ProductionRouteSeals{
				Family:      productionRouteFamilySeal(config.Market, scores),
				Scoring:     productionRouteScoringSeal(calibration, scores),
				Calibration: productionRouteCalibrationSeal(config.Market, calibration),
			}}
	}
	if err := tx.Commit(); err != nil {
		return ProductionRouteBatchAuthority{}, fmt.Errorf("%w: owner snapshot close", ErrProductionRouteUnavailable)
	}
	return ProductionRouteBatchAuthority{values: values, manifestDigest: config.ManifestDigest}, nil
}

func canonicalProductionRouteConfig(config ProductionRouteConfig) ProductionRouteConfig {
	config.ConfigDir = filepath.Clean(strings.TrimSpace(config.ConfigDir))
	config.JournalPath = filepath.Clean(strings.TrimSpace(config.JournalPath))
	config.AccountRef = strings.TrimSpace(config.AccountRef)
	config.Symbol = strings.ToUpper(strings.TrimSpace(config.Symbol))
	config.ManifestDigest = strings.TrimSpace(config.ManifestDigest)
	config.TrustedKeyID = strings.TrimSpace(config.TrustedKeyID)
	config.TrustedKey = append(ed25519.PublicKey(nil), config.TrustedKey...)
	config.ObservedAt = config.ObservedAt.UTC()
	config.ActivationExpiresAt = config.ActivationExpiresAt.UTC()
	return config
}

func validOwnerKeyValue(config ProductionRouteConfig) bool {
	return productionRouteIdentity(config.AccountRef) && validMarket(config.Market) && config.Symbol != "" &&
		config.Symbol == strings.ToUpper(strings.TrimSpace(config.Symbol))
}

func validProductionRouteBaseConfig(config ProductionRouteConfig) bool {
	return productionRouteIdentity(config.AccountRef) && validMarket(config.Market)
}

func canonicalProductionRouteTargets(targets []ProductionRouteTarget) ([]ProductionRouteTarget, bool) {
	if len(targets) == 0 || len(targets) > 10_000 {
		return nil, false
	}
	values := append([]ProductionRouteTarget{}, targets...)
	for index := range values {
		values[index].Symbol = strings.ToUpper(strings.TrimSpace(values[index].Symbol))
		if values[index].Symbol == "" {
			return nil, false
		}
	}
	sort.Slice(values, func(i, j int) bool { return values[i].Symbol < values[j].Symbol })
	for index := range values {
		if index > 0 && values[index-1].Symbol == values[index].Symbol {
			return nil, false
		}
	}
	return values, true
}

func decodeProductionRouteManifest(data []byte) (productionRouteManifest, error) {
	if len(data) == 0 || len(data) > productionRouteMaximumBytes {
		return productionRouteManifest{}, ErrProductionRouteUnavailable
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var manifest productionRouteManifest
	if err := decoder.Decode(&manifest); err != nil {
		return productionRouteManifest{}, err
	}
	if err := decoder.Decode(&struct{}{}); err != io.EOF {
		return productionRouteManifest{}, ErrProductionRouteUnavailable
	}
	canonical, err := json.Marshal(manifest)
	if err != nil || !bytes.Equal(canonical, data) {
		return productionRouteManifest{}, ErrProductionRouteUnavailable
	}
	return manifest, nil
}

func verifyProductionRouteManifest(manifest productionRouteManifest, config ProductionRouteConfig) (productionRouteScope, bool) {
	body := manifest.productionRouteBody
	activationExpires, activationOK := productionRouteTime(body.ActivationExpiresAt)
	observed, observedOK := productionRouteTime(body.ObservedAt)
	fresh, freshOK := productionRouteTime(body.FreshUntil)
	if body.SchemaVersion != productionRouteSchema || body.Domain != productionRouteDomain || body.SignatureAlgorithm != productionRouteAlgorithm ||
		body.KeyID != config.TrustedKeyID || body.Generation == 0 || body.AccountRef != config.AccountRef || body.Market != config.Market ||
		body.MarketRevision == 0 ||
		body.ActivationDigest != config.ActivationDigest || body.CalendarGeneration != config.CalendarGeneration ||
		body.CalendarDigest != config.CalendarDigest || body.ConfigVersion != config.SchedulerConfigVersion || body.SessionScope != "REGULAR" ||
		// 보정 봉인: 승인된 점수 버전과 보정 다이제스트가 없으면 이 매니페스트는
		// 가족 간 비교의 권한이 아니다. 하나라도 비면 시장 전체를 거절한다.
		!productionRouteIdentity(body.ArbitrationScoreVersion) || !productionRouteIdentity(body.CalibrationDigest) ||
		body.Timezone != map[Market]string{MarketKR: "Asia/Seoul", MarketUS: "America/New_York"}[config.Market] ||
		!productionRouteIdentity(body.Actor) || body.Revoked || !activationOK || !observedOK || !freshOK ||
		(!config.ActivationExpiresAt.IsZero() && !activationExpires.Equal(config.ActivationExpiresAt)) || observed.After(config.ObservedAt) || !config.ObservedAt.Before(fresh) ||
		!config.ObservedAt.Before(activationExpires) || observed.After(fresh) {
		return productionRouteScope{}, false
	}
	scope, scopeOK := validProductionRouteScopes(body.Market, body.Scopes, config.Symbol, config.PositionGeneration)
	if !scopeOK {
		return productionRouteScope{}, false
	}
	canonical, err := json.Marshal(body)
	if err != nil {
		return productionRouteScope{}, false
	}
	signature, err := base64.StdEncoding.Strict().DecodeString(manifest.Signature)
	verified := err == nil && base64.StdEncoding.EncodeToString(signature) == manifest.Signature && len(signature) == ed25519.SignatureSize &&
		ed25519.Verify(config.TrustedKey, canonical, signature)
	return scope, verified
}

func validProductionRouteScopes(market Market, scopes []productionRouteScope, symbol string, generation uint64) (productionRouteScope, bool) {
	if len(scopes) == 0 || len(scopes) > 10_000 {
		return productionRouteScope{}, false
	}
	var selected productionRouteScope
	found := false
	previous := ""
	for _, scope := range scopes {
		key := scope.Symbol + "\x00" + strconv.FormatUint(scope.PositionGeneration, 10)
		if scope.Symbol == "" || scope.Symbol != strings.ToUpper(strings.TrimSpace(scope.Symbol)) || scope.PositionGeneration == 0 ||
			scope.OwnerRevision == 0 || key <= previous || !validProductionRouteCandidates(market, scope.Candidates) {
			return productionRouteScope{}, false
		}
		previous = key
		if scope.Symbol == symbol && (generation == 0 || scope.PositionGeneration == generation) {
			if found {
				return productionRouteScope{}, false
			}
			selected, found = scope, true
		}
	}
	return selected, found
}

func validProductionRouteCandidates(market Market, values []productionRouteCandidate) bool {
	want := productionRouteDescriptors(market)
	if len(values) != len(want) {
		return false
	}
	seen := make(map[string]bool, len(values))
	for _, value := range values {
		descriptor, ok := want[value.LaneID]
		// 가족 봉인: 매니페스트가 말하는 가족이 표가 그 레인에 묶어 둔 가족과 같아야 한다.
		// "네 가족이 한 번씩"은 여기서 또 세지 않는다 — 레인이 중복 없이 네 개이고
		// 레인마다 표의 가족을 그대로 써야 하므로, 여기서 세는 조건은 어떤 입력으로도
		// 거짓이 될 수 없다. 표 자체가 네 가족을 한 번씩 담는다는 사실은
		// TestProductionRouteDescriptorsCoverFourFamiliesPerMarket 가 지킨다.
		// 채점 봉인: ppm 은 승인된 0..1,000,000 범위 안이어야 한다.
		if !ok || seen[value.LaneID] || descriptor.Horizon != value.Horizon || descriptor.LaneVersion != value.LaneVersion ||
			descriptor.Family != value.Family || value.ScorePPM > productionRouteScorePPMMax ||
			!validDesiredState(value.Desired) || !validDesiredState(value.Effective) ||
			!productionRouteIdentity(value.EvidenceDigest) || !productionRouteIdentity(value.ConfigDigest) {
			return false
		}
		seen[value.LaneID] = true
	}
	return len(seen) == len(want)
}

type productionLaneDescriptor struct {
	Family      Family
	Horizon     Horizon
	LaneVersion string
}

func productionRouteDescriptors(market Market) map[string]productionLaneDescriptor {
	if market == MarketKR {
		return map[string]productionLaneDescriptor{
			continuationlane.KRContinuationLaneID: {FamilyContinuation, HorizonShort, continuationlane.LaneVersionV1},
			reversallane.KRReversalLaneID:         {FamilyReversal, HorizonShort, reversallane.LaneVersionV1},
			weeklyvaluelane.KRWeeklyLaneID:        {FamilyWeeklyValue, HorizonWeekly, weeklyvaluelane.LaneVersionV1},
			breakoutlane.KRLaneID:                 {FamilyBreakoutRetest, HorizonShort, breakoutlane.LaneVersionV1},
		}
	}
	if market == MarketUS {
		return map[string]productionLaneDescriptor{
			continuationlane.USContinuationLaneID: {FamilyContinuation, HorizonShort, continuationlane.LaneVersionV1},
			reversallane.USReversalLaneID:         {FamilyReversal, HorizonShort, reversallane.LaneVersionV1},
			weeklyvaluelane.USWeeklyLaneID:        {FamilyWeeklyValue, HorizonWeekly, weeklyvaluelane.LaneVersionV1},
			breakoutlane.USLaneID:                 {FamilyBreakoutRetest, HorizonShort, breakoutlane.LaneVersionV1},
		}
	}
	return nil
}

type productionOwnerRow struct {
	prospective, laneID, campaignID, actual, acquired, released string
	overage, unknown                                            int
}

type productionRouteQueryer interface {
	QueryContext(context.Context, string, ...any) (*sql.Rows, error)
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func openProductionRouteSnapshot(ctx context.Context, journalPath string, ownerUID uint32) (*sql.DB, *sql.Tx, error) {
	if err := validateProductionRouteJournalFile(journalPath, ownerUID); err != nil {
		return nil, nil, err
	}
	dsn := url.URL{Scheme: "file", Path: journalPath, RawQuery: "mode=ro&_pragma=query_only(1)&_pragma=busy_timeout(250)"}
	db, err := sql.Open("sqlite", dsn.String())
	if err != nil {
		return nil, nil, err
	}
	db.SetMaxOpenConns(1)
	tx, err := db.BeginTx(ctx, &sql.TxOptions{ReadOnly: true})
	if err != nil {
		db.Close()
		return nil, nil, err
	}
	var version int
	if err := tx.QueryRowContext(ctx, `PRAGMA user_version`).Scan(&version); err != nil || version != productionRouteJournalV {
		tx.Rollback()
		db.Close()
		return nil, nil, ErrProductionRouteUnavailable
	}
	return db, tx, nil
}

func loadProductionRouteOwnersFrom(ctx context.Context, queryer productionRouteQueryer, key OwnerKey) ([]Owner, uint64, string, error) {
	if queryer == nil {
		return nil, 0, "", ErrProductionRouteUnavailable
	}
	rows, err := queryer.QueryContext(ctx, `SELECT prospective_generation,lane_id,campaign_id,coalesce(actual_generation,''),acquired_at,coalesce(released_at,''),risk_overage_latched,unknown_actual_latched FROM risk_bucket_owners WHERE account_ref=? AND market=? AND symbol=? ORDER BY acquired_at,prospective_generation`, key.AccountRef, string(key.Market), key.Symbol)
	if err != nil {
		return nil, 0, "", err
	}
	defer rows.Close()
	history := make([]productionOwnerRow, 0, 4)
	for rows.Next() {
		if len(history) >= productionRouteMaxOwners {
			return nil, 0, "", ErrProductionRouteUnavailable
		}
		var value productionOwnerRow
		if err := rows.Scan(&value.prospective, &value.laneID, &value.campaignID, &value.actual, &value.acquired, &value.released, &value.overage, &value.unknown); err != nil ||
			!validProductionOwnerRow(value) {
			return nil, 0, "", ErrProductionRouteUnavailable
		}
		history = append(history, value)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, "", err
	}
	digest := productionOwnerHistoryDigest(history)
	revision := uint64(len(history)) + 1
	active := make([]productionOwnerRow, 0, 1)
	for _, value := range history {
		if value.released == "" {
			active = append(active, value)
		}
	}
	if len(active) > 1 {
		return nil, 0, "", ErrProductionRouteUnavailable
	}
	if len(active) == 0 {
		return nil, revision, digest, nil
	}
	value := active[0]
	if value.actual == "" || value.overage != 0 || value.unknown != 0 {
		return nil, 0, "", ErrProductionRouteUnavailable
	}
	actual, err := strconv.ParseUint(value.actual, 10, 64)
	if err != nil || actual != key.PositionGeneration || strconv.FormatUint(actual, 10) != value.actual {
		return nil, 0, "", ErrProductionRouteUnavailable
	}
	var account, market, symbol, laneVersion, prospective, actualCampaign, state string
	var entryBlocked int
	if err := queryer.QueryRowContext(ctx, `SELECT account_ref,market,symbol,lane_version,prospective_token,coalesce(actual_position_generation,''),state,entry_blocked FROM position_campaigns WHERE id=? AND lane_id=?`, value.campaignID, value.laneID).
		Scan(&account, &market, &symbol, &laneVersion, &prospective, &actualCampaign, &state, &entryBlocked); err != nil ||
		account != key.AccountRef || market != string(key.Market) || symbol != key.Symbol || prospective != value.prospective ||
		actualCampaign != value.actual || state != "ACTIVE" || entryBlocked != 0 {
		return nil, 0, "", ErrProductionRouteUnavailable
	}
	descriptor, ok := productionRouteDescriptors(key.Market)[value.laneID]
	if !ok || descriptor.LaneVersion != laneVersion {
		return nil, 0, "", ErrProductionRouteUnavailable
	}
	owner := Owner{Key: key, Horizon: descriptor.Horizon, LaneID: value.laneID, LaneVersion: laneVersion,
		CampaignID: value.campaignID, Active: true, Desired: StateOn, Effective: StateOn}
	return []Owner{owner}, revision, digest, nil
}

func validProductionOwnerRow(value productionOwnerRow) bool {
	if !productionRouteIdentity(value.prospective) || !productionRouteIdentity(value.laneID) || !productionRouteIdentity(value.campaignID) ||
		(value.actual != "" && !productionRouteIdentity(value.actual)) || (value.overage != 0 && value.overage != 1) ||
		(value.unknown != 0 && value.unknown != 1) {
		return false
	}
	acquired, acquiredOK := productionRouteTime(value.acquired)
	if !acquiredOK {
		return false
	}
	if value.released != "" {
		released, releasedOK := productionRouteTime(value.released)
		return releasedOK && !released.Before(acquired)
	}
	return true
}

func productionOwnerHistoryDigest(history []productionOwnerRow) string {
	h := sha256.New()
	for _, value := range history {
		for _, item := range []string{value.prospective, value.laneID, value.campaignID, value.actual, value.acquired, value.released,
			strconv.Itoa(value.overage), strconv.Itoa(value.unknown)} {
			writeString(h, item)
		}
	}
	return "strategy-owner-history:v1:sha256:" + hex.EncodeToString(h.Sum(nil))
}

func productionRouteTime(raw string) (time.Time, bool) {
	parsed, err := time.Parse(time.RFC3339Nano, raw)
	return parsed.UTC(), err == nil && !parsed.IsZero() && parsed.Location() == time.UTC && parsed.UTC().Format(time.RFC3339Nano) == raw
}

func productionRouteIdentity(value string) bool {
	return value != "" && value == strings.TrimSpace(value) && len(value) <= 256 && !strings.ContainsAny(value, "\x00\r\n")
}

func productionRouteDigestValid(value string) bool {
	if len(value) != len("sha256:")+sha256.Size*2 || !strings.HasPrefix(value, "sha256:") {
		return false
	}
	decoded, err := hex.DecodeString(strings.TrimPrefix(value, "sha256:"))
	return err == nil && len(decoded) == sha256.Size && value == strings.ToLower(value)
}

func productionRouteDigest(data []byte) string {
	digest := sha256.Sum256(data)
	return "sha256:" + hex.EncodeToString(digest[:])
}
