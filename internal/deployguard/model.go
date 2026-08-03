// Package deployguard validates immutable dormant deployment evidence and
// produces plain, one-at-a-time replacement/rollback action values. It has no
// Docker, process, engine, broker, journal writer, configuration writer or
// protection mutation capability.
package deployguard

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"sort"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"
)

const PreimageSchemaVersion = "tossos.deploy-preimage/v1"

const MaxServiceTimeout = 5 * time.Minute

const MaxBaselineHealthAge = 5 * time.Minute

var frozenServiceOrder = [...]string{"httpapi", "tossos"}

func FrozenServiceOrder() []string { return append([]string(nil), frozenServiceOrder[:]...) }

type Digest string

func DigestBytes(value []byte) Digest {
	sum := sha256.Sum256(value)
	return Digest("sha256:" + hex.EncodeToString(sum[:]))
}

func ImmutableImageDigest(reference string) (Digest, error) {
	if reference != strings.TrimSpace(reference) || strings.Count(reference, "@") != 1 {
		return "", errors.New("deploy guard: image reference must contain one immutable digest")
	}
	parts := strings.SplitN(reference, "@", 2)
	digest := Digest(parts[1])
	if !validIdentity(parts[0]) || !validDigest(digest) {
		return "", errors.New("deploy guard: image reference is mutable or invalid")
	}
	return digest, nil
}

// ValidateRenderedTargetImages checks the image references extracted from a
// read-only `docker compose config` result against the sealed target set.
func ValidateRenderedTargetImages(plan Plan, images map[string]string) error {
	if err := validatePlan(plan); err != nil {
		return err
	}
	if len(images) != len(frozenServiceOrder) {
		return errors.New("deploy guard: rendered Compose service set is incomplete")
	}
	for _, service := range plan.preimage.Services {
		reference, ok := images[service.Name]
		if !ok {
			return errors.New("deploy guard: rendered Compose lacks a frozen service")
		}
		digest, err := ImmutableImageDigest(reference)
		if err != nil || digest != service.TargetImageDigest {
			return fmt.Errorf("deploy guard: rendered image for %s differs from the sealed target", service.Name)
		}
	}
	return nil
}

type State string

const (
	StateUnknown    State = "UNKNOWN"
	StateOff        State = "OFF"
	StateOn         State = "ON"
	StateUnapproved State = "UNAPPROVED"
	StateApproved   State = "APPROVED"
)

type Market string

const (
	MarketKR Market = "KR"
	MarketUS Market = "US"
)

type Refusal string

const RefusalNotConfigured Refusal = "NOT_CONFIGURED"

type MountType string

const (
	MountBind   MountType = "bind"
	MountVolume MountType = "volume"
)

type MountMode string

const (
	MountReadOnly  MountMode = "ro"
	MountReadWrite MountMode = "rw"
)

type MarketState struct {
	LaneDesired         State   `json:"laneDesired"`
	LaneEffective       State   `json:"laneEffective"`
	ActivationDesired   State   `json:"activationDesired"`
	ActivationEffective State   `json:"activationEffective"`
	EntryEffective      State   `json:"entryEffective"`
	Refusal             Refusal `json:"refusal"`
}

type StateEvidence struct {
	RenderedComposeDigest Digest                 `json:"renderedComposeDigest"`
	ConfigDigest          Digest                 `json:"configDigest"`
	ActivationDigest      Digest                 `json:"activationDigest"`
	LaneDigest            Digest                 `json:"laneDigest"`
	AutostartDigest       Digest                 `json:"autostartDigest"`
	AutomationDigest      Digest                 `json:"automationDigest"`
	LiveApprovalDigest    Digest                 `json:"liveApprovalDigest"`
	ProtectionDigest      Digest                 `json:"protectionDigest"`
	JournalDigest         Digest                 `json:"journalDigest"`
	Autostart             State                  `json:"autostart"`
	Automation            State                  `json:"automation"`
	LiveApproval          State                  `json:"liveApproval"`
	Markets               map[Market]MarketState `json:"markets"`
}

type MountIdentity struct {
	Type           MountType `json:"type"`
	Source         string    `json:"source"`
	Target         string    `json:"target"`
	Mode           MountMode `json:"mode"`
	IdentityDigest Digest    `json:"identityDigest"`
}

type VersionRange struct {
	Min uint64 `json:"min"`
	Max uint64 `json:"max"`
}

func (r VersionRange) Contains(version uint64) bool {
	return version != 0 && r.Min != 0 && r.Min <= version && version <= r.Max
}

type SchemaCompatibility struct {
	Readable VersionRange `json:"readable"`
	Writable VersionRange `json:"writable"`
}

type HealthEvidence struct {
	Healthy        bool      `json:"healthy"`
	ObservedAt     time.Time `json:"observedAt"`
	EvidenceDigest Digest    `json:"evidenceDigest"`
}

type ServicePreimage struct {
	Name                     string              `json:"name"`
	CurrentImageDigest       Digest              `json:"currentImageDigest"`
	TargetImageDigest        Digest              `json:"targetImageDigest"`
	Timeout                  time.Duration       `json:"timeout"`
	EnvironmentKeys          []string            `json:"environmentKeys"`
	Mounts                   []MountIdentity     `json:"mounts"`
	CurrentSchemaVersion     uint64              `json:"currentSchemaVersion"`
	PostReplaceSchemaVersion uint64              `json:"postReplaceSchemaVersion"`
	TargetSchema             SchemaCompatibility `json:"targetSchema"`
	RollbackSchema           SchemaCompatibility `json:"rollbackSchema"`
	BaselineHealth           HealthEvidence      `json:"baselineHealth"`
}

type Preimage struct {
	SchemaVersion string            `json:"schemaVersion"`
	CapturedAt    time.Time         `json:"capturedAt"`
	State         StateEvidence     `json:"state"`
	Services      []ServicePreimage `json:"services"`
}

type Plan struct {
	preimage Preimage
	digest   Digest
}

func (p Plan) Digest() Digest { return p.digest }

func Freeze(input Preimage) (Plan, error) {
	preimage := clonePreimage(input)
	if err := validatePreimage(preimage); err != nil {
		return Plan{}, err
	}
	body, err := json.Marshal(preimage)
	if err != nil {
		return Plan{}, fmt.Errorf("deploy guard: encode preimage: %w", err)
	}
	return Plan{preimage: preimage, digest: DigestBytes(body)}, nil
}

func validatePreimage(preimage Preimage) error {
	if preimage.SchemaVersion != PreimageSchemaVersion || preimage.CapturedAt.IsZero() ||
		preimage.CapturedAt.Location() != time.UTC || len(preimage.Services) != len(frozenServiceOrder) {
		return errors.New("deploy guard: invalid preimage envelope")
	}
	if err := validateState(preimage.State); err != nil {
		return err
	}
	seen := make(map[string]struct{}, len(preimage.Services))
	for index, service := range preimage.Services {
		if service.Name != frozenServiceOrder[index] {
			return errors.New("deploy guard: service order differs from the frozen manifest")
		}
		if err := validateService(service, preimage.CapturedAt); err != nil {
			return fmt.Errorf("deploy guard: service %d: %w", index, err)
		}
		if _, duplicate := seen[service.Name]; duplicate {
			return errors.New("deploy guard: duplicate service")
		}
		seen[service.Name] = struct{}{}
	}
	return nil
}

func validateState(state StateEvidence) error {
	for _, digest := range []Digest{state.RenderedComposeDigest, state.ConfigDigest, state.ActivationDigest,
		state.LaneDigest, state.AutostartDigest, state.AutomationDigest, state.LiveApprovalDigest,
		state.ProtectionDigest, state.JournalDigest} {
		if !validDigest(digest) {
			return errors.New("deploy guard: incomplete state digest preimage")
		}
	}
	if state.Autostart != StateOff || state.Automation != StateOff || state.LiveApproval != StateUnapproved || len(state.Markets) != 2 {
		return errors.New("deploy guard: deployment baseline is not dormant")
	}
	for _, market := range []Market{MarketKR, MarketUS} {
		value, ok := state.Markets[market]
		if !ok || value.LaneDesired != StateOff || value.LaneEffective != StateOff ||
			value.ActivationDesired != StateOff || value.ActivationEffective != StateOff ||
			value.EntryEffective != StateOff || value.Refusal != RefusalNotConfigured {
			return fmt.Errorf("deploy guard: market %s is not exact dormant truth", market)
		}
	}
	return nil
}

func validateService(service ServicePreimage, capturedAt time.Time) error {
	if !validIdentity(service.Name) || !validDigest(service.CurrentImageDigest) || !validDigest(service.TargetImageDigest) {
		return errors.New("invalid service identity or mutable image")
	}
	if service.Timeout <= 0 || service.Timeout > MaxServiceTimeout {
		return errors.New("service timeout is outside the frozen bound")
	}
	if !canonicalEnvironmentKeys(service.EnvironmentKeys) || !canonicalMounts(service.Mounts) {
		return errors.New("environment or mount identity is incomplete/noncanonical")
	}
	if service.CurrentSchemaVersion == 0 || service.PostReplaceSchemaVersion == 0 ||
		!validCompatibility(service.TargetSchema) || !validCompatibility(service.RollbackSchema) ||
		!service.TargetSchema.Readable.Contains(service.CurrentSchemaVersion) ||
		!service.TargetSchema.Writable.Contains(service.PostReplaceSchemaVersion) ||
		!service.RollbackSchema.Readable.Contains(service.PostReplaceSchemaVersion) ||
		!service.RollbackSchema.Writable.Contains(service.PostReplaceSchemaVersion) {
		return errors.New("schema compatibility preimage is incomplete")
	}
	if !service.BaselineHealth.Healthy || service.BaselineHealth.ObservedAt.IsZero() ||
		service.BaselineHealth.ObservedAt.Location() != time.UTC || service.BaselineHealth.ObservedAt.After(capturedAt) ||
		capturedAt.Sub(service.BaselineHealth.ObservedAt) > MaxBaselineHealthAge ||
		!validDigest(service.BaselineHealth.EvidenceDigest) {
		return errors.New("baseline health is incomplete")
	}
	return nil
}

func validCompatibility(value SchemaCompatibility) bool {
	return value.Readable.Min != 0 && value.Readable.Min <= value.Readable.Max &&
		value.Writable.Min != 0 && value.Writable.Min <= value.Writable.Max
}

func canonicalEnvironmentKeys(keys []string) bool {
	if len(keys) == 0 || !sort.StringsAreSorted(keys) {
		return false
	}
	for index, key := range keys {
		if !validEnvironmentKey(key) || index > 0 && keys[index-1] == key {
			return false
		}
	}
	return true
}

func validEnvironmentKey(value string) bool {
	if value == "" || len(value) > 128 {
		return false
	}
	for index, character := range value {
		if character == '_' || character >= 'A' && character <= 'Z' || index > 0 && character >= '0' && character <= '9' {
			continue
		}
		return false
	}
	return true
}

func canonicalMounts(mounts []MountIdentity) bool {
	if len(mounts) == 0 {
		return false
	}
	previous := ""
	for _, mount := range mounts {
		if mount.Type != MountBind && mount.Type != MountVolume || mount.Mode != MountReadOnly && mount.Mode != MountReadWrite ||
			!validPathIdentity(mount.Source) || !validPathIdentity(mount.Target) || !validDigest(mount.IdentityDigest) ||
			previous != "" && previous >= mount.Target {
			return false
		}
		previous = mount.Target
	}
	return true
}

func validPathIdentity(value string) bool {
	return strings.HasPrefix(value, "/") && value == strings.TrimSpace(value) && !strings.Contains(value, "\x00")
}

func validIdentity(value string) bool {
	if value == "" || len(value) > 128 || !utf8.ValidString(value) {
		return false
	}
	for _, character := range value {
		if unicode.IsSpace(character) || unicode.IsControl(character) {
			return false
		}
	}
	return true
}

func validDigest(value Digest) bool {
	text := string(value)
	if !strings.HasPrefix(text, "sha256:") || len(text) != len("sha256:")+sha256.Size*2 {
		return false
	}
	hexText := strings.TrimPrefix(text, "sha256:")
	if strings.ToLower(hexText) != hexText {
		return false
	}
	_, err := hex.DecodeString(hexText)
	return err == nil
}

func clonePreimage(input Preimage) Preimage {
	out := input
	out.State = cloneState(input.State)
	out.Services = make([]ServicePreimage, len(input.Services))
	for index, service := range input.Services {
		out.Services[index] = cloneService(service)
	}
	return out
}

func cloneState(input StateEvidence) StateEvidence {
	out := input
	out.Markets = make(map[Market]MarketState, len(input.Markets))
	for market, state := range input.Markets {
		out.Markets[market] = state
	}
	return out
}

func cloneService(input ServicePreimage) ServicePreimage {
	out := input
	out.EnvironmentKeys = append([]string(nil), input.EnvironmentKeys...)
	out.Mounts = append([]MountIdentity(nil), input.Mounts...)
	return out
}

func preservationMatches(service ServicePreimage, state StateEvidence, result Result) bool {
	return reflect.DeepEqual(state, result.State) && reflect.DeepEqual(service.EnvironmentKeys, result.EnvironmentKeys) &&
		reflect.DeepEqual(service.Mounts, result.Mounts)
}
