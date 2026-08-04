package officialfx

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

const (
	RiskPolicyManifestFile       = "fx-risk-policy-manifest.json"
	riskPolicyStateFile          = "fx-risk-policy-state.json"
	riskPolicyStateLockFile      = ".fx-risk-policy-state.lock"
	riskPolicySchema             = "fx-risk-policy/v1"
	riskPolicyStateSchema        = "fx-risk-policy-state/v1"
	riskPolicySignatureAlgorithm = "Ed25519"
	maximumProductionFile        = 1 << 20
)

var (
	ErrProductionAuthorityUnavailable = errors.New("officialfx: production authority unavailable")
	ErrProductionIdentityUnavailable  = errors.New("officialfx: production identity unavailable")
	ErrProductionPolicyUnavailable    = errors.New("officialfx: production FX policy unavailable")
	ErrProductionPolicyInvalid        = errors.New("officialfx: production FX policy invalid")
	ErrProductionPolicyStale          = errors.New("officialfx: production FX policy stale")
	ErrProductionPolicyScope          = errors.New("officialfx: production FX policy scope mismatch")
	ErrProductionAuthorityRollback    = errors.New("officialfx: production FX authority rollback")
	ErrProductionAuthorityState       = errors.New("officialfx: production FX authority state corrupt")
)

// ProductionAuthorityConfig contains only immutable scope and trust pins. It
// deliberately has no rate, haircut, evidence-digest or evidence-freshness
// field from which a caller could mint monetary authority.
type ProductionAuthorityConfig struct {
	ConfigDir       string
	AccountID       string
	AccountCurrency string
	ManifestDigest  string
	TrustedKeyID    string
	TrustedKey      ed25519.PublicKey
	Now             func() time.Time
}

type riskPolicyManifestBody struct {
	SchemaVersion      string `json:"schema_version"`
	AccountID          string `json:"account_id"`
	AccountCurrency    string `json:"account_currency"`
	Market             string `json:"market"`
	QuoteCurrency      string `json:"quote_currency"`
	Generation         uint64 `json:"generation"`
	PolicyID           string `json:"policy_id"`
	PolicyVersion      string `json:"policy_version"`
	Multiplier         string `json:"multiplier"`
	ObservedAt         string `json:"observed_at"`
	FreshUntil         string `json:"fresh_until"`
	Approver           string `json:"approver"`
	KeyID              string `json:"key_id"`
	SignatureAlgorithm string `json:"signature_algorithm"`
}

type riskPolicyManifest struct {
	riskPolicyManifestBody
	Signature string `json:"signature"`
}

type riskPolicyState struct {
	SchemaVersion    string `json:"schema_version"`
	TrustedTimeFloor string `json:"trusted_time_floor"`
	Generation       uint64 `json:"generation"`
	ManifestDigest   string `json:"manifest_digest"`
}

type productionOfficialReader interface {
	reverifyAccount(context.Context, string) error
	readOfficial(context.Context, string, string, HaircutPolicy) (Evidence, error)
}

type productionOfficialClient interface {
	AuthoritativeExchangeRateReader
	VerifyAuthoritativeAccountIdentity(context.Context, string) error
}

type officialProductionReader struct{ client productionOfficialClient }

func (reader officialProductionReader) reverifyAccount(ctx context.Context, accountID string) error {
	if reader.client == nil {
		return ErrProductionAuthorityUnavailable
	}
	return reader.client.VerifyAuthoritativeAccountIdentity(ctx, accountID)
}

func (reader officialProductionReader) readOfficial(ctx context.Context, quoteCurrency, accountCurrency string, policy HaircutPolicy) (Evidence, error) {
	return ReadOfficial(ctx, reader.client, quoteCurrency, accountCurrency, policy)
}

// ProductionAuthorityService is the sole production mint for paired KR
// identity and US official quote-to-account-base evidence.
type ProductionAuthorityService struct {
	config ProductionAuthorityConfig
	reader productionOfficialReader
}

func NewProductionAuthorityService(config ProductionAuthorityConfig, client productionOfficialClient) *ProductionAuthorityService {
	return newProductionAuthorityServiceWithReader(config, officialProductionReader{client: client})
}

func newProductionAuthorityServiceWithReader(config ProductionAuthorityConfig, reader productionOfficialReader) *ProductionAuthorityService {
	config.ConfigDir = filepath.Clean(strings.TrimSpace(config.ConfigDir))
	config.AccountID = strings.TrimSpace(config.AccountID)
	config.AccountCurrency = strings.TrimSpace(config.AccountCurrency)
	config.ManifestDigest = strings.TrimSpace(config.ManifestDigest)
	config.TrustedKeyID = strings.TrimSpace(config.TrustedKeyID)
	config.TrustedKey = append(ed25519.PublicKey(nil), config.TrustedKey...)
	return &ProductionAuthorityService{config: config, reader: reader}
}

// Collection preserves market-local outcomes. Its zero value cannot release
// authority, and callers receive only opaque Evidence values.
type Collection struct {
	kr, us       Evidence
	krErr, usErr error
	krSet, usSet bool
}

func (collection Collection) KR() (Evidence, error) {
	if !collection.krSet {
		return Evidence{}, ErrProductionAuthorityUnavailable
	}
	return collection.kr, collection.krErr
}

func (collection Collection) US() (Evidence, error) {
	if !collection.usSet {
		return Evidence{}, ErrProductionAuthorityUnavailable
	}
	return collection.us, collection.usErr
}

// CollectKR obtains only KR identity authority. It never starts or waits for a
// US FX read, so a slow or unavailable US endpoint cannot consume the KR
// worker's evaluation budget.
func (service *ProductionAuthorityService) CollectKR(ctx context.Context) (Evidence, error) {
	now, err := service.collectionTime(ctx)
	if err != nil {
		return Evidence{}, err
	}
	return service.collectKR(ctx, now)
}

// CollectUS obtains only US quote-to-account-base authority. It never starts
// or waits for a KR account-identity read. Paired delivery does not create a
// combined availability or latency domain.
func (service *ProductionAuthorityService) CollectUS(ctx context.Context) (Evidence, error) {
	now, err := service.collectionTime(ctx)
	if err != nil {
		return Evidence{}, err
	}
	return service.collectUS(ctx, now)
}

func (service *ProductionAuthorityService) collectionTime(ctx context.Context) (time.Time, error) {
	if service == nil || ctx == nil || service.config.Now == nil || service.reader == nil {
		return time.Time{}, ErrProductionAuthorityUnavailable
	}
	if err := ctx.Err(); err != nil {
		return time.Time{}, fmt.Errorf("%w: %v", ErrProductionAuthorityUnavailable, err)
	}
	now := service.config.Now()
	if now.IsZero() {
		return time.Time{}, ErrProductionAuthorityUnavailable
	}
	return now.UTC(), nil
}

// Collect attempts both market authorities with one frozen time observation.
// One market's refusal is never used to cancel the peer attempt.
func (service *ProductionAuthorityService) Collect(ctx context.Context) Collection {
	if service == nil || ctx == nil || service.config.Now == nil || service.reader == nil {
		return Collection{krErr: ErrProductionAuthorityUnavailable, usErr: ErrProductionAuthorityUnavailable, krSet: true, usSet: true}
	}
	if err := ctx.Err(); err != nil {
		return Collection{krErr: fmt.Errorf("%w: %v", ErrProductionAuthorityUnavailable, err),
			usErr: fmt.Errorf("%w: %v", ErrProductionAuthorityUnavailable, err), krSet: true, usSet: true}
	}
	now := service.config.Now()
	if now.IsZero() {
		return Collection{krErr: ErrProductionAuthorityUnavailable, usErr: ErrProductionAuthorityUnavailable, krSet: true, usSet: true}
	}
	now = now.UTC()
	type outcome struct {
		market string
		value  Evidence
		err    error
	}
	results := make(chan outcome, 2)
	go func() {
		value, err := service.collectKR(ctx, now)
		results <- outcome{market: "KR", value: value, err: err}
	}()
	go func() {
		value, err := service.collectUS(ctx, now)
		results <- outcome{market: "US", value: value, err: err}
	}()
	collection := Collection{}
	for index := 0; index < 2; index++ {
		result := <-results
		if result.market == "KR" {
			collection.kr, collection.krErr, collection.krSet = result.value, result.err, true
		} else {
			collection.us, collection.usErr, collection.usSet = result.value, result.err, true
		}
	}
	return collection
}

func (service *ProductionAuthorityService) collectKR(ctx context.Context, now time.Time) (Evidence, error) {
	if !canonicalProductionAccountID(service.config.AccountID) || !canonicalCurrency(service.config.AccountCurrency) {
		return Evidence{}, fmt.Errorf("%w: invalid scope", ErrProductionIdentityUnavailable)
	}
	if err := service.reader.reverifyAccount(ctx, service.config.AccountID); err != nil {
		return Evidence{}, fmt.Errorf("%w: %v", ErrProductionIdentityUnavailable, err)
	}
	snapshotID := "official-account-" + strings.TrimPrefix(sha256Identity(service.config.AccountID+"\x00"+now.Format(time.RFC3339Nano)), "sha256:")
	digest := sha256Identity(strings.Join([]string{"official-account-identity/v1", service.config.AccountID,
		service.config.AccountCurrency, now.Format(time.RFC3339Nano)}, "\x00"))
	snapshot, err := newIdentitySnapshot(service.config.AccountCurrency, snapshotID, digest, now, now.Add(maxIdentityWindow))
	if err != nil {
		return Evidence{}, fmt.Errorf("%w: %v", ErrProductionIdentityUnavailable, err)
	}
	evidence, err := Identity(snapshot)
	if err != nil {
		return Evidence{}, fmt.Errorf("%w: %v", ErrProductionIdentityUnavailable, err)
	}
	return evidence, nil
}

func (service *ProductionAuthorityService) collectUS(ctx context.Context, now time.Time) (Evidence, error) {
	config := service.config
	if !filepath.IsAbs(config.ConfigDir) || !canonicalProductionAccountID(config.AccountID) || !canonicalCurrency(config.AccountCurrency) ||
		!canonicalDigest(config.ManifestDigest) || !boundedIdentity(config.TrustedKeyID) || len(config.TrustedKey) != ed25519.PublicKeySize {
		return Evidence{}, ErrProductionPolicyUnavailable
	}
	owner, ok := productionOwnerUID()
	if !ok {
		return Evidence{}, ErrProductionPolicyUnavailable
	}
	release, anchored, markAnchored, err := acquireProductionStateLock(config.ConfigDir, owner)
	if err != nil {
		return Evidence{}, fmt.Errorf("%w: lock: %v", ErrProductionAuthorityState, err)
	}
	defer release()
	manifestPath := filepath.Join(config.ConfigDir, RiskPolicyManifestFile)
	_, data, err := readProductionFile(manifestPath, owner, 0o400)
	if err != nil || sha256Identity(string(data)) != config.ManifestDigest {
		return Evidence{}, ErrProductionPolicyUnavailable
	}
	manifest, err := decodeRiskPolicyManifest(data)
	if err != nil {
		return Evidence{}, fmt.Errorf("%w: %v", ErrProductionPolicyInvalid, err)
	}
	policy, observedAt, freshUntil, err := service.verifyRiskPolicyManifest(manifest, now)
	if err != nil {
		return Evidence{}, err
	}
	state, err := loadRiskPolicyState(config.ConfigDir, owner, anchored)
	if err != nil {
		return Evidence{}, err
	}
	if !stateTimeAndGenerationCurrent(state, now, manifest.Generation, config.ManifestDigest) {
		return Evidence{}, ErrProductionAuthorityRollback
	}
	evidence, err := service.reader.readOfficial(ctx, manifest.QuoteCurrency, manifest.AccountCurrency, policy)
	if err != nil {
		return Evidence{}, fmt.Errorf("%w: %v", ErrProductionPolicyUnavailable, err)
	}
	if _, err := evidence.EvidenceAt(now, manifest.QuoteCurrency, manifest.AccountCurrency); err != nil {
		return Evidence{}, fmt.Errorf("%w: %v", ErrProductionPolicyUnavailable, err)
	}
	_ = observedAt
	_ = freshUntil
	next := riskPolicyState{SchemaVersion: riskPolicyStateSchema, TrustedTimeFloor: now.Format(time.RFC3339Nano),
		Generation: manifest.Generation, ManifestDigest: config.ManifestDigest}
	if err := storeRiskPolicyState(config.ConfigDir, next); err != nil {
		return Evidence{}, fmt.Errorf("%w: store", ErrProductionAuthorityState)
	}
	if err := markAnchored(); err != nil {
		return Evidence{}, fmt.Errorf("%w: anchor", ErrProductionAuthorityState)
	}
	return evidence, nil
}

func (service *ProductionAuthorityService) verifyRiskPolicyManifest(manifest riskPolicyManifest, now time.Time) (HaircutPolicy, time.Time, time.Time, error) {
	config := service.config
	if manifest.SchemaVersion != riskPolicySchema || manifest.AccountID != config.AccountID || manifest.AccountCurrency != config.AccountCurrency ||
		manifest.Market != "US" || manifest.QuoteCurrency != "USD" || manifest.QuoteCurrency == manifest.AccountCurrency {
		return HaircutPolicy{}, time.Time{}, time.Time{}, ErrProductionPolicyScope
	}
	if manifest.Generation == 0 || !boundedIdentity(manifest.PolicyID) || !boundedIdentity(manifest.PolicyVersion) ||
		!boundedIdentity(manifest.Approver) || manifest.KeyID != config.TrustedKeyID || manifest.SignatureAlgorithm != riskPolicySignatureAlgorithm {
		return HaircutPolicy{}, time.Time{}, time.Time{}, ErrProductionPolicyInvalid
	}
	observedAt, observedOK := parseProductionTime(manifest.ObservedAt)
	freshUntil, freshOK := parseProductionTime(manifest.FreshUntil)
	if !observedOK || !freshOK || observedAt.After(now) || now.After(freshUntil) || observedAt.After(freshUntil) {
		return HaircutPolicy{}, time.Time{}, time.Time{}, ErrProductionPolicyStale
	}
	body, err := json.Marshal(manifest.riskPolicyManifestBody)
	if err != nil {
		return HaircutPolicy{}, time.Time{}, time.Time{}, ErrProductionPolicyInvalid
	}
	signature, err := base64.StdEncoding.Strict().DecodeString(manifest.Signature)
	if err != nil || len(signature) != ed25519.SignatureSize || !ed25519.Verify(config.TrustedKey, body, signature) {
		return HaircutPolicy{}, time.Time{}, time.Time{}, ErrProductionPolicyInvalid
	}
	policy, err := newHaircutPolicy(manifest.PolicyID, manifest.PolicyVersion, manifest.Multiplier, observedAt, freshUntil)
	if err != nil {
		return HaircutPolicy{}, time.Time{}, time.Time{}, ErrProductionPolicyInvalid
	}
	return policy, observedAt, freshUntil, nil
}

func decodeRiskPolicyManifest(data []byte) (riskPolicyManifest, error) {
	if len(data) == 0 || len(data) > maximumProductionFile {
		return riskPolicyManifest{}, errors.New("invalid document size")
	}
	duplicate, err := containsDuplicateProductionJSONKey(data)
	if err != nil || duplicate {
		return riskPolicyManifest{}, errors.New("invalid or duplicate JSON")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	var manifest riskPolicyManifest
	if err := decoder.Decode(&manifest); err != nil {
		return riskPolicyManifest{}, err
	}
	if err := ensureProductionJSONEOF(decoder); err != nil {
		return riskPolicyManifest{}, err
	}
	canonical, err := json.Marshal(manifest)
	if err != nil || !bytes.Equal(canonical, data) {
		return riskPolicyManifest{}, errors.New("non-canonical JSON")
	}
	return manifest, nil
}

func containsDuplicateProductionJSONKey(data []byte) (bool, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	var walk func() (bool, error)
	walk = func() (bool, error) {
		token, err := decoder.Token()
		if err != nil {
			return false, err
		}
		delimiter, ok := token.(json.Delim)
		if !ok {
			return false, nil
		}
		switch delimiter {
		case '{':
			seen := make(map[string]bool)
			for decoder.More() {
				keyToken, err := decoder.Token()
				if err != nil {
					return false, err
				}
				key, ok := keyToken.(string)
				if !ok || seen[key] {
					return true, nil
				}
				seen[key] = true
				duplicate, err := walk()
				if duplicate || err != nil {
					return duplicate, err
				}
			}
			_, err = decoder.Token()
			return false, err
		case '[':
			for decoder.More() {
				duplicate, err := walk()
				if duplicate || err != nil {
					return duplicate, err
				}
			}
			_, err = decoder.Token()
			return false, err
		default:
			return false, errors.New("unexpected delimiter")
		}
	}
	duplicate, err := walk()
	if err != nil || duplicate {
		return duplicate, err
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		if err == nil {
			return false, errors.New("trailing JSON")
		}
		return false, err
	}
	return false, nil
}

func ensureProductionJSONEOF(decoder *json.Decoder) error {
	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		if err == nil {
			return errors.New("trailing JSON value")
		}
		return err
	}
	return nil
}

func parseProductionTime(value string) (time.Time, bool) {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	return parsed, err == nil && parsed.Location() == time.UTC && parsed.Format(time.RFC3339Nano) == value
}

func canonicalProductionAccountID(value string) bool {
	if value == "" || value[0] == '0' || len(value) > 19 {
		return false
	}
	for _, character := range value {
		if character < '0' || character > '9' {
			return false
		}
	}
	parsed, err := strconv.ParseUint(value, 10, 63)
	return err == nil && parsed > 0
}

func stateTimeAndGenerationCurrent(state riskPolicyState, now time.Time, generation uint64, digest string) bool {
	if state.Generation == 0 {
		return state.TrustedTimeFloor == "" && state.ManifestDigest == ""
	}
	floor, ok := parseProductionTime(state.TrustedTimeFloor)
	if !ok || now.Before(floor) || generation < state.Generation {
		return false
	}
	return generation != state.Generation || state.ManifestDigest == digest
}

func loadRiskPolicyState(configDir string, owner uint32, anchored bool) (riskPolicyState, error) {
	_, data, err := readProductionFile(filepath.Join(configDir, riskPolicyStateFile), owner, 0o600)
	if errors.Is(err, os.ErrNotExist) && !anchored {
		return riskPolicyState{SchemaVersion: riskPolicyStateSchema}, nil
	}
	if err != nil {
		return riskPolicyState{}, fmt.Errorf("%w: read", ErrProductionAuthorityState)
	}
	var state riskPolicyState
	if err := decodeCanonicalProductionJSON(data, &state); err != nil {
		return riskPolicyState{}, ErrProductionAuthorityState
	}
	emptyState := state.Generation == 0 && state.TrustedTimeFloor == "" && state.ManifestDigest == ""
	committedState := state.Generation > 0 && state.TrustedTimeFloor != "" && canonicalDigest(state.ManifestDigest)
	if state.SchemaVersion != riskPolicyStateSchema || (!emptyState && !committedState) {
		return riskPolicyState{}, ErrProductionAuthorityState
	}
	return state, nil
}

func decodeCanonicalProductionJSON(data []byte, target any) error {
	if len(data) == 0 || len(data) > maximumProductionFile {
		return errors.New("invalid document size")
	}
	duplicate, err := containsDuplicateProductionJSONKey(data)
	if err != nil || duplicate {
		return errors.New("invalid JSON")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := ensureProductionJSONEOF(decoder); err != nil {
		return err
	}
	canonical, err := json.Marshal(target)
	if err != nil || !bytes.Equal(canonical, data) {
		return errors.New("non-canonical JSON")
	}
	return nil
}

func storeRiskPolicyState(configDir string, state riskPolicyState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	temporary, err := os.CreateTemp(configDir, ".fx-risk-policy-state-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	committed := false
	defer func() {
		if !committed {
			_ = os.Remove(temporaryPath)
		}
	}()
	if err := temporary.Chmod(0o600); err != nil {
		_ = temporary.Close()
		return err
	}
	if _, err := temporary.Write(data); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Sync(); err != nil {
		_ = temporary.Close()
		return err
	}
	if err := temporary.Close(); err != nil {
		return err
	}
	if err := os.Rename(temporaryPath, filepath.Join(configDir, riskPolicyStateFile)); err != nil {
		return err
	}
	directory, err := os.Open(configDir)
	if err != nil {
		return err
	}
	err = directory.Sync()
	_ = directory.Close()
	if err != nil {
		return err
	}
	committed = true
	return nil
}
