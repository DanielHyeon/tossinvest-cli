package protectionreadiness

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

const (
	ProductionManifestFile = "protection-readiness-manifest.json"
	productionStateFile    = "protection-readiness-state.json"
	productionSchema       = "protection-readiness-deployment/v1"
	productionStateSchema  = "protection-readiness-state/v1"
	maximumProductionFile  = 1 << 20
)

type ProductionConfig struct {
	ConfigDir            string
	AccountID            string
	ProfileID            string
	BuildDigest          string
	ToolDigest           string
	ManifestDigest       string
	Now                  func() time.Time
	SupervisorAssemblies []SupervisorAssembly
}

type SupervisorAssembly struct {
	Market          Market
	ComponentDigest string
	Wired           bool
}

type RuntimeContract struct {
	Market                 Market
	SessionScope           string
	TriggerSource          string
	ReplaceSemantics       string
	BrokerCapabilityDigest string
	ToolDigest             string
}

type productionManifest struct {
	SchemaVersion          string                   `json:"schema_version"`
	MaximumLifetimeSeconds int64                    `json:"maximum_lifetime_seconds"`
	MaximumOverlapSeconds  int64                    `json:"maximum_rotation_overlap_seconds"`
	Keys                   []productionKey          `json:"keys"`
	Markets                []productionMarketConfig `json:"markets"`
}

type productionKey struct {
	KeyID        string `json:"key_id"`
	PublicKey    string `json:"public_key"`
	AcceptFrom   string `json:"accept_from"`
	PrimaryUntil string `json:"primary_until"`
	OverlapUntil string `json:"overlap_until"`
	RevokedAt    string `json:"revoked_at,omitempty"`
}

type productionMarketConfig struct {
	Market           Market           `json:"market"`
	OrderType        string           `json:"order_type"`
	SessionScope     string           `json:"session_scope"`
	QuantityMin      uint64           `json:"quantity_min"`
	QuantityMax      uint64           `json:"quantity_max"`
	TriggerSource    string           `json:"trigger_source"`
	ReplaceSemantics string           `json:"replace_semantics"`
	Broker           brokerCapability `json:"broker"`
	SupervisorDigest string           `json:"supervisor_digest"`
	AttestationFile  string           `json:"attestation_file"`
	EvidenceFile     string           `json:"evidence_file"`
}

type productionState struct {
	SchemaVersion    string                  `json:"schema_version"`
	TrustedTimeFloor string                  `json:"trusted_time_floor,omitempty"`
	Serials          []productionStateSerial `json:"serials"`
}

type productionStateSerial struct {
	AccountID string `json:"account_id"`
	ProfileID string `json:"profile_id"`
	Market    Market `json:"market"`
	Serial    uint64 `json:"serial"`
}

type ProductionProvider struct {
	mu                sync.Mutex
	config            ProductionConfig
	ownerUID          uint32
	manifestPath      string
	manifest          productionManifest
	policy            pinnedTrustPolicy
	marketConfigs     map[Market]productionMarketConfig
	assemblies        map[Market]SupervisorAssembly
	contracts         []RuntimeContract
	configured        bool
	globalRefusal     RefusalCode
	cached            ReadinessSnapshot
	hasCached         bool
	fingerprints      map[Market]string
	stateBootstrapped bool
}

func NewProductionProvider(config ProductionConfig) *ProductionProvider {
	provider := &ProductionProvider{config: config, globalRefusal: RefusalInvalid}
	provider.initialize()
	return provider
}

func (provider *ProductionProvider) RuntimeContracts() []RuntimeContract {
	provider.mu.Lock()
	defer provider.mu.Unlock()
	return append([]RuntimeContract(nil), provider.contracts...)
}

func (provider *ProductionProvider) Current(ctx context.Context) (ReadinessSnapshot, error) {
	if provider == nil {
		return pairedRefusalSnapshot(RefusalProviderUnavailable), nil
	}
	provider.mu.Lock()
	defer provider.mu.Unlock()
	if err := ctx.Err(); err != nil {
		return pairedRefusalSnapshot(RefusalProviderUnavailable), nil
	}
	if !provider.configured {
		if provider.globalRefusal == RefusalMissingEvidence {
			return DefaultSnapshot(), nil
		}
		return pairedRefusalSnapshot(provider.globalRefusal), nil
	}
	if _, data, err := readOwnedFile(provider.manifestPath, provider.ownerUID, 0o400); err != nil || sha256Hex(data) != provider.config.ManifestDigest {
		return pairedRefusalSnapshot(RefusalInvalid), nil
	}
	releaseStateLock, stateLockBootstrapped, markStateLockBootstrapped, err := acquireStateLock(provider.config.ConfigDir, provider.ownerUID)
	if err != nil {
		return pairedRefusalSnapshot(RefusalStateCorrupt), nil
	}
	defer releaseStateLock()
	provider.stateBootstrapped = provider.stateBootstrapped || stateLockBootstrapped
	state, code := provider.loadState()
	if code != RefusalNone {
		return pairedRefusalSnapshot(code), nil
	}
	if provider.stateBootstrapped && !stateLockBootstrapped {
		if err := markStateLockBootstrapped(); err != nil {
			return pairedRefusalSnapshot(RefusalStateCorrupt), nil
		}
	}
	now := provider.config.Now()
	trusted := newTrustedTime(now, "engine-system-clock+durable-floor")
	if !validTrustedTime(trusted) {
		return pairedRefusalSnapshot(RefusalTrustedTimeUnavailable), nil
	}
	if !state.TrustedTimeFloor.IsZero() && now.Before(state.TrustedTimeFloor) {
		return pairedRefusalSnapshot(RefusalTrustedTimeRollback), nil
	}
	if provider.hasCached && !cachedSerialsMatchState(provider.cached, state, provider.config.AccountID, provider.config.ProfileID) {
		provider.hasCached = false
		provider.fingerprints = nil
	}
	inputs := make(map[Market]marketAssessmentInput, 2)
	fingerprints := make(map[Market]string, 2)
	for _, market := range []Market{MarketKR, MarketUS} {
		marketConfig := provider.marketConfigs[market]
		observed, evidenceDigest, fingerprint, ok := provider.readMarketEvidence(marketConfig)
		fingerprints[market] = fingerprint
		if !ok {
			continue
		}
		assembly := provider.assemblies[market]
		binding, err := newSupervisorBinding(supervisorBindingInput{
			AccountID: provider.config.AccountID, ProfileID: provider.config.ProfileID, Market: market,
			BuildDigest: provider.config.BuildDigest, ComponentDigest: assembly.ComponentDigest, Wired: assembly.Wired,
		})
		if err != nil || assembly.ComponentDigest != marketConfig.SupervisorDigest {
			continue
		}
		inputs[market] = marketAssessmentInput{Scope: runtimeScope{
			AccountID: provider.config.AccountID, ProfileID: provider.config.ProfileID, Market: market,
			OrderType: marketConfig.OrderType, SessionScope: marketConfig.SessionScope,
			QuantityMin: marketConfig.QuantityMin, QuantityMax: marketConfig.QuantityMax,
			TriggerSource: marketConfig.TriggerSource, ReplaceSemantics: marketConfig.ReplaceSemantics,
			Broker: marketConfig.Broker, ToolDigest: provider.config.ToolDigest,
			BuildDigest: provider.config.BuildDigest, EvidenceDigest: evidenceDigest,
		}, File: observed, Supervisor: binding}
	}
	changed := inputs
	cachedAtNow := provider.cached
	if provider.hasCached {
		// A peer market whose artifact did not change still crosses key and
		// attestation time boundaries independently. Revalidate the whole cached
		// pair before either verdict is reused; otherwise a KR-only file change at
		// the instant a US key is revoked could restore the stale US WIRED verdict.
		cachedAtNow = provider.revalidateCachedTime(now)
		changed = make(map[Market]marketAssessmentInput, 2)
		for _, market := range []Market{MarketKR, MarketUS} {
			if provider.fingerprints[market] != fingerprints[market] {
				if input, ok := inputs[market]; ok {
					changed[market] = input
				}
			}
		}
		if len(changed) == 0 && provider.fingerprints[MarketKR] == fingerprints[MarketKR] && provider.fingerprints[MarketUS] == fingerprints[MarketUS] {
			if now.After(state.TrustedTimeFloor) {
				state = cloneDurableState(state)
				state.TrustedTimeFloor = now.UTC()
				state.seal = durableStateSeal(state)
				if err := provider.storeState(state); err != nil {
					return pairedRefusalSnapshot(RefusalStateCorrupt), nil
				}
				if err := markStateLockBootstrapped(); err != nil {
					return pairedRefusalSnapshot(RefusalStateCorrupt), nil
				}
			}
			provider.cached = cachedAtNow
			return provider.cached, nil
		}
	}
	result := Assess(assessmentInput{Policy: provider.policy, State: state, Time: trusted, Markets: changed})
	if provider.hasCached {
		for _, market := range []Market{MarketKR, MarketUS} {
			if provider.fingerprints[market] == fingerprints[market] {
				setSnapshotVerdict(&result.Snapshot, market, cachedAtNow.Verdict(market))
			}
		}
		resealSnapshot(&result.Snapshot)
	}
	if result.StateCommitAllowed && result.Mutations > 0 {
		if err := provider.storeState(result.NextState); err != nil {
			return pairedRefusalSnapshot(RefusalStateCorrupt), nil
		}
		if err := markStateLockBootstrapped(); err != nil {
			return pairedRefusalSnapshot(RefusalStateCorrupt), nil
		}
	}
	provider.cached, provider.hasCached, provider.fingerprints = result.Snapshot, true, fingerprints
	return result.Snapshot, nil
}

func (provider *ProductionProvider) initialize() {
	config := &provider.config
	config.ConfigDir = filepath.Clean(strings.TrimSpace(config.ConfigDir))
	config.AccountID = strings.TrimSpace(config.AccountID)
	config.ProfileID = strings.TrimSpace(config.ProfileID)
	config.BuildDigest = strings.TrimSpace(config.BuildDigest)
	config.ToolDigest = strings.TrimSpace(config.ToolDigest)
	config.ManifestDigest = strings.TrimSpace(config.ManifestDigest)
	if config.ManifestDigest == "" {
		provider.globalRefusal = RefusalMissingEvidence
		return
	}
	uid, ok := currentOwnerUID()
	if !ok || !filepath.IsAbs(config.ConfigDir) || config.AccountID == "" || config.ProfileID == "" || config.Now == nil || !validDigest(config.BuildDigest) || !validDigest(config.ToolDigest) || !validDigest(config.ManifestDigest) {
		return
	}
	provider.ownerUID = uid
	provider.manifestPath = filepath.Join(config.ConfigDir, ProductionManifestFile)
	var manifest productionManifest
	_, data, err := readOwnedFile(provider.manifestPath, uid, 0o400)
	if err != nil || sha256Hex(data) != config.ManifestDigest || decodeCanonical(data, &manifest) != nil {
		return
	}
	if manifest.SchemaVersion != productionSchema || manifest.MaximumLifetimeSeconds <= 0 || manifest.MaximumOverlapSeconds <= 0 || len(manifest.Markets) != 2 {
		return
	}
	marketConfigs := make(map[Market]productionMarketConfig, 2)
	contracts := make([]RuntimeContract, 0, 2)
	expectedPaths := make(map[Market]string, 2)
	for _, marketConfig := range manifest.Markets {
		if !validProductionMarket(config.ConfigDir, marketConfig) || marketConfigs[marketConfig.Market].Market != "" {
			return
		}
		marketConfigs[marketConfig.Market] = marketConfig
		expectedPaths[marketConfig.Market] = filepath.Join(config.ConfigDir, marketConfig.AttestationFile)
		contracts = append(contracts, RuntimeContract{Market: marketConfig.Market, SessionScope: marketConfig.SessionScope,
			TriggerSource: marketConfig.TriggerSource, ReplaceSemantics: marketConfig.ReplaceSemantics,
			BrokerCapabilityDigest: brokerCapabilityDigest(marketConfig.Broker), ToolDigest: config.ToolDigest})
	}
	assemblies := make(map[Market]SupervisorAssembly, 2)
	for _, assembly := range config.SupervisorAssemblies {
		if !validMarket(assembly.Market) || !assembly.Wired || !validDigest(assembly.ComponentDigest) || assemblies[assembly.Market].Market != "" {
			return
		}
		assemblies[assembly.Market] = assembly
	}
	if len(assemblies) != 2 {
		return
	}
	keys := make([]pinnedKeyInput, 0, len(manifest.Keys))
	for _, key := range manifest.Keys {
		decoded, err := base64.StdEncoding.Strict().DecodeString(key.PublicKey)
		acceptFrom, ok1 := parseCanonicalTime(key.AcceptFrom)
		primaryUntil, ok2 := parseCanonicalTime(key.PrimaryUntil)
		overlapUntil, ok3 := parseCanonicalTime(key.OverlapUntil)
		var revokedAt time.Time
		ok4 := true
		if key.RevokedAt != "" {
			revokedAt, ok4 = parseCanonicalTime(key.RevokedAt)
		}
		if err != nil || len(decoded) != ed25519.PublicKeySize || !ok1 || !ok2 || !ok3 || !ok4 {
			return
		}
		keys = append(keys, pinnedKeyInput{ID: key.KeyID, PublicKey: ed25519.PublicKey(decoded), AcceptFrom: acceptFrom, PrimaryUntil: primaryUntil, OverlapUntil: overlapUntil, RevokedAt: revokedAt})
	}
	policy, err := newPinnedTrustPolicy(pinnedTrustPolicyInput{Release: ReadinessRelease, AllowedAlgorithms: []string{AlgorithmEd25519},
		MaximumLifetime:        time.Duration(manifest.MaximumLifetimeSeconds) * time.Second,
		MaximumRotationOverlap: time.Duration(manifest.MaximumOverlapSeconds) * time.Second,
		RequiredOwnerUID:       uid, RequiredMode: 0o600, MaximumFileBytes: maximumProductionFile,
		ExpectedPaths: expectedPaths, Keys: keys})
	if err != nil {
		return
	}
	sort.Slice(contracts, func(i, j int) bool { return contracts[i].Market < contracts[j].Market })
	// Publish the paired authority only after the entire manifest, both markets,
	// both assemblies and the trust policy validate. A malformed second market
	// must not leave a single contract that can fail engine construction.
	provider.manifest = manifest
	provider.marketConfigs = marketConfigs
	provider.assemblies = assemblies
	provider.contracts = contracts
	provider.policy = policy
	provider.configured = true
	provider.globalRefusal = RefusalNone
}

func validProductionMarket(configDir string, market productionMarketConfig) bool {
	return validMarket(market.Market) && market.OrderType != "" && market.SessionScope != "" && market.QuantityMin > 0 && market.QuantityMax >= market.QuantityMin &&
		market.TriggerSource != "" && (market.ReplaceSemantics == ReplaceAtomic || market.ReplaceSemantics == ReplaceContinuousCoverage) &&
		validBrokerCapability(market.Broker) && validDigest(market.SupervisorDigest) && safeRelativeArtifact(configDir, market.AttestationFile) && safeRelativeArtifact(configDir, market.EvidenceFile)
}

func safeRelativeArtifact(configDir, name string) bool {
	return name != "" && filepath.Base(name) == name && filepath.Clean(name) == name && filepath.Join(configDir, name) != configDir
}

func (provider *ProductionProvider) readMarketEvidence(config productionMarketConfig) (observedFile, string, string, bool) {
	attestationPath := filepath.Join(provider.config.ConfigDir, config.AttestationFile)
	info, data, err := readOwnedFile(attestationPath, provider.ownerUID, 0o600)
	if err != nil {
		return observedFile{}, "", "missing-attestation", false
	}
	evidencePath := filepath.Join(provider.config.ConfigDir, config.EvidenceFile)
	evidenceInfo, evidence, err := readOwnedFile(evidencePath, provider.ownerUID, 0o600)
	if err != nil {
		return observedFile{}, "", "missing-evidence", false
	}
	owner, _ := fileOwner(info)
	observed, err := newObservedFile(observedFileInput{Bytes: data, ResolvedPath: attestationPath, OwnerUID: owner, Mode: uint32(info.Mode().Perm()), Regular: info.Mode().IsRegular(), Symlink: info.Mode()&os.ModeSymlink != 0})
	fingerprint := sha256Hex([]byte(strings.Join([]string{sha256Hex(data), sha256Hex(evidence), formatTime(info.ModTime()), formatTime(evidenceInfo.ModTime())}, "\x00")))
	return observed, sha256Hex(evidence), fingerprint, err == nil
}

func (provider *ProductionProvider) loadState() (durableState, RefusalCode) {
	path := filepath.Join(provider.config.ConfigDir, productionStateFile)
	_, data, err := readOwnedFile(path, provider.ownerUID, 0o600)
	if errors.Is(err, os.ErrNotExist) {
		if provider.stateBootstrapped {
			return durableState{}, RefusalStateCorrupt
		}
		return newDurableState(), RefusalNone
	}
	if err != nil {
		return durableState{}, RefusalStateCorrupt
	}
	var stored productionState
	if decodeCanonical(data, &stored) != nil || stored.SchemaVersion != productionStateSchema {
		return durableState{}, RefusalStateCorrupt
	}
	var floor time.Time
	if stored.TrustedTimeFloor != "" {
		var ok bool
		floor, ok = parseCanonicalTime(stored.TrustedTimeFloor)
		if !ok {
			return durableState{}, RefusalStateCorrupt
		}
	}
	serials := make(map[serialScope]uint64, len(stored.Serials))
	for _, item := range stored.Serials {
		scope := serialScope{AccountID: item.AccountID, ProfileID: item.ProfileID, Market: item.Market}
		if item.Serial == 0 || scope.AccountID == "" || scope.ProfileID == "" || !validMarket(scope.Market) || serials[scope] != 0 {
			return durableState{}, RefusalStateCorrupt
		}
		serials[scope] = item.Serial
	}
	provider.stateBootstrapped = true
	return newDurableStateWith(floor, serials), RefusalNone
}

func (provider *ProductionProvider) storeState(state durableState) error {
	if !validDurableState(state) {
		return errors.New("invalid durable readiness state")
	}
	stored := productionState{SchemaVersion: productionStateSchema, TrustedTimeFloor: formatTime(state.TrustedTimeFloor)}
	for scope, serial := range state.Serials {
		stored.Serials = append(stored.Serials, productionStateSerial{AccountID: scope.AccountID, ProfileID: scope.ProfileID, Market: scope.Market, Serial: serial})
	}
	sort.Slice(stored.Serials, func(i, j int) bool {
		left, right := stored.Serials[i], stored.Serials[j]
		if left.AccountID != right.AccountID {
			return left.AccountID < right.AccountID
		}
		if left.ProfileID != right.ProfileID {
			return left.ProfileID < right.ProfileID
		}
		return left.Market < right.Market
	})
	data, err := json.Marshal(stored)
	if err != nil {
		return err
	}
	path := filepath.Join(provider.config.ConfigDir, productionStateFile)
	temporary, err := os.CreateTemp(provider.config.ConfigDir, ".protection-readiness-state-*")
	if err != nil {
		return err
	}
	temporaryPath := temporary.Name()
	ok := false
	defer func() {
		if !ok {
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
	if err := os.Rename(temporaryPath, path); err != nil {
		return err
	}
	directory, err := os.Open(provider.config.ConfigDir)
	if err != nil {
		return err
	}
	err = directory.Sync()
	_ = directory.Close()
	if err != nil {
		return err
	}
	ok = true
	provider.stateBootstrapped = true
	return nil
}

func (provider *ProductionProvider) revalidateCachedTime(now time.Time) ReadinessSnapshot {
	snapshot := provider.cached
	for _, market := range []Market{MarketKR, MarketUS} {
		// Read every verdict from the original sealed snapshot. Mutating the KR
		// copy invalidates snapshot's aggregate seal until the final reseal; asking
		// that partially mutated copy for US would therefore return STATE_CORRUPT
		// and accidentally skip the peer's time-boundary check.
		verdict := provider.cached.Verdict(market)
		if verdict.State != Wired || verdict.Code != RefusalNone {
			continue
		}
		key, ok := provider.policy.key(verdict.Provenance.KeyID)
		code := RefusalNone
		switch {
		case !ok:
			code = RefusalUnknownKey
		case !key.revokedAt.IsZero() && !now.Before(key.revokedAt):
			code = RefusalRevokedKey
		case now.Before(key.acceptFrom) || !now.Before(key.overlapUntil):
			code = RefusalRotationWindow
		case now.Before(verdict.Provenance.IssuedAt):
			code = RefusalIssuedInFuture
		case !now.Before(verdict.Provenance.ExpiresAt):
			code = RefusalExpired
		}
		if code != RefusalNone {
			verdict.State, verdict.Code = Unwired, code
			setSnapshotVerdict(&snapshot, market, verdict)
		}
	}
	resealSnapshot(&snapshot)
	return snapshot
}

func cachedSerialsMatchState(snapshot ReadinessSnapshot, state durableState, accountID, profileID string) bool {
	for _, market := range []Market{MarketKR, MarketUS} {
		provenance := snapshot.Verdict(market).Provenance
		if provenance.Serial == 0 {
			continue
		}
		if state.Serials[serialScope{AccountID: accountID, ProfileID: profileID, Market: market}] != provenance.Serial {
			return false
		}
	}
	return true
}

func decodeCanonical(data []byte, target any) error {
	if len(data) == 0 || len(data) > maximumProductionFile {
		return errors.New("invalid production document size")
	}
	if duplicate, err := containsDuplicateJSONKey(data); err != nil || duplicate {
		return errors.New("invalid or duplicate production JSON")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return err
	}
	if err := ensureJSONEOF(decoder); err != nil {
		return err
	}
	canonical, err := json.Marshal(target)
	if err != nil || !bytes.Equal(canonical, data) {
		return errors.New("production JSON is not canonical")
	}
	return nil
}

func sha256Hex(data []byte) string {
	digest := sha256.Sum256(data)
	return fmt.Sprintf("%x", digest[:])
}

func pairedRefusalSnapshot(code RefusalCode) ReadinessSnapshot {
	snapshot := DefaultSnapshot()
	snapshot.kr.Code, snapshot.us.Code = code, code
	snapshot.krSeal = marketVerdictSeal(snapshot.release, snapshot.kr)
	snapshot.usSeal = marketVerdictSeal(snapshot.release, snapshot.us)
	snapshot.seal = readinessSnapshotSeal(snapshot)
	return snapshot
}

func setSnapshotVerdict(snapshot *ReadinessSnapshot, market Market, verdict Verdict) {
	if market == MarketKR {
		snapshot.kr = verdict
	} else if market == MarketUS {
		snapshot.us = verdict
	}
}

func resealSnapshot(snapshot *ReadinessSnapshot) {
	snapshot.krSeal = marketVerdictSeal(snapshot.release, snapshot.kr)
	snapshot.usSeal = marketVerdictSeal(snapshot.release, snapshot.us)
	snapshot.seal = readinessSnapshotSeal(*snapshot)
}
