package strategyrouter

import (
	"crypto/sha256"
	"errors"
	"sync"
	"time"
)

type PhysicalQuotaKey struct {
	Endpoint        string
	ResetGeneration string
}

func validQuotaKey(key PhysicalQuotaKey) bool {
	return boundedNonEmpty("quota endpoint", key.Endpoint) == nil && boundedNonEmpty("quota reset generation", key.ResetGeneration) == nil
}

type QuotaSnapshot struct {
	Key                 PhysicalQuotaKey
	ReportedRemaining   uint64
	SafetyReserve       uint64
	ObservationCycleCap uint64
	AbsoluteIssuanceCap uint64
	ObservedAt          time.Time
	FreshUntil          time.Time
	Digest              string
	seal                [32]byte
}

type quotaSnapshotInput struct {
	Key                 PhysicalQuotaKey
	ReportedRemaining   uint64
	SafetyReserve       uint64
	ObservationCycleCap uint64
	AbsoluteIssuanceCap uint64
	ObservedAt          time.Time
	FreshUntil          time.Time
	Digest              string
}

func newQuotaSnapshot(input quotaSnapshotInput) (QuotaSnapshot, error) {
	snapshot := QuotaSnapshot{
		Key: input.Key, ReportedRemaining: input.ReportedRemaining, SafetyReserve: input.SafetyReserve,
		ObservationCycleCap: input.ObservationCycleCap, AbsoluteIssuanceCap: input.AbsoluteIssuanceCap,
		ObservedAt: input.ObservedAt.UTC(), FreshUntil: input.FreshUntil.UTC(), Digest: input.Digest,
	}
	if !validQuotaKey(snapshot.Key) || snapshot.ReportedRemaining < snapshot.SafetyReserve || snapshot.ObservationCycleCap == 0 || snapshot.AbsoluteIssuanceCap == 0 ||
		snapshot.ObservedAt.IsZero() || !snapshot.FreshUntil.After(snapshot.ObservedAt) || boundedNonEmpty("quota digest", snapshot.Digest) != nil {
		return QuotaSnapshot{}, errors.New("strategyrouter: invalid quota snapshot")
	}
	snapshot.seal = quotaSnapshotSeal(snapshot)
	return snapshot, nil
}

func quotaSnapshotSeal(snapshot QuotaSnapshot) [32]byte {
	h := sha256.New()
	for _, value := range []string{snapshot.Key.Endpoint, snapshot.Key.ResetGeneration, snapshot.ObservedAt.UTC().Format(time.RFC3339Nano), snapshot.FreshUntil.UTC().Format(time.RFC3339Nano), snapshot.Digest} {
		writeString(h, value)
	}
	for _, value := range []uint64{snapshot.ReportedRemaining, snapshot.SafetyReserve, snapshot.ObservationCycleCap, snapshot.AbsoluteIssuanceCap} {
		writeUint64(h, value)
	}
	var result [32]byte
	copy(result[:], h.Sum(nil))
	return result
}

type PollClass string

const PollLowPriority PollClass = "LOW_PRIORITY"

type AcquireRequest struct {
	Key           PhysicalQuotaKey
	Market        Market
	Horizon       Horizon
	PollClass     PollClass
	CoordinatorID string
	RequestID     string
	ObservedAt    time.Time
}

type Capability struct {
	Endpoint        string
	ResetGeneration string
	Market          Market
	Horizon         Horizon
	PollClass       PollClass
	CoordinatorID   string
	RequestID       string
	Ordinal         uint64
	Token           string
	seal            [32]byte
}

func capabilitySeal(capability Capability) [32]byte {
	h := sha256.New()
	for _, value := range []string{capability.Endpoint, capability.ResetGeneration, string(capability.Market), string(capability.Horizon), string(capability.PollClass), capability.CoordinatorID, capability.RequestID, capability.Token} {
		writeString(h, value)
	}
	writeUint64(h, capability.Ordinal)
	var result [32]byte
	copy(result[:], h.Sum(nil))
	return result
}

func capabilityToken(request AcquireRequest, ordinal uint64, digest string) string {
	h := sha256.New()
	for _, value := range []string{request.Key.Endpoint, request.Key.ResetGeneration, string(request.Market), string(request.Horizon), string(request.PollClass), request.CoordinatorID, request.RequestID, digest} {
		writeString(h, value)
	}
	writeUint64(h, ordinal)
	return stringHex(h.Sum(nil))
}

func stringHex(data []byte) string {
	const digits = "0123456789abcdef"
	encoded := make([]byte, len(data)*2)
	for index, value := range data {
		encoded[index*2] = digits[value>>4]
		encoded[index*2+1] = digits[value&0x0f]
	}
	return string(encoded)
}

type AcquireResult struct {
	Code       RefusalCode
	Capability Capability
}

type QuotaStatus struct {
	Issued        uint64
	Outstanding   uint64
	SafetyReserve uint64
	Available     uint64
}

type quotaCommitment struct {
	requestFingerprint [32]byte
	capability         Capability
	completed          bool
}

type quotaBucket struct {
	snapshot    QuotaSnapshot
	allowance   uint64
	issued      uint64
	outstanding uint64
	requests    map[string]*quotaCommitment
	tokens      map[string]*quotaCommitment
}

type QuotaAuthority struct {
	mu      sync.Mutex
	buckets map[PhysicalQuotaKey]*quotaBucket
	now     func() time.Time
}

func NewQuotaAuthority() *QuotaAuthority {
	return newQuotaAuthority(time.Now)
}

func newQuotaAuthority(now func() time.Time) *QuotaAuthority {
	if now == nil {
		now = time.Now
	}
	return &QuotaAuthority{buckets: make(map[PhysicalQuotaKey]*quotaBucket), now: now}
}

func (authority *QuotaAuthority) Install(snapshot QuotaSnapshot) error {
	if authority == nil || snapshot.seal != quotaSnapshotSeal(snapshot) || !validQuotaKey(snapshot.Key) {
		return errors.New("strategyrouter: invalid sealed quota snapshot")
	}
	authority.mu.Lock()
	defer authority.mu.Unlock()
	if _, exists := authority.buckets[snapshot.Key]; exists {
		return errors.New("strategyrouter: quota reset generation already installed")
	}
	allowance := snapshot.ReportedRemaining - snapshot.SafetyReserve
	if snapshot.ObservationCycleCap < allowance {
		allowance = snapshot.ObservationCycleCap
	}
	if snapshot.AbsoluteIssuanceCap < allowance {
		allowance = snapshot.AbsoluteIssuanceCap
	}
	authority.buckets[snapshot.Key] = &quotaBucket{snapshot: snapshot, allowance: allowance, requests: make(map[string]*quotaCommitment), tokens: make(map[string]*quotaCommitment)}
	return nil
}

func acquireFingerprint(request AcquireRequest) [32]byte {
	h := sha256.New()
	for _, value := range []string{request.Key.Endpoint, request.Key.ResetGeneration, string(request.Market), string(request.Horizon), string(request.PollClass), request.CoordinatorID, request.RequestID, request.ObservedAt.UTC().Format(time.RFC3339Nano)} {
		writeString(h, value)
	}
	var result [32]byte
	copy(result[:], h.Sum(nil))
	return result
}

func validAcquireRequest(request AcquireRequest) bool {
	return validQuotaKey(request.Key) && validMarket(request.Market) && validHorizon(request.Horizon) && request.PollClass == PollLowPriority &&
		boundedNonEmpty("quota coordinator", request.CoordinatorID) == nil && boundedNonEmpty("quota request id", request.RequestID) == nil && !request.ObservedAt.IsZero()
}

func (authority *QuotaAuthority) Acquire(request AcquireRequest) AcquireResult {
	if authority == nil || !validAcquireRequest(request) {
		return AcquireResult{Code: RefusalInvalid}
	}
	authority.mu.Lock()
	defer authority.mu.Unlock()
	bucket, ok := authority.buckets[request.Key]
	if !ok {
		return AcquireResult{Code: RefusalReplay}
	}
	trustedNow := authority.now().UTC()
	if trustedNow.Before(bucket.snapshot.ObservedAt) || !trustedNow.Before(bucket.snapshot.FreshUntil) {
		return AcquireResult{Code: RefusalBudgetDeferred}
	}
	fingerprint := acquireFingerprint(request)
	if commitment, exists := bucket.requests[request.RequestID]; exists {
		if commitment.requestFingerprint == fingerprint {
			return AcquireResult{Code: RefusalDuplicate, Capability: commitment.capability}
		}
		return AcquireResult{Code: RefusalReplay}
	}
	if bucket.issued >= bucket.allowance {
		return AcquireResult{Code: RefusalBudgetDeferred}
	}
	ordinal := bucket.issued + 1
	capability := Capability{
		Endpoint: request.Key.Endpoint, ResetGeneration: request.Key.ResetGeneration, Market: request.Market, Horizon: request.Horizon,
		PollClass: request.PollClass, CoordinatorID: request.CoordinatorID, RequestID: request.RequestID, Ordinal: ordinal,
	}
	capability.Token = capabilityToken(request, ordinal, bucket.snapshot.Digest)
	capability.seal = capabilitySeal(capability)
	commitment := &quotaCommitment{requestFingerprint: fingerprint, capability: capability}
	bucket.requests[request.RequestID] = commitment
	bucket.tokens[capability.Token] = commitment
	bucket.issued++
	bucket.outstanding++
	return AcquireResult{Capability: capability}
}

func (authority *QuotaAuthority) Complete(capability Capability, market Market, horizon Horizon) RefusalCode {
	if authority == nil || !validMarket(market) || !validHorizon(horizon) {
		return RefusalInvalid
	}
	if capability.Market != market || capability.Horizon != horizon {
		return RefusalScopeMismatch
	}
	if capability.Token == "" || capability.seal != capabilitySeal(capability) {
		return RefusalReplay
	}
	key := PhysicalQuotaKey{Endpoint: capability.Endpoint, ResetGeneration: capability.ResetGeneration}
	authority.mu.Lock()
	defer authority.mu.Unlock()
	bucket, ok := authority.buckets[key]
	if !ok {
		return RefusalReplay
	}
	commitment, ok := bucket.tokens[capability.Token]
	if !ok || commitment.completed || commitment.capability != capability {
		return RefusalReplay
	}
	commitment.completed = true
	if bucket.outstanding > 0 {
		bucket.outstanding--
	}
	return RefusalNone
}

func (authority *QuotaAuthority) Status(key PhysicalQuotaKey) QuotaStatus {
	if authority == nil {
		return QuotaStatus{}
	}
	authority.mu.Lock()
	defer authority.mu.Unlock()
	bucket, ok := authority.buckets[key]
	if !ok {
		return QuotaStatus{}
	}
	available := uint64(0)
	if bucket.allowance > bucket.issued {
		available = bucket.allowance - bucket.issued
	}
	return QuotaStatus{Issued: bucket.issued, Outstanding: bucket.outstanding, SafetyReserve: bucket.snapshot.SafetyReserve, Available: available}
}
