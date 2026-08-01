package httpapi

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"
)

func TestSecurityStoreConsumesCapabilityAndAuditsBeforeReservationReturns(t *testing.T) {
	t.Parallel()
	store, err := OpenSecurityStore(securityStoreTestPath(t))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	req := ledgerRequest("nonce-0123456789abcdef", "idem-0123456789abcdef", digestOf(`{"preset":"safe"}`))
	reservation, err := store.Reserve(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if reservation.ID == 0 || reservation.Replay != nil {
		t.Fatalf("reservation=%+v", reservation)
	}
	var auditCount, nonceCount int
	if err := store.db.QueryRow(`SELECT count(*) FROM mutation_audit WHERE stage='authorized'`).Scan(&auditCount); err != nil {
		t.Fatal(err)
	}
	if err := store.db.QueryRow(`SELECT count(*) FROM consumed_capability_nonces`).Scan(&nonceCount); err != nil {
		t.Fatal(err)
	}
	if auditCount != 1 || nonceCount != 1 {
		t.Fatalf("audit/nonce=%d/%d before command", auditCount, nonceCount)
	}
	if _, err := store.Reserve(context.Background(), req); !errors.Is(err, ErrCapabilitySpent) {
		t.Fatalf("reused capability error=%v", err)
	}
}

func TestSecurityStoreIdempotencyReplayAndBodyConflict(t *testing.T) {
	t.Parallel()
	store, _ := OpenSecurityStore(securityStoreTestPath(t))
	t.Cleanup(func() { _ = store.Close() })
	req := ledgerRequest("", "idem-0123456789abcdef", digestOf(`{"preset":"safe"}`))
	reservation, err := store.Reserve(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := store.Reserve(context.Background(), req); !errors.Is(err, ErrIdempotencyInProgress) {
		t.Fatalf("pending replay error=%v", err)
	}
	want := StoredMutationResponse{Status: 204, Version: `"8"`, Body: []byte(`{"ok":true}`)}
	if err := store.Complete(context.Background(), reservation.ID, want); err != nil {
		t.Fatal(err)
	}
	if err := store.Complete(context.Background(), reservation.ID, want); err != nil {
		t.Fatalf("idempotent completion retry: %v", err)
	}
	replayed, err := store.Reserve(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	if replayed.Replay == nil || replayed.Replay.Status != want.Status || string(replayed.Replay.Body) != string(want.Body) {
		t.Fatalf("replay=%+v", replayed.Replay)
	}
	conflict := req
	conflict.BodyDigest = digestOf(`{"preset":"other"}`)
	if _, err := store.Reserve(context.Background(), conflict); !errors.Is(err, ErrIdempotencyConflict) {
		t.Fatalf("body conflict error=%v", err)
	}
}

func TestCapabilityNonceConsumeIsAtomicAcrossConcurrentRequests(t *testing.T) {
	store, _ := OpenSecurityStore(securityStoreTestPath(t))
	t.Cleanup(func() { _ = store.Close() })
	const workers = 8
	var wg sync.WaitGroup
	results := make(chan error, workers)
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req := ledgerRequest("nonce-race-0123456789", "idem-race-0123456789", digestOf(`{"preset":"safe"}`))
			_, err := store.Reserve(context.Background(), req)
			results <- err
		}()
	}
	wg.Wait()
	close(results)
	success, spent := 0, 0
	for err := range results {
		switch {
		case err == nil:
			success++
		case errors.Is(err, ErrCapabilitySpent):
			spent++
		default:
			t.Fatalf("unexpected concurrent error: %v", err)
		}
	}
	if success != 1 || spent != workers-1 {
		t.Fatalf("success/spent=%d/%d", success, spent)
	}
}

func securityStoreTestPath(t *testing.T) string {
	t.Helper()
	directory := t.TempDir()
	if err := os.Chmod(directory, 0o700); err != nil {
		t.Fatal(err)
	}
	return filepath.Join(directory, "httpapi-security.db")
}

func ledgerRequest(nonce, key, digest string) MutationLedgerRequest {
	return MutationLedgerRequest{
		Identity: MutationIdentity{Actor: "operator:local", Client: "ios:device-a", Mode: AuthModeCapability},
		Method:   "POST", Resource: "/api/v1/optimization/previews", BodyDigest: digest,
		IdempotencyKey: key, IfMatch: `"7"`, CapabilityNonce: nonce,
		At: time.Date(2026, 8, 1, 1, 2, 3, 0, time.UTC),
	}
}
