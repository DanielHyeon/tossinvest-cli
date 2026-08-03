package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/console"
	"github.com/JungHoonGhae/tossinvest-cli/internal/domain"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/ratebudget"
	"github.com/JungHoonGhae/tossinvest-cli/internal/runlock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/verifylive"
	"github.com/spf13/cobra"
)

type a061StockMetadataBroker struct {
	verifylive.Broker
	mu      sync.Mutex
	batches [][]string
	quotes  []domain.Quote
	failAt  int
}

func (b *a061StockMetadataBroker) Stocks(_ context.Context, symbols []string) ([]domain.Quote, error) {
	b.mu.Lock()
	b.batches = append(b.batches, append([]string(nil), symbols...))
	call := len(b.batches)
	b.mu.Unlock()
	if b.failAt == call {
		return nil, fmt.Errorf("metadata batch %d failed", call)
	}
	out := make([]domain.Quote, 0, len(symbols))
	if b.quotes != nil {
		return append(out, b.quotes...), nil
	}
	for _, symbol := range symbols {
		if symbol == "DUP" {
			out = append(out, domain.Quote{Symbol: symbol, Name: "US duplicate", MarketCode: "NASDAQ", Currency: "USD"})
			continue
		}
		out = append(out, domain.Quote{Symbol: symbol, Name: "종목 " + symbol, MarketCode: "KOSPI", Currency: "KRW"})
	}
	return out, nil
}

type a061NoMetadataBroker struct{ verifylive.Broker }

type a061PromotableBroker struct {
	verifylive.Broker
	mu       sync.Mutex
	failures int
	calls    int
}

func (b *a061PromotableBroker) Accounts(context.Context) ([]domain.Account, error) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.calls++
	if b.calls <= b.failures {
		return nil, errors.New("accounts unavailable")
	}
	return []domain.Account{{ID: "7", DisplayName: "123-45-678901"}}, nil
}

func (b *a061StockMetadataBroker) seenBatches() [][]string {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([][]string, len(b.batches))
	for i := range b.batches {
		out[i] = append([]string(nil), b.batches[i]...)
	}
	return out
}

func TestA061InstrumentMetadataAdapterChunksOfficialReadsAtTwoHundred(t *testing.T) {
	broker := &a061StockMetadataBroker{}
	refs := make([]console.InstrumentRef, 0, 201)
	for i := 0; i < 201; i++ {
		refs = append(refs, console.InstrumentRef{Market: "kr", Symbol: fmt.Sprintf("%06d", i)})
	}

	rows, err := newConsoleInstrumentNamesWithBudget(&consoleBroker{client: broker}, "").Names(context.Background(), refs)
	if err != nil {
		t.Fatalf("Names: %v", err)
	}
	if len(rows) != len(refs) {
		t.Fatalf("Names returned %d rows, want %d", len(rows), len(refs))
	}
	batches := broker.seenBatches()
	if len(batches) != 2 || len(batches[0]) != 200 || len(batches[1]) != 1 {
		t.Fatalf("official stock batches = %+v, want sizes 200 and 1", batches)
	}
}

func TestA061InstrumentMetadataAdapterDoesNotCrossAttachMarkets(t *testing.T) {
	broker := &a061StockMetadataBroker{}
	rows, err := newConsoleInstrumentNamesWithBudget(&consoleBroker{client: broker}, "").Names(context.Background(), []console.InstrumentRef{
		{Market: "kr", Symbol: "DUP"},
		{Market: "us", Symbol: "DUP"},
	})
	if err != nil {
		t.Fatalf("Names: %v", err)
	}
	if len(rows) != 1 || rows[0].Market != "us" || rows[0].Name != "US duplicate" {
		t.Fatalf("market-keyed metadata = %+v, want only the US result", rows)
	}
}

func TestA061InstrumentMetadataRejectsContradictoryMarketSignals(t *testing.T) {
	broker := &a061StockMetadataBroker{quotes: []domain.Quote{
		{Symbol: "KRCONFLICT", Name: "wrong KR", MarketCode: "KOSPI", Currency: "USD"},
		{Symbol: "USCONFLICT", Name: "wrong US", MarketCode: "NASDAQ", Currency: "KRW"},
	}}
	rows, err := newConsoleInstrumentNamesWithBudget(&consoleBroker{client: broker}, "").Names(context.Background(), []console.InstrumentRef{
		{Market: "kr", Symbol: "KRCONFLICT"},
		{Market: "us", Symbol: "USCONFLICT"},
	})
	if err != nil {
		t.Fatalf("Names: %v", err)
	}
	if len(rows) != 0 {
		t.Fatalf("contradictory market metadata attached names: %+v", rows)
	}
}

func TestA061InstrumentMetadataReturnsFirstAndLaterChunkErrors(t *testing.T) {
	for _, failAt := range []int{1, 2} {
		t.Run(fmt.Sprintf("batch-%d", failAt), func(t *testing.T) {
			broker := &a061StockMetadataBroker{failAt: failAt}
			refs := make([]console.InstrumentRef, 201)
			for i := range refs {
				refs[i] = console.InstrumentRef{Market: "kr", Symbol: fmt.Sprintf("%06d", i)}
			}
			rows, err := newConsoleInstrumentNamesWithBudget(&consoleBroker{client: broker}, "").Names(context.Background(), refs)
			if err == nil || !strings.Contains(err.Error(), fmt.Sprintf("batch %d", failAt)) {
				t.Fatalf("Names error = %v, want batch %d failure", err, failAt)
			}
			if rows != nil {
				t.Fatalf("failed batch returned partial rows: %+v", rows)
			}
		})
	}
}

func TestA061InstrumentMetadataYieldsToTheCrossProcessVerificationBudget(t *testing.T) {
	budgetPath := filepath.Join(t.TempDir(), ratebudget.FileName)
	lease, ok, err := ratebudget.TryAcquire(budgetPath)
	if err != nil || !ok {
		t.Fatalf("hold rate budget = %v/%v", ok, err)
	}
	defer lease.Release()
	broker := &a061StockMetadataBroker{}
	_, err = newConsoleInstrumentNamesWithBudget(&consoleBroker{client: broker}, budgetPath).Names(
		context.Background(), []console.InstrumentRef{{Market: "kr", Symbol: "005930"}})
	if err == nil || !strings.Contains(err.Error(), "live verification owns") {
		t.Fatalf("contended metadata error = %v", err)
	}
	if len(broker.seenBatches()) != 0 {
		t.Fatal("metadata called /stocks while verification owned the cross-process budget")
	}
}

func TestA061InstrumentMetadataRechecksTheRunMarkerAfterTakingTheBudget(t *testing.T) {
	budgetPath := filepath.Join(t.TempDir(), ratebudget.FileName)
	runLockPath := filepath.Join(filepath.Dir(budgetPath), runlock.FileName)
	marker, err := runlock.Acquire(runLockPath, time.Now())
	if err != nil {
		t.Fatalf("hold verification marker: %v", err)
	}
	defer marker.Release()
	broker := &a061StockMetadataBroker{}
	_, err = newConsoleInstrumentNamesWithBudget(&consoleBroker{client: broker}, budgetPath).Names(
		context.Background(), []console.InstrumentRef{{Market: "kr", Symbol: "005930"}})
	if err == nil || !strings.Contains(err.Error(), "started before the metadata lease") {
		t.Fatalf("post-lease marker error = %v", err)
	}
	if len(broker.seenBatches()) != 0 {
		t.Fatal("metadata called /stocks after the post-lease verification recheck")
	}
}

func TestA061VerificationWaitsForAndThenOwnsTheSameRateBudget(t *testing.T) {
	root := &rootOptions{configDir: t.TempDir()}
	budgetPath, err := verifyRateBudgetPath(root)
	if err != nil {
		t.Fatalf("resolve budget path: %v", err)
	}
	first, ok, err := ratebudget.TryAcquire(budgetPath)
	if err != nil || !ok {
		t.Fatalf("hold metadata budget = %v/%v", ok, err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	if _, err := acquireVerifyRateBudget(ctx, io.Discard, root); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("verification contended budget error = %v", err)
	}
	first.Release()
	lease, err := acquireVerifyRateBudget(context.Background(), io.Discard, root)
	if err != nil {
		t.Fatalf("verification budget after release: %v", err)
	}
	lease.Release()
}

func TestA061RecordOverrideCannotMoveTheProfileRateBudgetLock(t *testing.T) {
	root := &rootOptions{configDir: t.TempDir()}
	before, err := verifyRateBudgetPath(root)
	if err != nil {
		t.Fatalf("budget path: %v", err)
	}
	for _, record := range []string{"/tmp/override-a.jsonl", "/var/tmp/override-b.jsonl"} {
		resolved, err := resolveVerifyRecordFor(root, record, verifylive.MarketKR)
		if err != nil || resolved != record {
			t.Fatalf("record override = %q/%v", resolved, err)
		}
		after, err := verifyRateBudgetPath(root)
		if err != nil || after != before {
			t.Fatalf("record override moved budget lock from %q to %q (%v)", before, after, err)
		}
	}
}

func TestA061RunAndConsoleVerificationDoNotBuildBrokerWhileTheProfileBudgetIsOccupied(t *testing.T) {
	previousFactory := verifyBrokerFactory
	var brokerCalls atomic.Int32
	verifyBrokerFactory = func(*rootOptions) (verifylive.Broker, string, error) {
		brokerCalls.Add(1)
		return nil, "", errors.New("broker must not be built while the rate budget is occupied")
	}
	t.Cleanup(func() { verifyBrokerFactory = previousFactory })

	for _, tc := range []struct {
		name string
		run  func(context.Context, *rootOptions) error
	}{
		{
			name: "runVerifyRun",
			run: func(ctx context.Context, root *rootOptions) error {
				cmd := &cobra.Command{}
				cmd.SetContext(ctx)
				cmd.SetOut(io.Discard)
				cmd.SetErr(io.Discard)
				return runVerifyRun(cmd, root, &verifyOptions{market: verifylive.MarketKR})
			},
		},
		{
			name: "consoleVerifyStarter",
			run: func(ctx context.Context, root *rootOptions) error {
				_, _, err := consoleVerifyStarter(root)(ctx, nil, io.Discard, verifylive.MarketKR, nil)
				return err
			},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := &rootOptions{configDir: t.TempDir()}
			budgetPath, err := verifyRateBudgetPath(root)
			if err != nil {
				t.Fatalf("resolve budget path: %v", err)
			}
			lease, ok, err := ratebudget.TryAcquire(budgetPath)
			if err != nil || !ok {
				t.Fatalf("occupy profile budget = %v/%v", ok, err)
			}
			defer lease.Release()

			before := brokerCalls.Load()
			ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
			defer cancel()
			err = tc.run(ctx, root)
			if !errors.Is(err, context.DeadlineExceeded) {
				t.Fatalf("occupied profile budget error = %v, want deadline exceeded", err)
			}
			if got := brokerCalls.Load() - before; got != 0 {
				t.Fatalf("verifyBrokerFactory called %d time(s) while the profile budget was occupied", got)
			}
		})
	}
}

func TestA061VerificationRefusesBeforeBrokerWhenItsIntentCannotBePublished(t *testing.T) {
	root := &rootOptions{configDir: t.TempDir()}
	budgetPath, err := verifyRateBudgetPath(root)
	if err != nil {
		t.Fatalf("resolve budget path: %v", err)
	}
	lease, ok, err := ratebudget.TryAcquire(budgetPath)
	if err != nil || !ok {
		t.Fatalf("prepare rate-budget file = %v/%v", ok, err)
	}
	lease.Release()
	executionLock, err := acquireVerifyExecutionLock(root)
	if err != nil {
		t.Fatalf("prepare execution lock file: %v", err)
	}
	executionLock.Release()
	markerPath := filepath.Join(filepath.Dir(budgetPath), runlock.FileName)
	if err := os.WriteFile(markerPath, []byte("stale marker\n"), 0o400); err != nil {
		t.Fatalf("write read-only marker: %v", err)
	}
	if err := os.Chmod(markerPath, 0o400); err != nil {
		t.Fatalf("make marker read-only: %v", err)
	}

	previousFactory := verifyBrokerFactory
	var brokerCalls atomic.Int32
	verifyBrokerFactory = func(*rootOptions) (verifylive.Broker, string, error) {
		brokerCalls.Add(1)
		return nil, "", errors.New("unexpected broker construction")
	}
	t.Cleanup(func() { verifyBrokerFactory = previousFactory })

	cmd := &cobra.Command{}
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	err = runVerifyRun(cmd, root, &verifyOptions{market: verifylive.MarketKR})
	if err == nil || !strings.Contains(err.Error(), "publishing the Open API rate-budget intent") {
		t.Fatalf("unpublishable verification intent error = %v", err)
	}
	if brokerCalls.Load() != 0 {
		t.Fatal("verification built a broker without publishing its admission intent")
	}
}

func TestA061RunVerifyAbortPublishesItsMarkerBeforeWaitingForTheProfileBudget(t *testing.T) {
	root := &rootOptions{configDir: t.TempDir()}
	recordPath := filepath.Join(root.configDir, verifylive.FileName)
	seedVerifyRecord(t, recordPath)
	budgetPath, err := verifyRateBudgetPath(root)
	if err != nil {
		t.Fatalf("resolve budget path: %v", err)
	}
	lease, ok, err := ratebudget.TryAcquire(budgetPath)
	if err != nil || !ok {
		t.Fatalf("occupy profile budget = %v/%v", ok, err)
	}
	defer lease.Release()

	previousFactory := verifyBrokerFactory
	var brokerCalls atomic.Int32
	verifyBrokerFactory = func(*rootOptions) (verifylive.Broker, string, error) {
		brokerCalls.Add(1)
		return nil, "", errors.New("broker must not be built while the rate budget is occupied")
	}
	t.Cleanup(func() { verifyBrokerFactory = previousFactory })

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	cmd := &cobra.Command{}
	cmd.SetContext(ctx)
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	done := make(chan error, 1)
	go func() {
		done <- runVerifyAbort(cmd, root, &verifyOptions{market: verifylive.MarketKR})
	}()

	markerPath := verifyRunLockPath(recordPath)
	deadline := time.NewTimer(time.Second)
	ticker := time.NewTicker(5 * time.Millisecond)
	defer deadline.Stop()
	defer ticker.Stop()
	for {
		fresh, _ := runlock.Fresh(markerPath, time.Now(), runlock.StaleAfter)
		if fresh {
			break
		}
		select {
		case err := <-done:
			t.Fatalf("runVerifyAbort returned before publishing %s: %v", markerPath, err)
		case <-deadline.C:
			t.Fatalf("runVerifyAbort did not publish %s before waiting for the occupied budget", markerPath)
		case <-ticker.C:
		}
	}
	if got := brokerCalls.Load(); got != 0 {
		t.Fatalf("verifyBrokerFactory called %d time(s) while abort waited for the occupied budget", got)
	}
	select {
	case err := <-done:
		t.Fatalf("runVerifyAbort returned while the profile budget was still occupied: %v", err)
	default:
	}

	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("canceled runVerifyAbort error = %v, want context canceled", err)
		}
	case <-time.After(time.Second):
		t.Fatal("runVerifyAbort did not honor cancellation while waiting for the profile budget")
	}
	if got := brokerCalls.Load(); got != 0 {
		t.Fatalf("verifyBrokerFactory called %d time(s) before canceled abort returned", got)
	}
}

func TestA061AbortListAndRefreshedEmptyRecordNeverBuildABroker(t *testing.T) {
	previousFactory := verifyBrokerFactory
	var brokerCalls atomic.Int32
	verifyBrokerFactory = func(*rootOptions) (verifylive.Broker, string, error) {
		brokerCalls.Add(1)
		return nil, "", errors.New("unexpected broker construction")
	}
	t.Cleanup(func() { verifyBrokerFactory = previousFactory })

	t.Run("list", func(t *testing.T) {
		root := &rootOptions{configDir: t.TempDir()}
		seedVerifyRecord(t, filepath.Join(root.configDir, verifylive.FileName))
		var out strings.Builder
		cmd := &cobra.Command{}
		cmd.SetOut(&out)
		if err := runVerifyAbort(cmd, root, &verifyOptions{list: true, market: verifylive.MarketKR}); err != nil {
			t.Fatalf("abort --list: %v", err)
		}
		if !strings.Contains(out.String(), "co-1") || !strings.Contains(out.String(), "아무것도 전송되지 않았다") {
			t.Fatalf("abort --list output:\n%s", out.String())
		}
	})

	t.Run("empty-live-abort", func(t *testing.T) {
		root := &rootOptions{configDir: t.TempDir()}
		var out strings.Builder
		cmd := &cobra.Command{}
		cmd.SetOut(&out)
		if err := runVerifyAbort(cmd, root, &verifyOptions{market: verifylive.MarketKR}); err != nil {
			t.Fatalf("empty abort: %v", err)
		}
		if !strings.Contains(out.String(), "살아 있는 객체가 없다") {
			t.Fatalf("empty abort output:\n%s", out.String())
		}
	})

	if brokerCalls.Load() != 0 {
		t.Fatalf("read-only/empty abort built a broker %d time(s)", brokerCalls.Load())
	}
}

func TestA061ActiveVerificationExclusionRefusesAbortWithoutErasingItsMarker(t *testing.T) {
	root := &rootOptions{configDir: t.TempDir()}
	recordPath := filepath.Join(root.configDir, verifylive.FileName)
	seedVerifyRecord(t, recordPath)
	executionLock, err := acquireVerifyExecutionLock(root)
	if err != nil {
		t.Fatalf("hold active verification exclusion: %v", err)
	}
	defer executionLock.Release()
	marker, err := runlock.Acquire(verifyRunLockPath(recordPath), time.Now())
	if err != nil {
		t.Fatalf("hold active verification marker: %v", err)
	}
	defer marker.Release()

	previousFactory := verifyBrokerFactory
	var brokerCalls atomic.Int32
	verifyBrokerFactory = func(*rootOptions) (verifylive.Broker, string, error) {
		brokerCalls.Add(1)
		return nil, "", errors.New("unexpected broker construction")
	}
	t.Cleanup(func() { verifyBrokerFactory = previousFactory })

	cmd := &cobra.Command{}
	cmd.SetOut(io.Discard)
	cmd.SetErr(io.Discard)
	err = runVerifyAbort(cmd, root, &verifyOptions{market: verifylive.MarketKR})
	if err == nil || !strings.Contains(err.Error(), "execution exclusion") {
		t.Fatalf("abort during active verification error = %v", err)
	}
	if brokerCalls.Load() != 0 {
		t.Fatal("abort built a broker despite active verification exclusion")
	}
	if fresh, _ := runlock.Fresh(verifyRunLockPath(recordPath), time.Now(), runlock.StaleAfter); !fresh {
		t.Fatal("refused abort erased the active verification's marker")
	}
}

func TestA061AbortReloadsOutstandingTargetsAfterExclusiveAdmission(t *testing.T) {
	root := &rootOptions{configDir: t.TempDir()}
	recordPath := filepath.Join(root.configDir, verifylive.FileName)
	seedVerifyRecord(t, recordPath)
	budgetPath, err := verifyRateBudgetPath(root)
	if err != nil {
		t.Fatalf("resolve budget path: %v", err)
	}
	lease, ok, err := ratebudget.TryAcquire(budgetPath)
	if err != nil || !ok {
		t.Fatalf("occupy profile budget = %v/%v", ok, err)
	}

	previousFactory := verifyBrokerFactory
	verifyBrokerFactory = func(*rootOptions) (verifylive.Broker, string, error) {
		return nil, "", errors.New("stop after refreshed target projection")
	}
	t.Cleanup(func() { verifyBrokerFactory = previousFactory })

	var out strings.Builder
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	cmd := &cobra.Command{}
	cmd.SetContext(ctx)
	cmd.SetOut(&out)
	cmd.SetErr(io.Discard)
	done := make(chan error, 1)
	go func() {
		done <- runVerifyAbort(cmd, root, &verifyOptions{market: verifylive.MarketKR})
	}()

	markerPath := verifyRunLockPath(recordPath)
	deadline := time.Now().Add(time.Second)
	for {
		if fresh, _ := runlock.Fresh(markerPath, time.Now(), runlock.StaleAfter); fresh {
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("abort did not publish intent before waiting")
		}
		time.Sleep(5 * time.Millisecond)
	}
	recorder, err := verifylive.OpenRecorder(recordPath)
	if err != nil {
		t.Fatalf("open record during admission wait: %v", err)
	}
	err = recorder.Append(verifylive.Entry{
		StepID:  verifylive.StepConditionalTrigger,
		Verdict: verifylive.VerdictFail,
		Artifacts: []verifylive.Artifact{{
			Kind: verifylive.KindOrder, ID: "new-child", Symbol: "005930",
			CreatedAt: time.Now().UTC(), Deliberate: true,
		}},
	})
	if err == nil {
		err = recorder.Append(verifylive.Entry{
			StepID:  verifylive.StepAbort,
			Verdict: verifylive.VerdictPass,
			Artifacts: []verifylive.Artifact{{
				Kind: verifylive.KindConditional, ID: "co-1", Symbol: "005930",
				Cancelled: true, CancelledAt: time.Now().UTC(),
			}},
		})
	}
	_ = recorder.Close()
	if err != nil {
		t.Fatalf("append target during admission wait: %v", err)
	}
	lease.Release()
	if err := <-done; err == nil || !strings.Contains(err.Error(), "refreshed target projection") {
		t.Fatalf("abort after refreshed projection error = %v", err)
	}
	if !strings.Contains(out.String(), "new-child") {
		t.Fatalf("abort used stale pre-admission targets:\n%s", out.String())
	}
	if strings.Contains(out.String(), "co-1") {
		t.Fatalf("abort retained a target settled during admission wait:\n%s", out.String())
	}
}

func TestA061ColdMetadataResolverHonorsRequestCancellation(t *testing.T) {
	shared := &consoleBroker{build: func(ctx context.Context, _ *rootOptions) (verifylive.Broker, string, error) {
		<-ctx.Done()
		return nil, "", ctx.Err()
	}}
	seam := newConsoleInstrumentNamesWithBudget(shared, "")
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	started := time.Now()
	_, err := seam.Names(ctx, []console.InstrumentRef{{Market: "kr", Symbol: "005930"}})
	if !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("cold resolver error = %v, want deadline exceeded", err)
	}
	if elapsed := time.Since(started); elapsed > time.Second {
		t.Fatalf("cold resolver ignored cancellation for %s", elapsed)
	}
}

func TestA061MetadataResolverHonorsCancellationWhileAccountResolutionOwnsTheBroker(t *testing.T) {
	shared := &consoleBroker{}
	if err := shared.lock(context.Background()); err != nil {
		t.Fatalf("hold broker: %v", err)
	}
	defer shared.unlock()
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()
	started := time.Now()
	_, err := shared.instrumentMetadata(ctx)
	if !errors.Is(err, context.DeadlineExceeded) || time.Since(started) > time.Second {
		t.Fatalf("contended metadata resolver error/elapsed = %v/%s", err, time.Since(started))
	}
}

func TestA061MetadataAndAccountReadsShareOneOfficialClient(t *testing.T) {
	var tokenCalls, stockCalls, accountCalls, builds atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			tokenCalls.Add(1)
			_, _ = w.Write([]byte(`{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`))
		case "/api/v1/stocks":
			stockCalls.Add(1)
			_, _ = w.Write([]byte(`{"result":[{"symbol":"005930","name":"삼성전자","market":"KOSPI","currency":"KRW","status":"ACTIVE"}]}`))
		case "/api/v1/accounts":
			accountCalls.Add(1)
			_, _ = w.Write([]byte(`{"result":[{"accountNo":"123-45-678901","accountSeq":7,"accountType":"BROKERAGE"}]}`))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)

	shared := &consoleBroker{root: &rootOptions{}, build: func(ctx context.Context, _ *rootOptions) (verifylive.Broker, string, error) {
		builds.Add(1)
		client := official.New(
			official.Credentials{APIKey: "k", SecretKey: "s"},
			filepath.Join(t.TempDir(), "token.json"),
			official.WithBaseURL(srv.URL),
			official.WithHTTPClient(srv.Client()),
		)
		ref, _, err := resolveVerifyAccount(ctx, client, func(context.Context, time.Duration) error { return nil })
		return client, ref, err
	}}
	if _, err := shared.instrumentMetadata(context.Background()); err != nil {
		t.Fatalf("prebuild metadata client: %v", err)
	}
	previous := verifyBrokerFactory
	var factoryCalls atomic.Int32
	verifyBrokerFactory = func(*rootOptions) (verifylive.Broker, string, error) {
		factoryCalls.Add(1)
		return nil, "", errors.New("unexpected second client")
	}
	t.Cleanup(func() { verifyBrokerFactory = previous })

	var wg sync.WaitGroup
	wg.Add(2)
	errs := make(chan error, 2)
	go func() {
		defer wg.Done()
		_, err := newConsoleInstrumentNamesWithBudget(shared, "").Names(context.Background(), []console.InstrumentRef{{Market: "kr", Symbol: "005930"}})
		errs <- err
	}()
	go func() {
		defer wg.Done()
		_, _, err := shared.resolve()
		errs <- err
	}()
	wg.Wait()
	close(errs)
	for err := range errs {
		if err != nil {
			t.Fatalf("concurrent metadata/account read: %v", err)
		}
	}
	if got := builds.Load(); got != 1 {
		t.Fatalf("official clients built = %d, want 1", got)
	}
	if got := factoryCalls.Load(); got != 0 {
		t.Fatalf("account factory built %d second client(s), want 0", got)
	}
	if got := tokenCalls.Load(); got != 1 {
		t.Fatalf("OAuth token exchanges = %d, want one shared token manager", got)
	}
	if stockCalls.Load() != 1 || accountCalls.Load() != 1 {
		t.Fatalf("stock/account calls = %d/%d, want 1/1", stockCalls.Load(), accountCalls.Load())
	}
}

func TestA061InstrumentMetadataCapabilityFailuresAreExplicit(t *testing.T) {
	if _, err := (&consoleBroker{}).instrumentMetadata(context.Background()); err == nil || !strings.Contains(err.Error(), "builder is not configured") {
		t.Fatalf("missing builder error = %v", err)
	}
	broker := &a061NoMetadataBroker{}
	shared := &consoleBroker{build: func(context.Context, *rootOptions) (verifylive.Broker, string, error) {
		return broker, "123-45-678901", nil
	}}
	if _, err := shared.instrumentMetadata(context.Background()); err == nil || !strings.Contains(err.Error(), "no official stock metadata read") {
		t.Fatalf("missing Stocks error = %v", err)
	}
	if shared.client != broker {
		t.Fatal("metadata capability failure did not retain the one shared client")
	}
}

func TestA061MetadataAdapterIsNarrowAndDelegatesClientLifecycleToTheSharedResolver(t *testing.T) {
	typ := reflect.TypeOf((*lazyInstrumentNames)(nil))
	if typ.NumMethod() != 1 || typ.Method(0).Name != "Names" {
		t.Fatalf("metadata adapter exported methods = %d/%s, want only Names", typ.NumMethod(), typ.Method(0).Name)
	}
	src := readSource(t, "instrument_names.go")
	if strings.Contains(src, ".Accounts(") || strings.Contains(src, "verifyBrokerFactory(") {
		t.Fatal("instrument metadata adapter bypasses the shared broker lifecycle")
	}
}
