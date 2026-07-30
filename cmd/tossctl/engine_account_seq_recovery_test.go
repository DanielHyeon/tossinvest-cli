//go:build tossos_testseams

package main

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/app/engine"
	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/config"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
	"github.com/JungHoonGhae/tossinvest-cli/internal/reconcile"
)

func assembleAccountSequenceTestEngine(
	t *testing.T,
	handler http.Handler,
	officialOptions ...official.Option,
) (*engine.Context, string, *httptest.Server, error) {
	t.Helper()
	dir := isolate(t)
	writeConfigFile(t, dir, config.DefaultFile())
	writeCredentials(t, dir)
	srv := httptest.NewServer(handler)
	options := []official.Option{
		official.WithBaseURL(srv.URL),
		official.WithHTTPClient(srv.Client()),
	}
	options = append(options, officialOptions...)
	ectx, err := assembleEngine(
		context.Background(),
		&rootOptions{configDir: dir},
		clock.System(),
		obs.NewLogger(obs.LogOptions{Writer: io.Discard}),
		func(opts *engine.Options) {
			engine.ConfigureCLIRegressionForTest(
				opts,
				journal.FixedFSProber(journal.FSInfo{Name: "ext4", Magic: journal.MagicExt}),
				nil,
			)
		},
		options...,
	)
	return ectx, dir, srv, err
}

func writeEngineAccountSequenceToken(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(w, `{"access_token":"AT","expires_in":3600,"token_type":"Bearer"}`)
}

func writeEngineAccountSequenceAccounts(w http.ResponseWriter, body string) {
	w.Header().Set("Content-Type", "application/json")
	_, _ = io.WriteString(w, `{"result":`+body+`}`)
}

func writeEmptyEngineSnapshotResponse(w http.ResponseWriter, path string) {
	w.Header().Set("Content-Type", "application/json")
	switch path {
	case "/api/v1/orders":
		_, _ = io.WriteString(w, `{"result":{"orders":[],"nextCursor":"","hasNext":false}}`)
	case "/api/v1/holdings":
		_, _ = io.WriteString(w, `{"result":{"items":[]}}`)
	case "/api/v1/buying-power":
		_, _ = io.WriteString(w, `{"result":{"cashBuyingPower":"1000","currency":"KRW"}}`)
	default:
		http.Error(w, "unexpected path", http.StatusNotFound)
	}
}

func assertActualEngineRecoveryReusesTheStartupAccountSequence(
	t *testing.T,
	officialOptions ...official.Option,
) {
	t.Helper()
	var accountCalls atomic.Int32
	var ordersCalls atomic.Int32
	var holdingsCalls atomic.Int32
	var balanceCalls atomic.Int32
	var wrongHeaders atomic.Int32
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			writeEngineAccountSequenceToken(w)
		case "/api/v1/accounts":
			if accountCalls.Add(1) != 1 {
				w.WriteHeader(http.StatusTooManyRequests)
				return
			}
			writeEngineAccountSequenceAccounts(w,
				`[{"accountNo":"123-45","accountSeq":7,"accountType":"BROKERAGE"}]`)
		case "/api/v1/orders", "/api/v1/holdings", "/api/v1/buying-power":
			if r.Header.Get("X-Tossinvest-Account") != "7" {
				wrongHeaders.Add(1)
			}
			switch r.URL.Path {
			case "/api/v1/orders":
				ordersCalls.Add(1)
			case "/api/v1/holdings":
				holdingsCalls.Add(1)
			case "/api/v1/buying-power":
				balanceCalls.Add(1)
			}
			writeEmptyEngineSnapshotResponse(w, r.URL.Path)
		default:
			http.NotFound(w, r)
		}
	})

	ectx, _, srv, err := assembleAccountSequenceTestEngine(t, handler, officialOptions...)
	t.Cleanup(srv.Close)
	if err != nil {
		t.Fatalf("assembleEngine: %v", err)
	}
	t.Cleanup(func() { _ = ectx.Close() })
	if ectx.AccountRef != "123-45" {
		t.Fatalf("AccountRef = %q, want 123-45", ectx.AccountRef)
	}

	recovery, err := ectx.Recovery(reconcile.Options{
		Clock: clock.System(),
		Stabilise: reconcile.Stabilisation{
			Interval:    time.Millisecond,
			Required:    2,
			MaxAttempts: 2,
		},
	})
	if err != nil {
		t.Fatalf("Context.Recovery: %v", err)
	}
	report, err := recovery.Run(context.Background())
	if err != nil {
		t.Fatalf("Recovery.Run: %v", err)
	}
	if !report.Complete || !recovery.Complete() {
		t.Fatal("actual engine recovery did not complete")
	}
	if calls := accountCalls.Load(); calls != 1 {
		t.Fatalf("/api/v1/accounts calls = %d, want startup discovery only", calls)
	}
	if calls := wrongHeaders.Load(); calls != 0 {
		t.Fatalf("snapshot requests with wrong account header = %d", calls)
	}
	if got := []int32{ordersCalls.Load(), holdingsCalls.Load(), balanceCalls.Load()}; got[0] != 2 || got[1] != 2 || got[2] != 2 {
		t.Fatalf("snapshot endpoint calls = %v, want [2 2 2]", got)
	}
}

func TestActualEngineRecoveryReusesTheStartupAccountSequence(t *testing.T) {
	assertActualEngineRecoveryReusesTheStartupAccountSequence(t)
}

func TestActualEngineRecoveryAcceptsAMatchingExplicitAccountSequence(t *testing.T) {
	assertActualEngineRecoveryReusesTheStartupAccountSequence(t, official.WithAccountSeq(7))
}

func TestActualEngineRecoveryStillFailsClosedOnASnapshot429(t *testing.T) {
	var accountCalls atomic.Int32
	var ordersCalls atomic.Int32
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/oauth2/token":
			writeEngineAccountSequenceToken(w)
		case "/api/v1/accounts":
			accountCalls.Add(1)
			writeEngineAccountSequenceAccounts(w,
				`[{"accountNo":"123-45","accountSeq":7,"accountType":"BROKERAGE"}]`)
		case "/api/v1/orders":
			ordersCalls.Add(1)
			w.WriteHeader(http.StatusTooManyRequests)
		default:
			http.NotFound(w, r)
		}
	})

	ectx, _, srv, err := assembleAccountSequenceTestEngine(t, handler)
	t.Cleanup(srv.Close)
	if err != nil {
		t.Fatalf("assembleEngine: %v", err)
	}
	t.Cleanup(func() { _ = ectx.Close() })
	recovery, err := ectx.Recovery(reconcile.Options{
		Clock: clock.System(),
		Stabilise: reconcile.Stabilisation{
			Interval:    time.Millisecond,
			Required:    2,
			MaxAttempts: 2,
		},
	})
	if err != nil {
		t.Fatalf("Context.Recovery: %v", err)
	}
	report, err := recovery.Run(context.Background())
	if !errors.Is(err, reconcile.ErrRecoveryIncomplete) {
		t.Fatalf("Recovery.Run error = %v, want ErrRecoveryIncomplete", err)
	}
	if report.Complete || recovery.Complete() {
		t.Fatal("recovery accepted a partial snapshot after orders 429")
	}
	if !report.Snapshot.Empty() {
		t.Fatalf("partial snapshot was retained: %+v", report.Snapshot)
	}
	if calls := accountCalls.Load(); calls != 1 {
		t.Fatalf("/api/v1/accounts calls = %d, want 1", calls)
	}
	if calls := ordersCalls.Load(); calls != 1 {
		t.Fatalf("/api/v1/orders calls = %d, want 1", calls)
	}
}

func TestEngineRefusesAnIncompleteFirstAccountRecordBeforeOpeningTheJournal(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{
			name: "blank account number",
			body: `[{"accountNo":"","accountSeq":7},{"accountNo":"later","accountSeq":8}]`,
		},
		{
			name: "zero sequence",
			body: `[{"accountNo":"first","accountSeq":0},{"accountNo":"later","accountSeq":8}]`,
		},
		{
			name: "negative sequence",
			body: `[{"accountNo":"first","accountSeq":-7},{"accountNo":"later","accountSeq":8}]`,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/oauth2/token":
					writeEngineAccountSequenceToken(w)
				case "/api/v1/accounts":
					writeEngineAccountSequenceAccounts(w, tc.body)
				default:
					http.NotFound(w, r)
				}
			})
			ectx, dir, srv, err := assembleAccountSequenceTestEngine(t, handler)
			t.Cleanup(srv.Close)
			if ectx != nil {
				_ = ectx.Close()
			}
			if !errors.Is(err, engine.ErrAccountUnresolved) {
				t.Fatalf("assembleEngine error = %v, want ErrAccountUnresolved", err)
			}
			if _, statErr := os.Stat(filepath.Join(dir, journal.DBFileName)); !os.IsNotExist(statErr) {
				t.Fatalf("journal was created before account refusal: %v", statErr)
			}
		})
	}
}

func TestEngineRefusesAnExplicitSequenceThatDoesNotMatchTheFirstRecord(t *testing.T) {
	for _, seq := range []int{99, -1} {
		t.Run(strconv.Itoa(seq), func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/oauth2/token":
					writeEngineAccountSequenceToken(w)
				case "/api/v1/accounts":
					writeEngineAccountSequenceAccounts(w,
						`[{"accountNo":"123-45","accountSeq":7,"accountType":"BROKERAGE"}]`)
				default:
					http.NotFound(w, r)
				}
			})
			ectx, dir, srv, err := assembleAccountSequenceTestEngine(
				t, handler, official.WithAccountSeq(seq),
			)
			t.Cleanup(srv.Close)
			if ectx != nil {
				_ = ectx.Close()
			}
			if !errors.Is(err, engine.ErrAccountUnresolved) {
				t.Fatalf("assembleEngine error = %v, want ErrAccountUnresolved", err)
			}
			if _, statErr := os.Stat(filepath.Join(dir, journal.DBFileName)); !os.IsNotExist(statErr) {
				t.Fatalf("journal was created before sequence mismatch refusal: %v", statErr)
			}
		})
	}
}
