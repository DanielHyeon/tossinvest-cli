package protectionreadiness

import (
	"crypto/ed25519"
	"crypto/sha256"
	"errors"
	"sort"
	"time"
)

type pinnedKeyInput struct {
	ID           string
	PublicKey    ed25519.PublicKey
	AcceptFrom   time.Time
	PrimaryUntil time.Time
	OverlapUntil time.Time
	RevokedAt    time.Time
}

type pinnedKey struct {
	id           string
	publicKey    ed25519.PublicKey
	acceptFrom   time.Time
	primaryUntil time.Time
	overlapUntil time.Time
	revokedAt    time.Time
}

type pinnedTrustPolicyInput struct {
	Release                string
	AllowedAlgorithms      []string
	MaximumLifetime        time.Duration
	MaximumRotationOverlap time.Duration
	RequiredOwnerUID       uint32
	RequiredMode           uint32
	MaximumFileBytes       int64
	ExpectedPaths          map[Market]string
	Keys                   []pinnedKeyInput
}

type pinnedTrustPolicy struct {
	release                string
	allowedAlgorithms      []string
	maximumLifetime        time.Duration
	maximumRotationOverlap time.Duration
	requiredOwnerUID       uint32
	requiredMode           uint32
	maximumFileBytes       int64
	expectedPaths          map[Market]string
	keys                   []pinnedKey
	seal                   [32]byte
}

func newPinnedTrustPolicy(input pinnedTrustPolicyInput) (pinnedTrustPolicy, error) {
	if input.Release != ReadinessRelease || input.MaximumLifetime <= 0 || input.MaximumRotationOverlap <= 0 || input.RequiredMode != 0o600 || input.MaximumFileBytes <= 0 || len(input.Keys) == 0 {
		return pinnedTrustPolicy{}, errors.New("protectionreadiness: invalid pinned policy")
	}
	algorithms := append([]string(nil), input.AllowedAlgorithms...)
	sort.Strings(algorithms)
	if len(algorithms) != 1 || algorithms[0] != AlgorithmEd25519 {
		return pinnedTrustPolicy{}, errors.New("protectionreadiness: algorithm allowlist must be explicit Ed25519")
	}
	paths := make(map[Market]string, len(input.ExpectedPaths))
	for _, market := range []Market{MarketKR, MarketUS} {
		path := input.ExpectedPaths[market]
		if path == "" {
			return pinnedTrustPolicy{}, errors.New("protectionreadiness: missing market path")
		}
		paths[market] = path
	}
	keys := make([]pinnedKey, 0, len(input.Keys))
	seen := make(map[string]bool, len(input.Keys))
	for _, candidate := range input.Keys {
		if candidate.ID == "" || seen[candidate.ID] || len(candidate.PublicKey) != ed25519.PublicKeySize || candidate.AcceptFrom.IsZero() ||
			!candidate.PrimaryUntil.After(candidate.AcceptFrom) || candidate.OverlapUntil.Before(candidate.PrimaryUntil) ||
			candidate.OverlapUntil.Sub(candidate.PrimaryUntil) > input.MaximumRotationOverlap {
			return pinnedTrustPolicy{}, errors.New("protectionreadiness: invalid pinned key lifecycle")
		}
		seen[candidate.ID] = true
		keys = append(keys, pinnedKey{id: candidate.ID, publicKey: append(ed25519.PublicKey(nil), candidate.PublicKey...), acceptFrom: candidate.AcceptFrom.UTC(), primaryUntil: candidate.PrimaryUntil.UTC(), overlapUntil: candidate.OverlapUntil.UTC(), revokedAt: candidate.RevokedAt.UTC()})
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i].id < keys[j].id })
	policy := pinnedTrustPolicy{release: input.Release, allowedAlgorithms: algorithms, maximumLifetime: input.MaximumLifetime, maximumRotationOverlap: input.MaximumRotationOverlap,
		requiredOwnerUID: input.RequiredOwnerUID, requiredMode: input.RequiredMode, maximumFileBytes: input.MaximumFileBytes, expectedPaths: paths, keys: keys}
	policy.seal = pinnedPolicySeal(policy)
	return policy, nil
}

func pinnedPolicySeal(policy pinnedTrustPolicy) [32]byte {
	hash := sha256.New()
	writeString(hash, policy.release)
	for _, algorithm := range policy.allowedAlgorithms {
		writeString(hash, algorithm)
	}
	writeUint64(hash, uint64(policy.maximumLifetime))
	writeUint64(hash, uint64(policy.maximumRotationOverlap))
	writeUint64(hash, uint64(policy.requiredOwnerUID))
	writeUint64(hash, uint64(policy.requiredMode))
	writeUint64(hash, uint64(policy.maximumFileBytes))
	for _, market := range sortedMarkets(policy.expectedPaths) {
		writeString(hash, string(market))
		writeString(hash, policy.expectedPaths[market])
	}
	for _, key := range policy.keys {
		writeString(hash, key.id)
		writeString(hash, string(key.publicKey))
		writeString(hash, formatTime(key.acceptFrom))
		writeString(hash, formatTime(key.primaryUntil))
		writeString(hash, formatTime(key.overlapUntil))
		writeString(hash, formatTime(key.revokedAt))
	}
	var result [32]byte
	copy(result[:], hash.Sum(nil))
	return result
}

func (policy pinnedTrustPolicy) key(id string) (pinnedKey, bool) {
	for _, key := range policy.keys {
		if key.id == id {
			return key, true
		}
	}
	return pinnedKey{}, false
}
