package exitpolicy

import (
	"crypto/sha256"
	"errors"
	"fmt"
	"regexp"
	"strings"
)

const (
	DefaultPolicyVersion = "1.0.0"
	RatchetPolicyID      = "RATCHET"

	defaultLadderPolicyDigest  = "sha256:e320c21c67852d98aad3ee7b9daa7afa7d5fd619f91f9dfaed93f5857744ea25"
	balancedPolicyDigest       = "sha256:81efea35eb31f02ef9736ceac920a4b895af4bcfad96b0884e41ad4190585a38"
	runnerPolicyDigest         = "sha256:4ff1db694f777c3aa58310566593f26edb537fcfc380115c5ce6ee3da06bce1a"
	hybrid50PolicyDigest       = "sha256:a4e2df8f7971abfdf00beb8be840bd0370bccd5d6ef1ba56a757340013514e4a"
	adoptedRunnerPolicyVersion = "1.0.0-adopted.1"
	adoptedRunnerPolicyDigest  = "sha256:97ae29e33c5c9530b43ecc2c2a830defc16a805904f5e28d7a96960623f9dd3f"
	defaultRatchetPolicyDigest = "sha256:e466b80c94435f8b737eeaf8458d2ffb70ec82d6407f0e6ed93c8dafa2a3f032"
)

var (
	ErrPolicyIdentityConflict = errors.New("exitpolicy: policy identity conflict")
	semanticVersion           = regexp.MustCompile(`^[0-9]+\.[0-9]+\.[0-9]+(?:-[0-9A-Za-z.-]+)?(?:\+[0-9A-Za-z.-]+)?$`)
)

// PolicyIdentity names policy semantics, not merely a registry slot. Digest is
// sha256 over a canonical exact-decimal representation of every decision value.
type PolicyIdentity struct {
	ID      string
	Version string
	Digest  string
}

// LegacyRatchetPolicyIdentity is the only meaning an ID-only pre-a042 RATCHET
// row may have. Keeping this literal separate from the active config makes a
// same-ID table edit fail closed instead of silently reinterpreting live state.
func LegacyRatchetPolicyIdentity() (PolicyIdentity, error) {
	identity := PolicyIdentity{
		ID: RatchetPolicyID, Version: DefaultPolicyVersion, Digest: defaultRatchetPolicyDigest,
	}
	return identity, identity.Validate()
}

// LegacyLadderPolicyIdentity binds an ID-only pre-a042 ladder row to the exact
// semantics that existed when that schema was current. Unknown IDs are not a
// request to consult today's registry: without a stored version/digest their
// meaning is unknowable and evaluation must stop.
func LegacyLadderPolicyIdentity(id string, adopted bool) (PolicyIdentity, error) {
	id = strings.TrimSpace(id)
	if id == "" {
		id = "default_v1"
	}
	version := DefaultPolicyVersion
	digest := ""
	switch id {
	case "default_v1":
		digest = defaultLadderPolicyDigest
	case CommonLadderBalanced:
		digest = balancedPolicyDigest
	case CommonLadderRunner:
		digest = runnerPolicyDigest
		if adopted {
			version, digest = adoptedRunnerPolicyVersion, adoptedRunnerPolicyDigest
		}
	case CommonLadderHybrid50:
		digest = hybrid50PolicyDigest
	default:
		return PolicyIdentity{}, fmt.Errorf("%w: legacy ladder policy %q has no pinned meaning",
			ErrPolicyIdentityConflict, id)
	}
	identity := PolicyIdentity{ID: id, Version: version, Digest: digest}
	return identity, identity.Validate()
}

func (i PolicyIdentity) Validate() error {
	if strings.TrimSpace(i.ID) == "" || !semanticVersion.MatchString(strings.TrimSpace(i.Version)) {
		return fmt.Errorf("%w: policy id and semantic version are required", ErrPolicyIdentityConflict)
	}
	digest := strings.TrimSpace(i.Digest)
	if len(digest) != len("sha256:")+sha256.Size*2 || !strings.HasPrefix(digest, "sha256:") {
		return fmt.Errorf("%w: policy digest must be a sha256 identity", ErrPolicyIdentityConflict)
	}
	return nil
}

func versionOrDefault(version string) string {
	version = strings.TrimSpace(version)
	if version == "" {
		return DefaultPolicyVersion
	}
	return version
}

func digestFields(domain string, fields ...string) string {
	h := sha256.New()
	writeCanonical(h, domain)
	for _, field := range fields {
		writeCanonical(h, field)
	}
	return "sha256:" + fmt.Sprintf("%x", h.Sum(nil))
}

type stringWriter interface{ Write([]byte) (int, error) }

func writeCanonical(w stringWriter, value string) {
	_, _ = fmt.Fprintf(w, "%d:", len(value))
	_, _ = w.Write([]byte(value))
}

func canonicalNumber(field, value string) (string, error) {
	r, err := parseRat(field, value)
	if err != nil {
		return "", err
	}
	return r.RatString(), nil
}

func (p LadderPolicy) Identity() (PolicyIdentity, error) {
	id := strings.TrimSpace(p.PolicyID)
	version := versionOrDefault(p.PolicyVersion)
	if id == "" || !semanticVersion.MatchString(version) {
		return PolicyIdentity{}, fmt.Errorf("%w: ladder policy id/version is invalid", ErrPolicyIdentityConflict)
	}
	fields := []string{id, version, fmt.Sprintf("final=%t", p.FinalTakeFull)}
	for index, rung := range p.Rungs {
		target, err := canonicalNumber("rung target percent", rung.TargetPct)
		if err != nil {
			return PolicyIdentity{}, err
		}
		stop, err := canonicalNumber("rung stop percent", rung.StopPct)
		if err != nil {
			return PolicyIdentity{}, err
		}
		partialRatio, err := parseRatio("rung partial ratio", rung.PartialRatio)
		if err != nil {
			return PolicyIdentity{}, err
		}
		partial := partialRatio.RatString()
		fields = append(fields, fmt.Sprintf("rung=%d", index), target, stop, partial)
	}
	if strings.TrimSpace(p.RunnerTrailPct) == "" {
		fields = append(fields, "runner=")
	} else {
		runner, err := canonicalNumber("runner trail percent", p.RunnerTrailPct)
		if err != nil {
			return PolicyIdentity{}, err
		}
		fields = append(fields, "runner="+runner)
	}
	digest := digestFields("tossos.exitpolicy.ladder.v1", fields...)
	if claimed := strings.TrimSpace(p.PolicyDigest); claimed != "" && claimed != digest {
		return PolicyIdentity{}, fmt.Errorf("%w: %s@%s claims %s, canonical table is %s",
			ErrPolicyIdentityConflict, id, version, claimed, digest)
	}
	identity := PolicyIdentity{ID: id, Version: version, Digest: digest}
	return identity, identity.Validate()
}

func RatchetPolicyIdentity(config RatchetConfig) (PolicyIdentity, error) {
	if err := config.Validate(); err != nil {
		return PolicyIdentity{}, err
	}
	values := []string{
		config.HalfRiskTriggerR, config.BreakevenTriggerR, config.PartialTriggerR,
		config.PartialLockTriggerR, config.ProfitLockTriggerR, config.HalfRiskStopR,
		config.PartialLockStopR, config.ProfitLockStopR, config.PartialRatio,
	}
	fields := []string{RatchetPolicyID, DefaultPolicyVersion}
	for index, value := range values {
		canonical, err := canonicalNumber(fmt.Sprintf("ratchet field %d", index), value)
		if err != nil {
			return PolicyIdentity{}, err
		}
		fields = append(fields, canonical)
	}
	identity := PolicyIdentity{
		ID: RatchetPolicyID, Version: DefaultPolicyVersion,
		Digest: digestFields("tossos.exitpolicy.ratchet.v1", fields...),
	}
	return identity, identity.Validate()
}

// PolicyRegistry rejects a second meaning for an existing (id, version). It
// owns value copies so later caller mutation cannot rewrite registered policy.
type PolicyRegistry struct {
	policies map[string]LadderPolicy
	digests  map[string]string
}

func NewPolicyRegistry(policies ...LadderPolicy) (*PolicyRegistry, error) {
	r := &PolicyRegistry{policies: map[string]LadderPolicy{}, digests: map[string]string{}}
	for _, policy := range policies {
		if err := r.Register(policy); err != nil {
			return nil, err
		}
	}
	return r, nil
}

func (r *PolicyRegistry) Register(policy LadderPolicy) error {
	if r == nil {
		return fmt.Errorf("%w: nil registry", ErrPolicyIdentityConflict)
	}
	if err := policy.Validate(); err != nil {
		return err
	}
	identity, err := policy.Identity()
	if err != nil {
		return err
	}
	key := identity.ID + "@" + identity.Version
	if digest, exists := r.digests[key]; exists && digest != identity.Digest {
		return fmt.Errorf("%w: %s is already %s, refused %s",
			ErrPolicyIdentityConflict, key, digest, identity.Digest)
	}
	policy.PolicyID, policy.PolicyVersion, policy.PolicyDigest = identity.ID, identity.Version, identity.Digest
	policy.Rungs = append([]Rung(nil), policy.Rungs...)
	r.policies[key] = policy
	r.digests[key] = identity.Digest
	return nil
}

func (r *PolicyRegistry) Resolve(identity PolicyIdentity) (LadderPolicy, error) {
	if err := identity.Validate(); err != nil {
		return LadderPolicy{}, err
	}
	key := identity.ID + "@" + identity.Version
	policy, ok := r.policies[key]
	if !ok || r.digests[key] != identity.Digest {
		return LadderPolicy{}, fmt.Errorf("%w: unregistered identity %s", ErrPolicyIdentityConflict, key)
	}
	policy.Rungs = append([]Rung(nil), policy.Rungs...)
	return policy, nil
}
