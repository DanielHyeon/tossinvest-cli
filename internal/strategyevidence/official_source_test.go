package strategyevidence

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestMintSourcePolicyFreezesOfficialAuthorityContracts(t *testing.T) {
	t.Parallel()
	sec, err := MintSourcePolicy(validSECConfig())
	if err != nil {
		t.Fatal(err)
	}
	if sec.EndpointIdentity != "https://data.sec.gov/submissions/CIK##########.json" || sec.CredentialRequired || sec.PageSize != 100 {
		t.Fatalf("SEC policy = %+v", sec)
	}
	dart, err := MintSourcePolicy(validDARTConfig())
	if err != nil {
		t.Fatal(err)
	}
	if dart.EndpointIdentity != "https://opendart.fss.or.kr/api/list.json" || !dart.CredentialRequired || dart.PageSize != 100 {
		t.Fatalf("OpenDART policy = %+v", dart)
	}

	badIdentity := validSECConfig()
	badIdentity.RequestIdentity = "anonymous-bot"
	if _, err := MintSourcePolicy(badIdentity); !errors.Is(err, ErrSourceDisabled) {
		t.Fatalf("SEC identity error = %v", err)
	}
	overRate := validSECConfig()
	overRate.MaxCalls = 11
	if _, err := MintSourcePolicy(overRate); !errors.Is(err, ErrSourceDisabled) {
		t.Fatalf("SEC over-rate error = %v", err)
	}
	krx := validSECConfig()
	krx.Authority = AuthorityKRX
	if _, err := MintSourcePolicy(krx); !errors.Is(err, ErrSourceUnavailable) {
		t.Fatalf("KRX contract error = %v", err)
	}
}

func TestSECOfficialAdapterCollectsFrozenPaginatedFixtureWithDeclaredIdentity(t *testing.T) {
	t.Parallel()
	observed := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	transport := &scriptedOfficialTransport{responses: []TransportResponse{
		{Status: 200, Body: fixture(t, "sec_submissions_v1_page1.json"), ObservedAt: observed},
		{Status: 200, Body: fixture(t, "sec_submissions_v1_page2.json"), ObservedAt: observed.Add(time.Second)},
	}}
	adapter := mustOfficialAdapter(t, validSECConfig(), transport, nil, NewSharedRateBudget())
	sink := &recordingBatchSink{}
	batch, err := adapter.CollectAndCommit(context.Background(), OfficialCollectionRequest{OperationID: "sec-aapl", EntityID: "0000320193"}, sink)
	if err != nil {
		t.Fatal(err)
	}
	if !batch.Complete || batch.Pages != 2 || len(batch.Records) != 2 || sink.Calls() != 1 {
		t.Fatalf("batch=%+v commits=%d", batch, sink.Calls())
	}
	requests := transport.Requests()
	if len(requests) != 2 || requests[0].Resource != "CIK0000320193.json" || requests[1].Resource != "CIK0000320193-submissions-001.json" {
		t.Fatalf("requests=%+v", requests)
	}
	for _, request := range requests {
		if request.requestIdentity != validSECConfig().RequestIdentity || request.Metadata.RequestIdentityConfigured != true {
			t.Fatalf("SEC identity missing: %+v", request)
		}
	}
	if transport.LastCredential() != "" {
		t.Fatal("SEC public data adapter unexpectedly requested a credential")
	}
}

func TestOpenDARTOfficialAdapterKeepsKeyAtSecretBoundaryAndCollectsAllPages(t *testing.T) {
	t.Parallel()
	observed := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	transport := &scriptedOfficialTransport{responses: []TransportResponse{
		{Status: 200, Body: fixture(t, "opendart_list_v1_page1.json"), ObservedAt: observed},
		{Status: 200, Body: fixture(t, "opendart_list_v1_page2.json"), ObservedAt: observed.Add(time.Second)},
	}}
	secret := "fixture-dart-key-never-persist"
	adapter := mustOfficialAdapter(t, validDARTConfig(), transport, StaticCredential(secret), NewSharedRateBudget())
	sink := &recordingBatchSink{}
	batch, err := adapter.CollectAndCommit(context.Background(), OfficialCollectionRequest{OperationID: "dart-005930", EntityID: "00126380"}, sink)
	if err != nil {
		t.Fatal(err)
	}
	if len(batch.Records) != 2 || batch.Pages != 2 || sink.Calls() != 1 || transport.LastCredential() != secret {
		t.Fatalf("batch=%+v commits=%d credential-delivered=%v", batch, sink.Calls(), transport.LastCredential() == secret)
	}
	encoded, err := json.Marshal(batch)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), secret) || strings.Contains(batch.Digest, secret) {
		t.Fatalf("OpenDART credential leaked: %s", encoded)
	}
}

func TestOfficialAdaptersDoNotCommitIncompleteOrUntrustedBatches(t *testing.T) {
	t.Parallel()
	observed := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name       string
		config     SourcePolicyConfig
		credential CredentialProvider
		request    OfficialCollectionRequest
		responses  []TransportResponse
		want       error
	}{
		{
			name: "SEC schema drift", config: validSECConfig(), request: OfficialCollectionRequest{OperationID: "schema", EntityID: "0000320193"},
			responses: []TransportResponse{{Status: 200, Body: []byte(`{"cik":"0000320193","name":"APPLE INC","tickers":["AAPL"],"filings":{}}`), ObservedAt: observed}}, want: ErrSourceSchemaDrift,
		},
		{
			name: "SEC partial page bound", config: validSECConfig(), request: OfficialCollectionRequest{OperationID: "partial", EntityID: "0000320193", PageLimit: 1},
			responses: []TransportResponse{{Status: 200, Body: fixture(t, "sec_submissions_v1_page1.json"), ObservedAt: observed}}, want: ErrSourceIncomplete,
		},
		{
			name: "OpenDART auth failure", config: validDARTConfig(), credential: StaticCredential("bad-key"), request: OfficialCollectionRequest{OperationID: "auth", EntityID: "00126380"},
			responses: []TransportResponse{{Status: 200, Body: []byte(`{"status":"010","message":"등록되지 않은 키입니다."}`), ObservedAt: observed}}, want: ErrSourceCredential,
		},
		{
			name: "retry exhaustion", config: validSECConfig(), request: OfficialCollectionRequest{OperationID: "retry", EntityID: "0000320193"},
			responses: []TransportResponse{{Status: 503}, {Status: 503}, {Status: 503}}, want: ErrSourceRetriesExhausted,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			transport := &scriptedOfficialTransport{responses: tt.responses}
			adapter := mustOfficialAdapter(t, tt.config, transport, tt.credential, NewSharedRateBudget())
			sink := &recordingBatchSink{}
			batch, err := adapter.CollectAndCommit(context.Background(), tt.request, sink)
			if !errors.Is(err, tt.want) {
				t.Fatalf("want %v, got %v (batch=%+v)", tt.want, err, batch)
			}
			if sink.Calls() != 0 || batch.Complete {
				t.Fatalf("untrusted batch committed: commits=%d batch=%+v", sink.Calls(), batch)
			}
		})
	}
}

func TestSharedRateBudgetAppliesAcrossOfficialAdapterInstances(t *testing.T) {
	t.Parallel()
	config := validSECConfig()
	config.MaxCalls = 1
	budget := NewSharedRateBudget()
	firstTransport := &scriptedOfficialTransport{responses: []TransportResponse{{Status: 200, Body: singlePageSECFixture(), ObservedAt: time.Now().UTC()}}}
	secondTransport := &scriptedOfficialTransport{responses: []TransportResponse{{Status: 200, Body: singlePageSECFixture(), ObservedAt: time.Now().UTC()}}}
	first := mustOfficialAdapter(t, config, firstTransport, nil, budget)
	second := mustOfficialAdapter(t, config, secondTransport, nil, budget)
	if _, err := first.Collect(context.Background(), OfficialCollectionRequest{OperationID: "first", EntityID: "0000320193"}); err != nil {
		t.Fatal(err)
	}
	if _, err := second.Collect(context.Background(), OfficialCollectionRequest{OperationID: "second", EntityID: "0000320193"}); !errors.Is(err, ErrSourceRateLimited) {
		t.Fatalf("want shared rate limit, got %v", err)
	}
	if secondTransport.Calls() != 0 {
		t.Fatalf("second transport calls=%d, want 0", secondTransport.Calls())
	}
}

func TestOfficialAdapterRejectsTamperedMintedContractBeforeTransport(t *testing.T) {
	t.Parallel()
	policy, err := MintSourcePolicy(validSECConfig())
	if err != nil {
		t.Fatal(err)
	}
	for _, mutate := range []func(*SourcePolicy){
		func(p *SourcePolicy) { p.EndpointIdentity = "https://example.invalid/arbitrary" },
		func(p *SourcePolicy) { p.Method = "DELETE" },
		func(p *SourcePolicy) { p.SchemaVersion = "unreviewed-v2" },
	} {
		tampered := policy
		mutate(&tampered)
		transport := &scriptedOfficialTransport{}
		if _, err := NewOfficialAdapter(tampered, transport, nil, NewSharedRateBudget()); !errors.Is(err, ErrSourceDisabled) {
			t.Fatalf("tampered policy error=%v", err)
		}
		if transport.Calls() != 0 {
			t.Fatal("tampered official contract reached transport")
		}
	}
}

func TestOfficialAdapterRejectsPostMintRequestIdentityMutationBeforeTransport(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		config     SourcePolicyConfig
		mutated    string
		credential CredentialProvider
	}{
		{name: "SEC another syntactically valid identity", config: validSECConfig(), mutated: "Other Research ops@example.invalid"},
		{name: "OpenDART alternate credential boundary", config: validDARTConfig(), mutated: "other-credential-provider", credential: StaticCredential("test-only")},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy, err := MintSourcePolicy(tt.config)
			if err != nil {
				t.Fatal(err)
			}
			policy.RequestIdentity = tt.mutated
			transport := &scriptedOfficialTransport{}
			if _, err := NewOfficialAdapter(policy, transport, tt.credential, NewSharedRateBudget()); !errors.Is(err, ErrSourceDisabled) {
				t.Fatalf("request identity mutation error=%v", err)
			}
			if transport.Calls() != 0 {
				t.Fatal("request identity mutation reached transport")
			}
		})
	}
}

func TestOfficialPaginationAndAggregateByteBoundsFailClosed(t *testing.T) {
	t.Parallel()
	observed := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	t.Run("changed DART cursor metadata", func(t *testing.T) {
		page2 := strings.Replace(string(fixture(t, "opendart_list_v1_page2.json")), `"total_page": 2`, `"total_page": 3`, 1)
		transport := &scriptedOfficialTransport{responses: []TransportResponse{
			{Status: 200, Body: fixture(t, "opendart_list_v1_page1.json"), ObservedAt: observed},
			{Status: 200, Body: []byte(page2), ObservedAt: observed.Add(time.Second)},
		}}
		adapter := mustOfficialAdapter(t, validDARTConfig(), transport, StaticCredential("test-only"), NewSharedRateBudget())
		sink := &recordingBatchSink{}
		if _, err := adapter.CollectAndCommit(context.Background(), OfficialCollectionRequest{OperationID: "partial", EntityID: "00126380"}, sink); !errors.Is(err, ErrSourceSchemaDrift) {
			t.Fatalf("cursor drift error=%v", err)
		}
		if sink.Calls() != 0 {
			t.Fatal("cursor-drift batch committed")
		}
	})
	t.Run("aggregate bytes", func(t *testing.T) {
		first, second := fixture(t, "sec_submissions_v1_page1.json"), fixture(t, "sec_submissions_v1_page2.json")
		transport := &scriptedOfficialTransport{responses: []TransportResponse{{Status: 200, Body: first, ObservedAt: observed}, {Status: 200, Body: second, ObservedAt: observed}}}
		adapter := mustOfficialAdapter(t, validSECConfig(), transport, nil, NewSharedRateBudget())
		limit := int64(len(first) + len(second) - 1)
		if _, err := adapter.Collect(context.Background(), OfficialCollectionRequest{OperationID: "bytes", EntityID: "0000320193", ResponseByteLimit: limit}); !errors.Is(err, ErrSourceBoundExceeded) {
			t.Fatalf("aggregate byte error=%v", err)
		}
		requests := transport.Requests()
		if len(requests) != 2 || requests[1].ResponseByteLimit != limit-int64(len(first)) {
			t.Fatalf("streaming byte limits not propagated: %+v", requests)
		}
	})
}

func TestOfficialSchemaRejectsDuplicateKeysAndSetsDeadlines(t *testing.T) {
	t.Parallel()
	body := []byte(`{"cik":"0000320193","cik":"0000320193","name":"APPLE INC","tickers":["AAPL"],"exchanges":["Nasdaq"],"filings":{}}`)
	transport := &scriptedOfficialTransport{responses: []TransportResponse{{Status: 200, Body: body, ObservedAt: time.Now().UTC()}}}
	adapter := mustOfficialAdapter(t, validSECConfig(), transport, nil, NewSharedRateBudget())
	if _, err := adapter.Collect(context.Background(), OfficialCollectionRequest{OperationID: "duplicate", EntityID: "0000320193"}); !errors.Is(err, ErrSourceSchemaDrift) {
		t.Fatalf("duplicate-key error=%v", err)
	}
	transport.mu.Lock()
	deadlineSet := transport.deadlineSet
	transport.mu.Unlock()
	if !deadlineSet {
		t.Fatal("official transport request lacked a deadline")
	}
}

func TestSharedConcurrencyBudgetAppliesAcrossOfficialAdapterInstances(t *testing.T) {
	t.Parallel()
	config := validSECConfig()
	config.MaxConcurrency = 1
	budget := NewSharedRateBudget()
	blocking := &blockingTransport{started: make(chan struct{}), release: make(chan struct{})}
	first := mustOfficialAdapter(t, config, blocking, nil, budget)
	secondTransport := &scriptedOfficialTransport{responses: []TransportResponse{{Status: 200, Body: singlePageSECFixture(), ObservedAt: time.Now().UTC()}}}
	second := mustOfficialAdapter(t, config, secondTransport, nil, budget)
	done := make(chan error, 1)
	go func() {
		_, err := first.Collect(context.Background(), OfficialCollectionRequest{OperationID: "active", EntityID: "0000320193"})
		done <- err
	}()
	select {
	case <-blocking.started:
	case <-time.After(time.Second):
		t.Fatal("first shared-budget call did not start")
	}
	if _, err := second.Collect(context.Background(), OfficialCollectionRequest{OperationID: "blocked", EntityID: "0000320193"}); !errors.Is(err, ErrSourceRateLimited) {
		t.Fatalf("shared concurrency error=%v", err)
	}
	if secondTransport.Calls() != 0 {
		t.Fatal("shared concurrency rejection occurred after transport")
	}
	close(blocking.release)
	if err := <-done; !errors.Is(err, ErrSourceSchemaDrift) {
		t.Fatalf("first fixture completion error=%v", err)
	}
}

func validSECConfig() SourcePolicyConfig {
	return SourcePolicyConfig{
		Authority: AuthoritySEC, ContractID: "sec-submissions-2025-04-08",
		RequestIdentity:    "TossOS Research admin@example.invalid",
		AbsoluteCallWindow: time.Second, MaxCalls: 10, MaxPages: 4, PageSize: 100,
		MaxResponseBytes: 1 << 20, MaxConcurrency: 2,
		RequestDeadline: time.Second, OperationDeadline: 5 * time.Second,
		RetryableStatuses: []int{429, 503}, MaxRetries: 2, RetryAfterPolicy: RetryAfterBounded,
	}
}

func validDARTConfig() SourcePolicyConfig {
	return SourcePolicyConfig{
		Authority: AuthorityOpenDART, ContractID: "opendart-disclosure-list-2019001",
		AbsoluteCallWindow: 24 * time.Hour, MaxCalls: 1000, MaxPages: 4, PageSize: 100,
		MaxResponseBytes: 1 << 20, MaxConcurrency: 2,
		RequestDeadline: time.Second, OperationDeadline: 5 * time.Second,
		RetryableStatuses: []int{429, 503}, MaxRetries: 2, RetryAfterPolicy: RetryAfterBounded,
	}
}

func mustOfficialAdapter(t *testing.T, config SourcePolicyConfig, transport Transport, credentials CredentialProvider, budget *SharedRateBudget) *OfficialAdapter {
	t.Helper()
	policy, err := MintSourcePolicy(config)
	if err != nil {
		t.Fatal(err)
	}
	adapter, err := NewOfficialAdapter(policy, transport, credentials, budget)
	if err != nil {
		t.Fatal(err)
	}
	adapter.base.waiter = &fakeWaiter{}
	return adapter
}

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	body, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func singlePageSECFixture() []byte {
	return []byte(`{"cik":"0000320193","name":"APPLE INC","tickers":["AAPL"],"exchanges":["Nasdaq"],"filings":{"recent":{"accessionNumber":["0000320193-26-000001"],"filingDate":["2026-08-01"],"reportDate":["2026-07-31"],"acceptanceDateTime":["2026-08-01T16:30:00.000Z"],"form":["8-K"],"primaryDocument":["aapl.htm"]},"files":[]}}`)
}

type recordingBatchSink struct {
	mu      sync.Mutex
	commits int
}

func (s *recordingBatchSink) Commit(_ context.Context, batch OfficialBatch) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if !batch.Complete {
		return errors.New("test sink received incomplete batch")
	}
	s.commits++
	return nil
}

func (s *recordingBatchSink) Calls() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.commits
}

type scriptedOfficialTransport struct {
	mu          sync.Mutex
	responses   []TransportResponse
	requests    []TransportRequest
	credential  string
	deadlineSet bool
}

func (t *scriptedOfficialTransport) Do(ctx context.Context, request TransportRequest, credential Credential) (TransportResponse, error) {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.requests = append(t.requests, request)
	t.credential = credential.secret()
	_, t.deadlineSet = ctx.Deadline()
	if len(t.responses) == 0 {
		return TransportResponse{}, errors.New("fixture transport exhausted")
	}
	response := t.responses[0]
	t.responses = t.responses[1:]
	return response, nil
}

func (t *scriptedOfficialTransport) Calls() int {
	t.mu.Lock()
	defer t.mu.Unlock()
	return len(t.requests)
}

func (t *scriptedOfficialTransport) Requests() []TransportRequest {
	t.mu.Lock()
	defer t.mu.Unlock()
	return append([]TransportRequest(nil), t.requests...)
}

func (t *scriptedOfficialTransport) LastCredential() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.credential
}
