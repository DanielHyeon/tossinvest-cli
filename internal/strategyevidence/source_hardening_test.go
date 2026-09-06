package strategyevidence

// 공식 출처 어댑터의 세 가지 안전 성질을 잰다.
//
//  1. 전송 계층 오류가 자격증명을 밖으로 나르지 않는다. OpenDART 는 crtfc_key 를
//     **질의 문자열**로 인증하므로 net/http 의 *url.Error 는 키가 통째로 들어간 URL 을
//     메시지에 담는다. 그 메시지를 그대로 돌려주면 로그에 키가 남는다.
//  2. 원격 응답이 다음 요청의 경로를 고르지 못한다.
//  3. 정책의 모든 필드가 "0보다 큰가"가 아니라 동결된 공식 계약 안에 있는가로 검사된다.
//     그리고 그 검사 하나를 지우면 정확히 그 필드의 하위 시험이 빨개진다.

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"testing"
	"time"
)

const leakedCredential = "crtfc-key-must-never-appear"

// leakingTransport 는 실제 net/http 가 만드는 오류 모양을 그대로 흉내 낸다.
// *url.Error 의 stripPassword 는 userinfo 만 지우고 질의 문자열은 남긴다.
type leakingTransport struct {
	calls int
	err   error
}

func (t *leakingTransport) Do(context.Context, TransportRequest, Credential) (TransportResponse, error) {
	t.calls++
	return TransportResponse{}, t.err
}

func credentialBearingURLError(cause error) error {
	return &url.Error{
		Op:  "Get",
		URL: "https://opendart.fss.or.kr/api/list.json?crtfc_key=" + leakedCredential + "&page_no=1",
		Err: cause,
	}
}

func TestTransportErrorsCarryNoCredentialOrURL(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		name  string
		cause error
		want  error
	}{
		{name: "deadline", cause: context.DeadlineExceeded, want: context.DeadlineExceeded},
		{name: "cancelled", cause: context.Canceled, want: context.Canceled},
		{name: "retries-exhausted", cause: errors.New("dial tcp 1.2.3.4:443: i/o timeout"), want: ErrSourceRetriesExhausted},
	} {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			transport := &leakingTransport{err: credentialBearingURLError(test.cause)}
			adapter := newAdapter(validDARTSourcePolicy(), transport, StaticCredential(leakedCredential))
			adapter.waiter = &fakeWaiter{}
			_, err := adapter.Fetch(context.Background(), FetchRequest{OperationID: "leak"})
			if err == nil {
				t.Fatal("transport failure produced no error")
			}
			if !errors.Is(err, test.want) {
				t.Fatalf("error %v is not %v; downstream cannot classify it", err, test.want)
			}
			message := err.Error()
			if strings.Contains(message, leakedCredential) {
				t.Fatalf("transport error leaked the credential: %s", message)
			}
			if strings.Contains(message, "crtfc_key") || strings.Contains(message, "opendart.fss.or.kr") {
				t.Fatalf("transport error leaked the authenticated URL: %s", message)
			}
		})
	}
}

// TestTransportErrorRedactionKeepsTheClassification 은 위 시험이 "메시지를 통째로
// 지웠다"로 통과하지 않게 한다. 지운 뒤에도 무엇이 실패했는지는 남아야 한다.
func TestTransportErrorRedactionKeepsTheClassification(t *testing.T) {
	t.Parallel()
	transport := &leakingTransport{err: credentialBearingURLError(context.DeadlineExceeded)}
	adapter := newAdapter(validDARTSourcePolicy(), transport, StaticCredential(leakedCredential))
	_, err := adapter.Fetch(context.Background(), FetchRequest{OperationID: "classify"})
	if !strings.Contains(err.Error(), "transport") {
		t.Fatalf("redacted transport error no longer says what failed: %v", err)
	}
}

func TestSECHistoricalPageResourceIsNotChosenByTheRemoteBody(t *testing.T) {
	t.Parallel()
	// 첫 페이지의 files[].name 은 원격 서버가 준 값이고, 그대로 다음 요청 경로가 된다.
	// 진짜 SEC 값은 "CIK<요청한 CIK>-submissions-NNN.json" 뿐이다.
	for _, spoofed := range []string{
		"../../../etc/passwd",
		"CIK9999999999-submissions-001.json",
		"CIK0000320193-submissions-001.json.evil",
		"https://attacker.example/CIK0000320193-submissions-001.json",
		"CIK0000320193-submissions-1.json",
	} {
		t.Run(spoofed, func(t *testing.T) {
			t.Parallel()
			body := fmt.Sprintf(`{"cik":"0000320193","name":"APPLE INC","tickers":["AAPL"],"exchanges":["Nasdaq"],`+
				`"filings":{"recent":{"accessionNumber":["0000320193-26-000001"],"filingDate":["2026-08-01"],`+
				`"reportDate":["2026-07-31"],"acceptanceDateTime":["2026-08-01T16:30:00.000Z"],"form":["8-K"],`+
				`"primaryDocument":["aapl.htm"]},"files":[{"name":%q,"filingCount":1,"filingFrom":"2025-01-01","filingTo":"2025-12-31"}]}}`, spoofed)
			transport := &scriptedOfficialTransport{responses: []TransportResponse{
				{Status: 200, Body: []byte(body), ObservedAt: time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)},
				{Status: 200, Body: fixture(t, "sec_submissions_v1_page2.json"), ObservedAt: time.Date(2026, 8, 3, 12, 0, 1, 0, time.UTC)},
			}}
			adapter := mustOfficialAdapter(t, validSECConfig(), transport, nil, NewSharedRateBudget())
			_, err := adapter.Collect(context.Background(), OfficialCollectionRequest{OperationID: "sec-spoof", EntityID: "0000320193"})
			if !errors.Is(err, ErrSourceSchemaDrift) {
				t.Fatalf("spoofed historical page name %q produced err=%v", spoofed, err)
			}
			if calls := transport.Calls(); calls != 1 {
				t.Fatalf("spoofed page name reached the transport: calls=%d requests=%+v", calls, transport.Requests())
			}
		})
	}
}

// TestSECHistoricalPageResourceAcceptsTheFrozenFixtureShape 은 위 거부 규칙이 실제
// SEC 응답까지 같이 막지 않는다는 양성 대조군이다. 이것이 없으면 "전부 거부"가
// 안전 시험을 모두 통과한다.
func TestSECHistoricalPageResourceAcceptsTheFrozenFixtureShape(t *testing.T) {
	t.Parallel()
	observed := time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)
	transport := &scriptedOfficialTransport{responses: []TransportResponse{
		{Status: 200, Body: fixture(t, "sec_submissions_v1_page1.json"), ObservedAt: observed},
		{Status: 200, Body: fixture(t, "sec_submissions_v1_page2.json"), ObservedAt: observed.Add(time.Second)},
	}}
	adapter := mustOfficialAdapter(t, validSECConfig(), transport, nil, NewSharedRateBudget())
	batch, err := adapter.Collect(context.Background(), OfficialCollectionRequest{OperationID: "sec-ok", EntityID: "0000320193"})
	if err != nil || !batch.Complete || batch.Pages != 2 {
		t.Fatalf("frozen SEC fixture refused: batch=%+v err=%v", batch, err)
	}
}

func TestSharedRateBudgetCapsEveryAdapterOnOneContract(t *testing.T) {
	t.Parallel()
	budget := NewSharedRateBudget()
	config := validSECConfig()
	config.MaxCalls = 1
	first := mustOfficialAdapter(t, config, &scriptedOfficialTransport{responses: []TransportResponse{
		{Status: 200, Body: singlePageSECFixture(), ObservedAt: time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)},
	}}, nil, budget)
	secondTransport := &scriptedOfficialTransport{responses: []TransportResponse{
		{Status: 200, Body: singlePageSECFixture(), ObservedAt: time.Date(2026, 8, 3, 12, 0, 0, 0, time.UTC)},
	}}
	second := mustOfficialAdapter(t, config, secondTransport, nil, budget)
	if _, err := first.Collect(context.Background(), OfficialCollectionRequest{OperationID: "one", EntityID: "0000320193"}); err != nil {
		t.Fatal(err)
	}
	if _, err := second.Collect(context.Background(), OfficialCollectionRequest{OperationID: "two", EntityID: "0000320193"}); !errors.Is(err, ErrSourceRateLimited) {
		t.Fatalf("second adapter on the same contract got its own budget: err=%v", err)
	}
	if secondTransport.Calls() != 0 {
		t.Fatalf("rate-limited adapter still called the transport %d times", secondTransport.Calls())
	}
}

// TestEveryPolicyFieldHasAZeroCallRefusal 은 정책 필드 **하나하나**가 거부를 만드는지
// 잰다. 봉인(contractSeal)은 변형 뒤 다시 찍는다 — 그러지 않으면 봉인 검사가 모든
// 변형을 대신 거부해서, 필드 검사를 지워도 시험이 초록으로 남는다(a064 의 원래 결함).
func TestEveryPolicyFieldHasAZeroCallRefusal(t *testing.T) {
	t.Parallel()
	type mutation struct {
		mutate func(*SourcePolicy)
		reseal bool
		want   error
	}
	field := func(mutate func(*SourcePolicy)) mutation {
		return mutation{mutate: mutate, reseal: true, want: ErrSourceDisabled}
	}
	mutations := map[string][]mutation{
		"Version":            {field(func(p *SourcePolicy) { p.Version = "" })},
		"ContractID":         {field(func(p *SourcePolicy) { p.ContractID = "sec-submissions-2099-01-01" })},
		"Authority":          {{mutate: func(p *SourcePolicy) { p.Authority = AuthorityTossOpenAPI }, reseal: true, want: ErrSourceUnavailable}},
		"ContractVerified":   {field(func(p *SourcePolicy) { p.ContractVerified = false })},
		"EndpointIdentity":   {field(func(p *SourcePolicy) { p.EndpointIdentity = "https://data.sec.example/submissions.json" })},
		"EndpointVersion":    {field(func(p *SourcePolicy) { p.EndpointVersion = "" })},
		"Method":             {field(func(p *SourcePolicy) { p.Method = "POST" })},
		"SchemaVersion":      {field(func(p *SourcePolicy) { p.SchemaVersion = "sec-submissions-v2" })},
		"AccessContract":     {field(func(p *SourcePolicy) { p.AccessContract = "wts" })},
		"RequestIdentity":    {field(func(p *SourcePolicy) { p.RequestIdentity = "anonymous-bot" })},
		"CredentialRequired": {field(func(p *SourcePolicy) { p.CredentialRequired = true })},
		"AbsoluteCallWindow": {
			field(func(p *SourcePolicy) { p.AbsoluteCallWindow = 0 }),
			field(func(p *SourcePolicy) { p.AbsoluteCallWindow = 48 * time.Hour }),
		},
		"MaxCalls": {
			field(func(p *SourcePolicy) { p.MaxCalls = 0 }),
			field(func(p *SourcePolicy) { p.MaxCalls = 100_000 }),
		},
		"MaxPages": {
			field(func(p *SourcePolicy) { p.MaxPages = 0 }),
			field(func(p *SourcePolicy) { p.MaxPages = 1_000 }),
		},
		"PageSize": {
			field(func(p *SourcePolicy) { p.PageSize = 0 }),
			field(func(p *SourcePolicy) { p.PageSize = 101 }),
		},
		"MaxResponseBytes": {
			field(func(p *SourcePolicy) { p.MaxResponseBytes = 0 }),
			field(func(p *SourcePolicy) { p.MaxResponseBytes = 1 << 40 }),
		},
		"MaxConcurrency": {
			field(func(p *SourcePolicy) { p.MaxConcurrency = 0 }),
			field(func(p *SourcePolicy) { p.MaxConcurrency = 64 }),
		},
		"RequestDeadline": {
			field(func(p *SourcePolicy) { p.RequestDeadline = 0 }),
			// 상한만 어긴다: 2분은 1분 상한 밖이지만 5분 operation deadline 안이므로
			// 순서 검사가 대신 거부하지 못한다. 앞선 판본은 1시간을 썼고, 그때는
			// `RequestDeadline > OperationDeadline` 이 먼저 걸려서 상한을 지워도
			// 시험이 초록이었다(2026-09-07 독립 리뷰 S4).
			{mutate: func(p *SourcePolicy) {
				p.RequestDeadline = 2 * time.Minute
				p.OperationDeadline = 5 * time.Minute
			}, reseal: true, want: ErrSourceDisabled},
			// 순서만 어긴다: 둘 다 상한 안이고 요청 deadline 이 operation 보다 길다.
			{mutate: func(p *SourcePolicy) {
				p.RequestDeadline = 3 * time.Second
				p.OperationDeadline = 2 * time.Second
			}, reseal: true, want: ErrSourceDisabled},
		},
		"OperationDeadline": {
			field(func(p *SourcePolicy) { p.OperationDeadline = 0 }),
			field(func(p *SourcePolicy) { p.OperationDeadline = 24 * time.Hour }),
		},
		"RetryableStatuses": {field(func(p *SourcePolicy) { p.RetryableStatuses = map[int]struct{}{} })},
		"MaxRetries": {
			field(func(p *SourcePolicy) { p.MaxRetries = -1 }),
			field(func(p *SourcePolicy) { p.MaxRetries = 1_000 }),
		},
		"RetryAfterPolicy": {field(func(p *SourcePolicy) { p.RetryAfterPolicy = "IGNORE_RETRY_AFTER" })},
		// 봉인만은 다시 찍지 않는다. 봉인은 "밖에서 필드를 바꿔치기했다"를 잡는 검사다.
		"contractSeal": {{mutate: func(p *SourcePolicy) { p.contractSeal = "forged" }, reseal: false, want: ErrSourceDisabled}},
	}
	assertCoversEveryField(t, reflect.TypeOf(SourcePolicy{}), mutations, "SourcePolicy")

	for name, cases := range mutations {
		for index, test := range cases {
			t.Run(fmt.Sprintf("%s/%d", name, index), func(t *testing.T) {
				t.Parallel()
				transport := &fakeTransport{}
				policy := validSourcePolicy()
				test.mutate(&policy)
				if test.reseal {
					policy.contractSeal = policy.officialSeal()
				}
				adapter := newAdapter(policy, transport, StaticCredential("secret"))
				_, err := adapter.Fetch(context.Background(), FetchRequest{OperationID: "bounded"})
				if !errors.Is(err, test.want) {
					t.Fatalf("SourcePolicy.%s mutation was accepted: err=%v want=%v", name, err, test.want)
				}
				if transport.Calls() != 0 {
					t.Fatalf("SourcePolicy.%s mutation still reached the transport %d times", name, transport.Calls())
				}
			})
		}
	}
}

// TestMintedPoliciesSurviveTheirOwnBounds 는 위 상한이 실제로 쓰는 두 정책까지
// 막지는 않는다는 양성 대조군이다. 이것이 없으면 "전부 거부"가 통과한다.
func TestMintedPoliciesSurviveTheirOwnBounds(t *testing.T) {
	t.Parallel()
	for name, policy := range map[string]SourcePolicy{"sec": validSourcePolicy(), "opendart": validDARTSourcePolicy()} {
		if err := policy.validateOfficial(); err != nil {
			t.Fatalf("minted %s policy fails its own bounds: %v", name, err)
		}
	}
	for name, config := range map[string]SourcePolicyConfig{"sec": validSECConfig(), "opendart": validDARTConfig()} {
		if _, err := MintSourcePolicy(config); err != nil {
			t.Fatalf("official %s contract config no longer mints: %v", name, err)
		}
	}
}
