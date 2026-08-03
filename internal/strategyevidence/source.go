package strategyevidence

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"
)

var (
	ErrSourceDisabled         = errors.New("strategy evidence: source disabled")
	ErrSourceUnavailable      = errors.New("strategy evidence: official source contract unavailable")
	ErrSourceBoundExceeded    = errors.New("strategy evidence: source bound exceeded")
	ErrSourceCredential       = errors.New("strategy evidence: source credential unavailable")
	ErrSourceRateLimited      = errors.New("strategy evidence: source rate limited")
	ErrSourceIncomplete       = errors.New("strategy evidence: source response incomplete")
	ErrSourceSchemaDrift      = errors.New("strategy evidence: source schema drift")
	ErrSourceRetriesExhausted = errors.New("strategy evidence: source retries exhausted")
)

type RetryAfterPolicy string

const RetryAfterBounded RetryAfterPolicy = "HONOR_BOUNDED"

type SourcePolicy struct {
	Version            string
	ContractID         string
	Authority          SourceAuthority
	ContractVerified   bool
	EndpointIdentity   string
	EndpointVersion    string
	Method             string
	SchemaVersion      string
	AccessContract     string
	RequestIdentity    string
	CredentialRequired bool
	AbsoluteCallWindow time.Duration
	MaxCalls           int
	MaxPages           int
	PageSize           int
	MaxResponseBytes   int64
	MaxConcurrency     int
	RequestDeadline    time.Duration
	OperationDeadline  time.Duration
	RetryableStatuses  map[int]struct{}
	MaxRetries         int
	RetryAfterPolicy   RetryAfterPolicy
	contractSeal       string
}

type SourcePolicyConfig struct {
	Authority          SourceAuthority
	ContractID         string
	RequestIdentity    string
	AbsoluteCallWindow time.Duration
	MaxCalls           int
	MaxPages           int
	PageSize           int
	MaxResponseBytes   int64
	MaxConcurrency     int
	RequestDeadline    time.Duration
	OperationDeadline  time.Duration
	RetryableStatuses  []int
	MaxRetries         int
	RetryAfterPolicy   RetryAfterPolicy
}

const (
	secContractID  = "sec-submissions-2025-04-08"
	dartContractID = "opendart-disclosure-list-2019001"
)

func MintSourcePolicy(config SourcePolicyConfig) (SourcePolicy, error) {
	if config.Authority == AuthorityKRX {
		return SourcePolicy{}, ErrSourceUnavailable
	}
	p := SourcePolicy{
		Authority: config.Authority, ContractID: strings.TrimSpace(config.ContractID),
		RequestIdentity: strings.TrimSpace(config.RequestIdentity), AbsoluteCallWindow: config.AbsoluteCallWindow,
		MaxCalls: config.MaxCalls, MaxPages: config.MaxPages, PageSize: config.PageSize,
		MaxResponseBytes: config.MaxResponseBytes, MaxConcurrency: config.MaxConcurrency,
		RequestDeadline: config.RequestDeadline, OperationDeadline: config.OperationDeadline,
		RetryableStatuses: make(map[int]struct{}, len(config.RetryableStatuses)), MaxRetries: config.MaxRetries,
		RetryAfterPolicy: config.RetryAfterPolicy, ContractVerified: true, AccessContract: "official", Method: "GET",
	}
	for _, status := range config.RetryableStatuses {
		if status < 400 || status > 599 {
			return SourcePolicy{}, ErrSourceDisabled
		}
		p.RetryableStatuses[status] = struct{}{}
	}
	switch config.Authority {
	case AuthoritySEC:
		p.Version = "sec-edgar-source-v1"
		p.EndpointIdentity = "https://data.sec.gov/submissions/CIK##########.json"
		p.EndpointVersion = "2025-04-08"
		p.SchemaVersion = "sec-submissions-v1"
		p.CredentialRequired = false
		if p.ContractID != secContractID || !validSECRequestIdentity(p.RequestIdentity) || p.MaxCalls > 10 || p.AbsoluteCallWindow < time.Second || p.PageSize != 100 {
			return SourcePolicy{}, ErrSourceDisabled
		}
	case AuthorityOpenDART:
		p.Version = "opendart-source-v1"
		p.EndpointIdentity = "https://opendart.fss.or.kr/api/list.json"
		p.EndpointVersion = "2019001"
		p.SchemaVersion = "opendart-list-json-v1"
		p.RequestIdentity = "credential-provider"
		p.CredentialRequired = true
		if p.ContractID != dartContractID || p.PageSize <= 0 || p.PageSize > 100 {
			return SourcePolicy{}, ErrSourceDisabled
		}
	default:
		return SourcePolicy{}, ErrSourceDisabled
	}
	if err := p.validate(); err != nil {
		return SourcePolicy{}, err
	}
	p.contractSeal = p.officialSeal()
	return p, nil
}

func validSECRequestIdentity(value string) bool {
	value = strings.TrimSpace(value)
	at := strings.LastIndexByte(value, '@')
	space := strings.IndexByte(value, ' ')
	return space > 0 && at > space+1 && at < len(value)-1 && strings.Contains(value[at:], ".")
}

func (p SourcePolicy) officialSeal() string {
	statuses := make([]int, 0, len(p.RetryableStatuses))
	for status := range p.RetryableStatuses {
		statuses = append(statuses, status)
	}
	sort.Ints(statuses)
	return fmt.Sprint(p.Version, "\x00", p.ContractID, "\x00", p.Authority, "\x00", p.EndpointIdentity, "\x00", p.EndpointVersion, "\x00", p.Method, "\x00", p.SchemaVersion, "\x00", p.AccessContract,
		"\x00", p.RequestIdentity, "\x00", p.CredentialRequired, "\x00", p.AbsoluteCallWindow, "\x00", p.MaxCalls, "\x00", p.MaxPages, "\x00", p.PageSize, "\x00", p.MaxResponseBytes, "\x00", p.MaxConcurrency,
		"\x00", p.RequestDeadline, "\x00", p.OperationDeadline, "\x00", p.MaxRetries, "\x00", statuses, "\x00", p.RetryAfterPolicy)
}

func (p SourcePolicy) validateOfficial() error {
	if err := p.validate(); err != nil {
		return err
	}
	if p.contractSeal == "" || p.contractSeal != p.officialSeal() {
		return ErrSourceDisabled
	}
	switch p.Authority {
	case AuthoritySEC:
		if p.ContractID != secContractID || p.EndpointIdentity != "https://data.sec.gov/submissions/CIK##########.json" || p.Method != "GET" || p.SchemaVersion != "sec-submissions-v1" || p.CredentialRequired || !validSECRequestIdentity(p.RequestIdentity) {
			return ErrSourceDisabled
		}
	case AuthorityOpenDART:
		if p.ContractID != dartContractID || p.EndpointIdentity != "https://opendart.fss.or.kr/api/list.json" || p.Method != "GET" || p.SchemaVersion != "opendart-list-json-v1" || p.RequestIdentity != "credential-provider" || !p.CredentialRequired {
			return ErrSourceDisabled
		}
	default:
		return ErrSourceUnavailable
	}
	return nil
}

func (p SourcePolicy) validate() error {
	if !authorityValid(p.Authority) {
		return ErrSourceDisabled
	}
	if !p.ContractVerified {
		if p.Authority == AuthorityKRX {
			return ErrSourceUnavailable
		}
		return ErrSourceDisabled
	}
	stringsRequired := []string{p.Version, p.EndpointIdentity, p.EndpointVersion, p.Method, p.SchemaVersion, p.AccessContract, p.RequestIdentity}
	for _, value := range stringsRequired {
		if strings.TrimSpace(value) == "" {
			return ErrSourceDisabled
		}
	}
	if p.AccessContract != "official" || p.AbsoluteCallWindow <= 0 || p.MaxCalls <= 0 || p.MaxPages <= 0 || p.MaxResponseBytes <= 0 || p.MaxConcurrency <= 0 || p.RequestDeadline <= 0 || p.OperationDeadline <= 0 || p.RequestDeadline > p.OperationDeadline || len(p.RetryableStatuses) == 0 || p.MaxRetries < 0 || p.RetryAfterPolicy != RetryAfterBounded {
		return ErrSourceDisabled
	}
	return nil
}

type Credential struct{ reveal func() string }

func (c Credential) String() string { return "[REDACTED]" }

func (c Credential) secret() string {
	if c.reveal == nil {
		return ""
	}
	return c.reveal()
}

type CredentialProvider interface {
	Credential(context.Context, SourceAuthority) (Credential, error)
}

type staticCredential func() string

func StaticCredential(value string) CredentialProvider {
	return staticCredential(func() string { return value })
}

func (staticCredential) String() string { return "[REDACTED]" }

func (s staticCredential) Credential(_ context.Context, _ SourceAuthority) (Credential, error) {
	value := s()
	if strings.TrimSpace(value) == "" {
		return Credential{}, ErrSourceCredential
	}
	return Credential{reveal: func() string { return value }}, nil
}

type RequestMetadata struct {
	PolicyVersion             string          `json:"policy_version"`
	Authority                 SourceAuthority `json:"authority"`
	EndpointIdentity          string          `json:"endpoint_identity"`
	EndpointVersion           string          `json:"endpoint_version"`
	SchemaVersion             string          `json:"schema_version"`
	Method                    string          `json:"method"`
	RequestIdentityConfigured bool            `json:"request_identity_configured"`
	OperationID               string          `json:"operation_id"`
	PageLimit                 int             `json:"page_limit"`
	ResponseByteLimit         int64           `json:"response_byte_limit"`
}

func (m RequestMetadata) String() string {
	return fmt.Sprintf("%s/%s %s op=%s pages=%d bytes=%d", m.Authority, m.EndpointVersion, m.Method, m.OperationID, m.PageLimit, m.ResponseByteLimit)
}

type FetchRequest struct {
	OperationID       string
	PageLimit         int
	ResponseByteLimit int64
	Concurrency       int
	RequestDeadline   time.Duration
	OperationDeadline time.Duration
	Page              int
	PageSize          int
	Resource          string
}

type TransportRequest struct {
	Metadata          RequestMetadata
	Attempt           int
	Page              int
	PageSize          int
	Resource          string
	ResponseByteLimit int64
	requestIdentity   string
}

type TransportResponse struct {
	Status     int
	Body       []byte
	RetryAfter time.Duration
	ObservedAt time.Time
}

type Transport interface {
	Do(context.Context, TransportRequest, Credential) (TransportResponse, error)
}

type FetchResult struct {
	Metadata   RequestMetadata
	Body       []byte
	Attempts   int
	observedAt time.Time
}

type Adapter struct {
	policy      SourcePolicy
	transport   Transport
	credentials CredentialProvider
	now         func() time.Time
	waiter      Waiter
	mu          sync.Mutex
	windowFrom  time.Time
	calls       int
	active      int
	shared      *SharedRateBudget
	budgetKey   string
}

type SharedRateBudget struct {
	mu      sync.Mutex
	windows map[string]*rateWindow
}

type rateWindow struct {
	from   time.Time
	calls  int
	active int
}

func NewSharedRateBudget() *SharedRateBudget {
	return &SharedRateBudget{windows: make(map[string]*rateWindow)}
}

func NewAdapter(policy SourcePolicy, transport Transport, credentials CredentialProvider) *Adapter {
	return &Adapter{policy: policy, transport: transport, credentials: credentials, now: time.Now, waiter: timerWaiter{}}
}

type Waiter interface {
	Wait(context.Context, time.Duration) error
}

type timerWaiter struct{}

func (timerWaiter) Wait(ctx context.Context, duration time.Duration) error {
	if duration <= 0 {
		return nil
	}
	timer := time.NewTimer(duration)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (a *Adapter) Fetch(ctx context.Context, request FetchRequest) (FetchResult, error) {
	if err := a.policy.validateOfficial(); err != nil {
		return FetchResult{}, err
	}
	if a.transport == nil {
		return FetchResult{}, ErrSourceDisabled
	}
	request.OperationID = strings.TrimSpace(request.OperationID)
	if request.OperationID == "" {
		return FetchResult{}, ErrSourceBoundExceeded
	}
	if request.PageLimit == 0 {
		request.PageLimit = a.policy.MaxPages
	}
	if request.ResponseByteLimit == 0 {
		request.ResponseByteLimit = a.policy.MaxResponseBytes
	}
	if request.Concurrency == 0 {
		request.Concurrency = 1
	}
	if request.RequestDeadline == 0 {
		request.RequestDeadline = a.policy.RequestDeadline
	}
	if request.OperationDeadline == 0 {
		request.OperationDeadline = a.policy.OperationDeadline
	}
	if request.Page == 0 {
		request.Page = 1
	}
	if request.PageSize == 0 {
		request.PageSize = a.policy.PageSize
	}
	if request.PageLimit <= 0 || request.ResponseByteLimit <= 0 || request.Concurrency <= 0 || request.RequestDeadline <= 0 || request.OperationDeadline <= 0 || request.RequestDeadline > request.OperationDeadline || request.PageLimit > a.policy.MaxPages || request.ResponseByteLimit > a.policy.MaxResponseBytes || request.Concurrency > a.policy.MaxConcurrency || request.RequestDeadline > a.policy.RequestDeadline || request.OperationDeadline > a.policy.OperationDeadline || request.Page <= 0 || request.Page > request.PageLimit || request.PageSize < 0 || (a.policy.PageSize > 0 && request.PageSize > a.policy.PageSize) {
		return FetchResult{}, ErrSourceBoundExceeded
	}
	credential := Credential{}
	if a.policy.CredentialRequired {
		if a.credentials == nil {
			return FetchResult{}, ErrSourceCredential
		}
		var err error
		credential, err = a.credentials.Credential(ctx, a.policy.Authority)
		if err != nil || strings.TrimSpace(credential.secret()) == "" {
			return FetchResult{}, ErrSourceCredential
		}
	}
	metadata := RequestMetadata{
		PolicyVersion: a.policy.Version, Authority: a.policy.Authority,
		EndpointIdentity: a.policy.EndpointIdentity, EndpointVersion: a.policy.EndpointVersion,
		SchemaVersion: a.policy.SchemaVersion, Method: a.policy.Method,
		RequestIdentityConfigured: true, OperationID: request.OperationID,
		PageLimit: request.PageLimit, ResponseByteLimit: request.ResponseByteLimit,
	}
	if err := a.acquire(); err != nil {
		return FetchResult{}, err
	}
	defer a.release()
	operationCtx, cancel := context.WithTimeout(ctx, request.OperationDeadline)
	defer cancel()
	for attempt := 0; attempt <= a.policy.MaxRetries; attempt++ {
		if err := a.consumeCall(); err != nil {
			return FetchResult{Metadata: metadata, Attempts: attempt}, err
		}
		requestCtx, requestCancel := context.WithTimeout(operationCtx, request.RequestDeadline)
		response, err := a.transport.Do(requestCtx, TransportRequest{Metadata: metadata, Attempt: attempt, Page: request.Page, PageSize: request.PageSize, Resource: request.Resource, ResponseByteLimit: request.ResponseByteLimit, requestIdentity: a.policy.RequestIdentity}, credential)
		requestCancel()
		if err != nil {
			if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
				return FetchResult{Metadata: metadata, Attempts: attempt + 1}, err
			}
			if attempt == a.policy.MaxRetries {
				return FetchResult{Metadata: metadata, Attempts: attempt + 1}, fmt.Errorf("%w: %v", ErrSourceRetriesExhausted, err)
			}
			continue
		}
		if int64(len(response.Body)) > request.ResponseByteLimit {
			return FetchResult{Metadata: metadata, Attempts: attempt + 1}, ErrSourceBoundExceeded
		}
		if response.Status >= 200 && response.Status < 300 {
			return FetchResult{Metadata: metadata, Body: append([]byte(nil), response.Body...), Attempts: attempt + 1, observedAt: response.ObservedAt}, nil
		}
		if response.Status == 401 || response.Status == 403 {
			return FetchResult{Metadata: metadata, Attempts: attempt + 1}, ErrSourceCredential
		}
		if _, retryable := a.policy.RetryableStatuses[response.Status]; !retryable {
			return FetchResult{Metadata: metadata, Attempts: attempt + 1}, ErrSourceIncomplete
		} else if attempt == a.policy.MaxRetries {
			if response.Status == 429 {
				return FetchResult{Metadata: metadata, Attempts: attempt + 1}, ErrSourceRateLimited
			}
			return FetchResult{Metadata: metadata, Attempts: attempt + 1}, ErrSourceRetriesExhausted
		}
		if response.RetryAfter < 0 || response.RetryAfter > request.OperationDeadline {
			return FetchResult{Metadata: metadata, Attempts: attempt + 1}, ErrSourceBoundExceeded
		}
		if err := a.waiter.Wait(operationCtx, response.RetryAfter); err != nil {
			return FetchResult{Metadata: metadata, Attempts: attempt + 1}, err
		}
	}
	return FetchResult{Metadata: metadata}, ErrSourceIncomplete
}

func (a *Adapter) acquire() error {
	if a.shared != nil {
		return a.shared.acquire(a.budgetKey, a.policy)
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.active >= a.policy.MaxConcurrency {
		return ErrSourceRateLimited
	}
	a.active++
	return nil
}

func (a *Adapter) release() {
	if a.shared != nil {
		a.shared.release(a.budgetKey)
		return
	}
	a.mu.Lock()
	a.active--
	a.mu.Unlock()
}

func (a *Adapter) consumeCall() error {
	if a.shared != nil {
		return a.shared.consume(a.budgetKey, a.policy, a.now())
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	now := a.now()
	if a.windowFrom.IsZero() || now.Sub(a.windowFrom) >= a.policy.AbsoluteCallWindow {
		a.windowFrom = now
		a.calls = 0
	}
	if a.calls >= a.policy.MaxCalls {
		return ErrSourceRateLimited
	}
	a.calls++
	return nil
}

func (b *SharedRateBudget) window(key string) *rateWindow {
	w := b.windows[key]
	if w == nil {
		w = &rateWindow{}
		b.windows[key] = w
	}
	return w
}

func (b *SharedRateBudget) acquire(key string, policy SourcePolicy) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	w := b.window(key)
	if w.active >= policy.MaxConcurrency {
		return ErrSourceRateLimited
	}
	w.active++
	return nil
}

func (b *SharedRateBudget) release(key string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	w := b.window(key)
	if w.active > 0 {
		w.active--
	}
}

func (b *SharedRateBudget) consume(key string, policy SourcePolicy, now time.Time) error {
	b.mu.Lock()
	defer b.mu.Unlock()
	w := b.window(key)
	if w.from.IsZero() || now.Sub(w.from) >= policy.AbsoluteCallWindow {
		w.from = now
		w.calls = 0
	}
	if w.calls >= policy.MaxCalls {
		return ErrSourceRateLimited
	}
	w.calls++
	return nil
}
