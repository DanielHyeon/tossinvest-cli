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
