package strategyevidence

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestIncompleteOrUnverifiedPolicyMakesZeroCalls(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name   string
		mutate func(*SourcePolicy)
		want   error
	}{
		{"missing endpoint version", func(p *SourcePolicy) { p.EndpointVersion = "" }, ErrSourceDisabled},
		{"missing retryable statuses", func(p *SourcePolicy) { p.RetryableStatuses = map[int]struct{}{} }, ErrSourceDisabled},
		{"unofficial access", func(p *SourcePolicy) { p.AccessContract = "wts" }, ErrSourceDisabled},
		{"unverified KRX contract", func(p *SourcePolicy) { p.Authority = AuthorityKRX; p.ContractVerified = false }, ErrSourceUnavailable},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			transport := &fakeTransport{}
			policy := validSourcePolicy()
			tt.mutate(&policy)
			adapter := NewAdapter(policy, transport, StaticCredential("top-secret"))
			_, err := adapter.Fetch(context.Background(), FetchRequest{OperationID: "op-1"})
			if !errors.Is(err, tt.want) {
				t.Fatalf("want %v, got %v", tt.want, err)
			}
			if transport.Calls() != 0 {
				t.Fatalf("transport calls = %d", transport.Calls())
			}
		})
	}
}

func TestAdapterKeepsCredentialOutOfRequestMetadata(t *testing.T) {
	t.Parallel()
	transport := &fakeTransport{responses: []TransportResponse{{Status: 200, Body: []byte(`{"ok":true}`)}}}
	policy := validDARTSourcePolicy()
	adapter := NewAdapter(policy, transport, StaticCredential("top-secret"))
	result, err := adapter.Fetch(context.Background(), FetchRequest{OperationID: "op-1", PageLimit: 1})
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(result.Metadata)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(encoded), "top-secret") || strings.Contains(result.Metadata.String(), "top-secret") || strings.Contains(string(encoded), policy.RequestIdentity) {
		t.Fatalf("credential leaked in metadata: %s", encoded)
	}
	if transport.LastCredential() != "top-secret" {
		t.Fatal("credential was not delivered through the separate transport boundary")
	}
}

func TestAdapterRejectsRuntimeBoundExpansionBeforeTransport(t *testing.T) {
	t.Parallel()
	policy := validSourcePolicy()
	requests := []FetchRequest{
		{OperationID: "pages", PageLimit: policy.MaxPages + 1},
		{OperationID: "negative-pages", PageLimit: -1},
		{OperationID: "negative-bytes", ResponseByteLimit: -1},
		{OperationID: "negative-concurrency", Concurrency: -1},
		{OperationID: "deadline-order", RequestDeadline: 2 * time.Second, OperationDeadline: time.Second},
	}
	for _, request := range requests {
		request := request
		t.Run(request.OperationID, func(t *testing.T) {
			t.Parallel()
			transport := &fakeTransport{}
			adapter := NewAdapter(policy, transport, StaticCredential("secret"))
			_, err := adapter.Fetch(context.Background(), request)
			if !errors.Is(err, ErrSourceBoundExceeded) || transport.Calls() != 0 {
				t.Fatalf("err=%v calls=%d", err, transport.Calls())
			}
		})
	}
}

func TestAdapterEnforcesResponseBytesWindowAndRetryAfter(t *testing.T) {
	t.Parallel()
	t.Run("response bytes", func(t *testing.T) {
		transport := &fakeTransport{responses: []TransportResponse{{Status: 200, Body: make([]byte, 8)}}}
		policy := validSourcePolicy()
		policy.MaxResponseBytes = 4
		policy.contractSeal = policy.officialSeal()
		adapter := NewAdapter(policy, transport, StaticCredential("secret"))
		_, err := adapter.Fetch(context.Background(), FetchRequest{OperationID: "bytes"})
		if !errors.Is(err, ErrSourceBoundExceeded) {
			t.Fatalf("want bound error, got %v", err)
		}
	})
	t.Run("absolute window", func(t *testing.T) {
		transport := &fakeTransport{}
		policy := validSourcePolicy()
		policy.MaxCalls = 1
		policy.contractSeal = policy.officialSeal()
		adapter := NewAdapter(policy, transport, StaticCredential("secret"))
		if _, err := adapter.Fetch(context.Background(), FetchRequest{OperationID: "first"}); err != nil {
			t.Fatal(err)
		}
		if _, err := adapter.Fetch(context.Background(), FetchRequest{OperationID: "second"}); !errors.Is(err, ErrSourceRateLimited) {
			t.Fatalf("want rate limited, got %v", err)
		}
	})
	t.Run("bounded retry after", func(t *testing.T) {
		transport := &fakeTransport{responses: []TransportResponse{
			{Status: 429, RetryAfter: 25 * time.Millisecond},
			{Status: 200, Body: []byte(`{"ok":true}`)},
		}}
		waiter := &fakeWaiter{}
		adapter := NewAdapter(validSourcePolicy(), transport, StaticCredential("secret"))
		adapter.waiter = waiter
		result, err := adapter.Fetch(context.Background(), FetchRequest{OperationID: "retry"})
		if err != nil || result.Attempts != 2 {
			t.Fatalf("result=%+v err=%v", result, err)
		}
		if len(waiter.durations) != 1 || waiter.durations[0] != 25*time.Millisecond {
			t.Fatalf("waits=%v", waiter.durations)
		}
	})
}

func TestAdapterRejectsConcurrentCallBeforeSecondTransportRequest(t *testing.T) {
	t.Parallel()
	transport := &blockingTransport{started: make(chan struct{}), release: make(chan struct{})}
	policy := validSourcePolicy()
	policy.MaxConcurrency = 1
	adapter := NewAdapter(policy, transport, StaticCredential("secret"))
	firstDone := make(chan error, 1)
	go func() {
		_, err := adapter.Fetch(context.Background(), FetchRequest{OperationID: "first"})
		firstDone <- err
	}()
	select {
	case <-transport.started:
	case <-time.After(time.Second):
		t.Fatal("first transport call did not start")
	}
	if _, err := adapter.Fetch(context.Background(), FetchRequest{OperationID: "second"}); !errors.Is(err, ErrSourceRateLimited) {
		t.Fatalf("want concurrent rate limit, got %v", err)
	}
	if calls := transport.Calls(); calls != 1 {
		t.Fatalf("transport calls = %d, want 1", calls)
	}
	close(transport.release)
	if err := <-firstDone; err != nil {
		t.Fatal(err)
	}
}

func validSourcePolicy() SourcePolicy {
	policy, err := MintSourcePolicy(SourcePolicyConfig{
		Authority: AuthoritySEC, ContractID: secContractID, RequestIdentity: "TossOS Research test@example.invalid",
		AbsoluteCallWindow: time.Second, MaxCalls: 10, MaxPages: 2, PageSize: 100,
		MaxResponseBytes: 1024, MaxConcurrency: 1, RequestDeadline: time.Second, OperationDeadline: 5 * time.Second,
		RetryableStatuses: []int{429, 503}, MaxRetries: 2, RetryAfterPolicy: RetryAfterBounded,
	})
	if err != nil {
		panic(err)
	}
	return policy
}

func validDARTSourcePolicy() SourcePolicy {
	policy, err := MintSourcePolicy(SourcePolicyConfig{
		Authority: AuthorityOpenDART, ContractID: dartContractID,
		AbsoluteCallWindow: time.Hour, MaxCalls: 10, MaxPages: 2, PageSize: 100,
		MaxResponseBytes: 1024, MaxConcurrency: 1, RequestDeadline: time.Second, OperationDeadline: 5 * time.Second,
		RetryableStatuses: []int{429, 503}, MaxRetries: 2, RetryAfterPolicy: RetryAfterBounded,
	})
	if err != nil {
		panic(err)
	}
	return policy
}

type fakeTransport struct {
	mu         sync.Mutex
	calls      int
	credential string
	responses  []TransportResponse
}

type fakeWaiter struct{ durations []time.Duration }

type blockingTransport struct {
	mu      sync.Mutex
	calls   int
	started chan struct{}
	release chan struct{}
}

func (b *blockingTransport) Do(ctx context.Context, _ TransportRequest, _ Credential) (TransportResponse, error) {
	b.mu.Lock()
	b.calls++
	if b.calls == 1 {
		close(b.started)
	}
	b.mu.Unlock()
	select {
	case <-ctx.Done():
		return TransportResponse{}, ctx.Err()
	case <-b.release:
		return TransportResponse{Status: 200, Body: []byte(`{}`)}, nil
	}
}

func (b *blockingTransport) Calls() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.calls
}

func (f *fakeWaiter) Wait(_ context.Context, duration time.Duration) error {
	f.durations = append(f.durations, duration)
	return nil
}

func (f *fakeTransport) Do(_ context.Context, _ TransportRequest, credential Credential) (TransportResponse, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.calls++
	f.credential = credential.secret()
	if len(f.responses) == 0 {
		return TransportResponse{Status: 200, Body: []byte(`{}`)}, nil
	}
	response := f.responses[0]
	f.responses = f.responses[1:]
	return response, nil
}

func (f *fakeTransport) Calls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.calls
}

func (f *fakeTransport) LastCredential() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.credential
}
