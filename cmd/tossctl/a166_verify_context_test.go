package main

import (
	"context"
	"errors"
	"fmt"
	"github.com/JungHoonGhae/tossinvest-cli/internal/accountbudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
	"go/ast"
	"go/parser"
	"go/token"
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestA166VerifyFactoryUsesCallerContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, _, err := buildVerifyBroker(ctx, &rootOptions{})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("caller cancellation lost: %v", err)
	}
}

func TestA166ConsoleBrokerHonorsLifetimeEvenCached(t *testing.T) {
	lifetime, cancelLifetime := context.WithCancel(context.Background())
	prior := verifyBrokerFactory
	verifyBrokerFactory = func(context.Context, *rootOptions) (verifylive.Broker, string, error) {
		return official.New(official.Credentials{}, ""), "fake", nil
	}
	t.Cleanup(func() { verifyBrokerFactory = prior })
	broker := newConsoleBroker(&rootOptions{}, lifetime)
	if _, _, err := broker.resolve(context.Background()); err != nil {
		t.Fatal(err)
	}
	cancelLifetime()
	if _, _, err := broker.resolve(context.Background()); !errors.Is(err, context.Canceled) {
		t.Fatalf("lifetime cancellation lost: %v", err)
	}
}

func TestA166HTTPAPIAccountRefGetsRequestContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	seen := false
	r := &httpAPIReader{holdings: &httpAPIHoldingsFixture{}, accountRef: func(ctx context.Context) (string, error) { seen = true; return "", ctx.Err() }}
	_, err := r.readPositionBrokerSnapshot(ctx)
	if !seen || !errors.Is(err, context.Canceled) {
		t.Fatalf("request context lost: seen=%v err=%v", seen, err)
	}
}

func TestA166HTTPAPIShutdownTimeoutClosesConnection(t *testing.T) {
	started := make(chan struct{})
	release := make(chan struct{})
	defer close(release)
	server := &http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { close(started); <-release })}
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	served := make(chan error, 1)
	go func() { served <- server.Serve(ln) }()
	go func() {
		resp, err := http.Get("http://" + ln.Addr().String())
		if err == nil {
			resp.Body.Close()
		}
	}()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("handler did not start")
	}
	sentinel := errors.New("fake server error")
	err = finishHTTPAPIServer(server, served, sentinel, true, 30*time.Millisecond)
	if !errors.Is(err, context.DeadlineExceeded) || !errors.Is(err, sentinel) {
		t.Fatalf("shutdown timeout lost: %v", err)
	}
}

func TestA166VerifyFactoryBudgetLocalTLS(t *testing.T) {
	dir := isolate(t)
	writeCredentials(t, dir)
	var calls atomic.Int32
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		calls.Add(1)
		switch r.URL.Path {
		case "/oauth2/token":
			fmt.Fprint(w, `{"access_token":"fake","expires_in":3600,"token_type":"Bearer"}`)
		case "/api/v1/accounts":
			fmt.Fprint(w, `{"result":[{"accountNo":"fake","accountSeq":17,"accountType":"BROKERAGE"}]}`)
		default:
			t.Errorf("unexpected path %s", r.URL.Path)
			http.Error(w, "unexpected", 500)
		}
	}))
	defer server.Close()
	tr := server.Client().Transport.(*http.Transport).Clone()
	tr.Proxy = nil
	opts := []official.Option{official.WithBaseURL(server.URL), official.WithHTTPClient(&http.Client{Transport: tr})}
	owner := a166LocalRequiredOwner(t)
	factory, err := owner.get(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	budget, err := factory.Budget()
	if err != nil {
		t.Fatal(err)
	}
	lease, err := budget.Acquire(context.Background(), accountbudget.Critical)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Millisecond)
	_, _, err = buildVerifyBroker(ctx, &rootOptions{configDir: dir, budgetOwner: owner, officialOptions: opts})
	cancel()
	if err == nil || calls.Load() != 0 {
		t.Fatalf("first client bypassed budget: err=%v calls=%d", err, calls.Load())
	}
	if err = lease.Release(); err != nil {
		t.Fatal(err)
	}
	broker, _, err := buildVerifyBroker(context.Background(), &rootOptions{configDir: dir, budgetOwner: owner, officialOptions: opts})
	if err != nil {
		t.Fatal(err)
	}
	if !broker.(*official.Client).UsesAccountBudget(budget) {
		t.Fatal("seq-bound client lost budget")
	}
	if calls.Load() != 2 {
		t.Fatalf("local calls=%d", calls.Load())
	}
	again, _, err := buildVerifyBroker(context.Background(), &rootOptions{configDir: dir, budgetOwner: owner, officialOptions: opts})
	if err != nil {
		t.Fatal(err)
	}
	if again == broker || !again.(*official.Client).UsesAccountBudget(budget) {
		t.Fatal("each run needs a new client with the shared budget")
	}
}

func TestA166VerifyFactoryLocalOptionsGuard(t *testing.T) {
	f, err := parser.ParseFile(token.NewFileSet(), "verify.go", nil, 0)
	if err != nil {
		t.Fatal(err)
	}
	found := false
	for _, d := range f.Decls {
		fn, ok := d.(*ast.FuncDecl)
		if !ok || fn.Name.Name != "buildVerifyBroker" {
			continue
		}
		ast.Inspect(fn.Body, func(n ast.Node) bool {
			c, ok := n.(*ast.CallExpr)
			if !ok || len(c.Args) != 5 || !c.Ellipsis.IsValid() {
				return true
			}
			id, ok := c.Fun.(*ast.Ident)
			if !ok || id.Name != "newRootOfficialClient" {
				return true
			}
			s, ok := c.Args[4].(*ast.SelectorExpr)
			if !ok || s.Sel.Name != "officialOptions" {
				return true
			}
			v, ok := s.X.(*ast.Ident)
			if ok && v.Name == "root" {
				found = true
			}
			return true
		})
	}
	if !found {
		t.Fatal("fixture options are not bound to verify client")
	}
}

func TestA166ConsoleBrokerRetriesFailureAndCachesSuccess(t *testing.T) {
	prior := verifyBrokerFactory
	attempts := 0
	verifyBrokerFactory = func(context.Context, *rootOptions) (verifylive.Broker, string, error) {
		attempts++
		if attempts == 1 {
			return nil, "", errors.New("fake failure")
		}
		return official.New(official.Credentials{}, ""), "fake", nil
	}
	t.Cleanup(func() { verifyBrokerFactory = prior })
	broker := newConsoleBroker(&rootOptions{})
	if _, _, err := broker.resolve(context.Background()); err == nil {
		t.Fatal("first failure hidden")
	}
	if _, _, err := broker.resolve(context.Background()); err != nil {
		t.Fatal(err)
	}
	if _, _, err := broker.resolve(context.Background()); err != nil || attempts != 2 {
		t.Fatalf("cache: attempts=%d err=%v", attempts, err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, _, err := broker.resolve(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("request cancellation lost: %v", err)
	}
}

func TestA166ConsoleBrokerConcurrentSingleSuccess(t *testing.T) {
	prior := verifyBrokerFactory
	var builds atomic.Int32
	verifyBrokerFactory = func(context.Context, *rootOptions) (verifylive.Broker, string, error) {
		builds.Add(1)
		return official.New(official.Credentials{}, ""), "fake", nil
	}
	t.Cleanup(func() { verifyBrokerFactory = prior })
	broker := newConsoleBroker(&rootOptions{})
	var wg sync.WaitGroup
	errs := make(chan error, 8)
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); _, _, err := broker.resolve(context.Background()); errs <- err }()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatal(err)
		}
	}
	if builds.Load() != 1 {
		t.Fatalf("build count=%d", builds.Load())
	}
}

func TestA166ConsoleBrokerLifetimeCancelsPending(t *testing.T) {
	lifetime, cancel := context.WithCancel(context.Background())
	defer cancel()
	prior := verifyBrokerFactory
	started := make(chan struct{})
	verifyBrokerFactory = func(ctx context.Context, _ *rootOptions) (verifylive.Broker, string, error) {
		close(started)
		<-ctx.Done()
		return nil, "", ctx.Err()
	}
	t.Cleanup(func() { verifyBrokerFactory = prior })
	broker := newConsoleBroker(&rootOptions{}, lifetime)
	result := make(chan error, 1)
	go func() { _, _, err := broker.resolve(context.Background()); result <- err }()
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("factory did not start")
	}
	cancel()
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("lifetime error=%v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("pending factory did not cancel")
	}
}

type a166BlockingCalendar struct {
	*official.Client
	started chan struct{}
	release chan struct{}
	calls   atomic.Int32
}

func (b *a166BlockingCalendar) TypedMarketCalendar(ctx context.Context, _, _ string) (official.MarketCalendarResponse, error) {
	b.calls.Add(1)
	close(b.started)
	select {
	case <-ctx.Done():
		return official.MarketCalendarResponse{}, ctx.Err()
	case <-b.release:
		return official.MarketCalendarResponse{}, errors.New("fixture released")
	}
}
func TestA166ReadLifetimeAfterCachedBootstrap(t *testing.T) {
	lifetime, cancel := context.WithCancel(context.Background())
	defer cancel()
	fixture := &a166BlockingCalendar{Client: official.New(official.Credentials{}, ""), started: make(chan struct{}), release: make(chan struct{})}
	defer close(fixture.release)
	prior := verifyBrokerFactory
	verifyBrokerFactory = func(context.Context, *rootOptions) (verifylive.Broker, string, error) { return fixture, "fake", nil }
	t.Cleanup(func() { verifyBrokerFactory = prior })
	broker := newConsoleBroker(&rootOptions{}, lifetime)
	if _, _, err := broker.resolve(context.Background()); err != nil {
		t.Fatal(err)
	}
	ctx, requestCancel := context.WithCancel(context.Background())
	requestCancel()
	if _, err := broker.TypedMarketCalendar(ctx, "KR", "20261031"); !errors.Is(err, context.Canceled) || fixture.calls.Load() != 0 {
		t.Fatalf("pre-cancel err=%v calls=%d", err, fixture.calls.Load())
	}
	result := make(chan error, 1)
	go func() { _, err := broker.TypedMarketCalendar(context.Background(), "KR", "20261031"); result <- err }()
	select {
	case <-fixture.started:
	case <-time.After(time.Second):
		t.Fatal("read did not start")
	}
	cancel()
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("lifetime cancellation lost: %v", err)
		}
	case <-time.After(200 * time.Millisecond):
		t.Fatal("read outlived signal lifetime")
	}
}

type a166BlockingReads struct {
	*official.Client
	started  chan struct{}
	release  chan struct{}
	once     sync.Once
	observed chan error
}

func (b *a166BlockingReads) wait(ctx context.Context) error {
	b.once.Do(func() { close(b.started) })
	select {
	case <-ctx.Done():
		err := ctx.Err()
		select {
		case b.observed <- err:
		default:
		}
		return err
	case <-b.release:
		return errors.New("fixture released")
	}
}
func (b *a166BlockingReads) Holdings(ctx context.Context, _ string) ([]domain.Position, error) {
	return nil, b.wait(ctx)
}
func (b *a166BlockingReads) OrdersRaw(ctx context.Context, _ official.OrdersFilter) (official.RawOrderList, error) {
	return official.RawOrderList{}, b.wait(ctx)
}
func (b *a166BlockingReads) ConditionalOrdersRaw(ctx context.Context, _, _, _ string, _ int) (official.RawConditionalOrderList, error) {
	return official.RawConditionalOrderList{}, b.wait(ctx)
}
func (b *a166BlockingReads) Stocks(ctx context.Context, _ []string) ([]domain.Quote, error) {
	return nil, b.wait(ctx)
}

func TestA166ReadLifetimeForHoldingsOrdersAndNames(t *testing.T) {
	for _, kind := range []string{"holdings", "orders", "names"} {
		t.Run(kind, func(t *testing.T) {
			lifetime, cancel := context.WithCancel(context.Background())
			defer cancel()
			fixture := &a166BlockingReads{Client: official.New(official.Credentials{}, ""), started: make(chan struct{}), release: make(chan struct{}), observed: make(chan error, 4)}
			defer close(fixture.release)
			prior := verifyBrokerFactory
			verifyBrokerFactory = func(context.Context, *rootOptions) (verifylive.Broker, string, error) { return fixture, "fake", nil }
			t.Cleanup(func() { verifyBrokerFactory = prior })
			broker := newConsoleBroker(&rootOptions{}, lifetime)
			if _, _, err := broker.resolve(context.Background()); err != nil {
				t.Fatal(err)
			}
			done := make(chan struct{})
			go func() {
				defer close(done)
				switch kind {
				case "holdings":
					_, _ = newConsoleHoldings(broker).Holdings(context.Background(), "")
				case "orders":
					_, _ = consoleOrdersSeam(broker).Orders(context.Background())
				case "names":
					_, _ = newConsoleInstrumentNamesWithBudget(broker, "").Names(context.Background(), []consoleInstrumentRef{{Market: "kr", Symbol: "005930"}})
				}
			}()
			select {
			case <-fixture.started:
			case <-time.After(time.Second):
				t.Fatal("read did not start")
			}
			cancel()
			select {
			case err := <-fixture.observed:
				if !errors.Is(err, context.Canceled) {
					t.Fatalf("read ctx=%v", err)
				}
			case <-time.After(time.Second):
				t.Fatal("signal did not cancel read")
			}
			select {
			case <-done:
			case <-time.After(time.Second):
				t.Fatal("read did not finish")
			}
		})
	}
}

func TestA166ReadOriginalRequestCancellation(t *testing.T) {
	fixture := &a166BlockingCalendar{Client: official.New(official.Credentials{}, ""), started: make(chan struct{}), release: make(chan struct{})}
	defer close(fixture.release)
	prior := verifyBrokerFactory
	verifyBrokerFactory = func(context.Context, *rootOptions) (verifylive.Broker, string, error) { return fixture, "fake", nil }
	t.Cleanup(func() { verifyBrokerFactory = prior })
	broker := newConsoleBroker(&rootOptions{}, context.Background())
	if _, _, err := broker.resolve(context.Background()); err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	result := make(chan error, 1)
	go func() { _, err := broker.TypedMarketCalendar(ctx, "KR", "20261031"); result <- err }()
	select {
	case <-fixture.started:
	case <-time.After(time.Second):
		t.Fatal("read did not start")
	}
	cancel()
	select {
	case err := <-result:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("request cancellation lost: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("request did not cancel")
	}
}
