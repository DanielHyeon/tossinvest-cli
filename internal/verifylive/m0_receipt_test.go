package verifylive

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/official"
)

func acquireReceiptLease(t *testing.T, receipt *CausalReceipt) *CausalReceiptLease {
	t.Helper()
	lease, err := receipt.AcquireRunLease()
	if err != nil {
		t.Fatalf("AcquireRunLease: %v", err)
	}
	return lease
}

func TestM0ClosedReceiptCannotFallBackToUnmeasuredTriggerRun(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	receipt, err := OpenCausalReceipt(dir)
	if err != nil {
		t.Fatal(err)
	}
	recorder, err := OpenRecorder(t.TempDir() + "/record.jsonl")
	if err != nil {
		t.Fatal(err)
	}
	defer recorder.Close()
	op := alwaysConfirm()
	prior := make([]Entry, 0, len(Steps()))
	for _, step := range Steps() {
		if step.ID != StepConditionalTrigger {
			prior = append(prior, Entry{Kind: KindStep, StepID: step.ID, Verdict: VerdictPass})
		}
	}
	runner, err := New(Options{Broker: official.New(official.Credentials{APIKey: "k", SecretKey: "s"}, t.TempDir()+"/token", official.WithBaseURL("http://127.0.0.1")), Recorder: recorder,
		Confirm: op.confirmer(), ConfirmBatch: op.batchConfirmer(), AccountRef: "1", IncludeTrigger: true,
		ConfirmEach: true, Resume: true, Redo: []StepID{StepConditionalTrigger}, Receipt: receipt, Prior: prior})
	if err != nil {
		t.Fatal(err)
	}
	if err := receipt.Close(); err != nil {
		t.Fatal(err)
	}
	summary, err := runner.Run(context.Background())
	if err == nil || !summary.Halted {
		t.Fatalf("closed receipt Run = %+v err=%v, want terminal HOLD", summary, err)
	}
}

func TestM0NewRejectsTriggerWithoutExactReceiptMode(t *testing.T) {
	h := triggerHarness(t, nil)
	r, closeRecord, err := h.build(Options{HoldingSymbol: "005930", IncludeTrigger: true})
	if closeRecord != nil {
		defer closeRecord()
	}
	if err == nil {
		_ = r
		t.Fatal("New accepted include-trigger without M0 receipt, confirm-each, resume, and exact redo")
	}
}

func TestM0ReceiptReadinessCanOnlyBeMintedByTheSecureConstructor(t *testing.T) {
	src, err := os.ReadFile("receipt.go")
	if err != nil {
		t.Fatal(err)
	}
	f, err := parser.ParseFile(token.NewFileSet(), "receipt.go", src, 0)
	if err != nil {
		t.Fatal(err)
	}
	assignments := 0
	owner := ""
	ast.Inspect(f, func(n ast.Node) bool {
		assign, ok := n.(*ast.AssignStmt)
		if !ok {
			return true
		}
		for i, lhs := range assign.Lhs {
			if sel, ok := lhs.(*ast.SelectorExpr); ok && sel.Sel.Name == "ready" && i < len(assign.Rhs) {
				value, isTrue := assign.Rhs[i].(*ast.Ident)
				if !isTrue || value.Name != "true" {
					continue
				}
				assignments++
				owner = "OpenCausalReceipt"
			}
		}
		return true
	})
	if assignments != 1 || owner != "OpenCausalReceipt" {
		t.Fatalf("receipt readiness assignments = %d owner=%q; only secure constructor may mint it", assignments, owner)
	}
}

func TestM0ReceiptCannotBeForgedByAnExternalZeroValue(t *testing.T) {
	h := triggerHarness(t, nil)
	_, closeRecord, err := h.build(Options{
		HoldingSymbol: "005930", IncludeTrigger: true, ConfirmEach: true, Resume: true,
		Redo: []StepID{StepConditionalTrigger}, Receipt: &CausalReceipt{},
	})
	if closeRecord != nil {
		defer closeRecord()
	}
	if err == nil {
		t.Fatal("New accepted a zero-value causal receipt; only OpenCausalReceipt may create receipt authority")
	}
}

func TestM0ReceiptHeaderIsFreshAndDurable(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	r, err := OpenCausalReceipt(dir)
	if err != nil {
		t.Fatalf("OpenCausalReceipt: %v", err)
	}
	defer r.Close()
	if r.RunID() == "" || !r.ready {
		t.Fatalf("receipt = %+v, want fresh ready run", r)
	}
	info, err := os.Stat(filepath.Join(dir, r.RunID()+".jsonl"))
	if err != nil {
		t.Fatalf("stat receipt: %v", err)
	}
	if !info.Mode().IsRegular() || info.Mode().Perm() != 0o600 {
		t.Fatalf("receipt mode/type = %v, want regular 0600", info.Mode())
	}
	lease := acquireReceiptLease(t, r)
	defer lease.Release()
	if _, err := lease.RecordCausal("test", m0CausalFieldsV1{}); err != nil {
		t.Fatalf("Append: %v", err)
	}
	r2, err := OpenCausalReceipt(dir)
	if err != nil {
		t.Fatalf("second receipt should use a fresh O_EXCL path: %v", err)
	}
	defer r2.Close()
}

func TestM0ReceiptClosedAfterHeaderFailureIsUnusable(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	r, err := OpenCausalReceipt(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := r.AcquireRunLease(); err == nil {
		t.Fatal("closed receipt minted write authority")
	}
}

func TestM0ReceiptRejectsInsecureAndSymlinkParents(t *testing.T) {
	insecure := t.TempDir()
	if err := os.Chmod(insecure, 0o755); err != nil {
		t.Fatal(err)
	}
	if _, err := OpenCausalReceipt(insecure); err == nil {
		t.Fatal("receipt accepted a non-0700 parent")
	}
	target := t.TempDir()
	if err := os.Chmod(target, 0o700); err != nil {
		t.Fatal(err)
	}
	link := filepath.Join(t.TempDir(), "receipt-link")
	if err := os.Symlink(target, link); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	if _, err := OpenCausalReceipt(link); err == nil {
		t.Fatal("receipt accepted a symlink parent")
	}
	realAncestor := filepath.Join(t.TempDir(), "real")
	if err := os.Mkdir(realAncestor, 0o700); err != nil {
		t.Fatal(err)
	}
	leaf := filepath.Join(realAncestor, "receipt")
	if err := os.Mkdir(leaf, 0o700); err != nil {
		t.Fatal(err)
	}
	ancestorLink := filepath.Join(t.TempDir(), "ancestor-link")
	if err := os.Symlink(realAncestor, ancestorLink); err != nil {
		t.Skipf("ancestor symlink unavailable: %v", err)
	}
	if _, err := OpenCausalReceipt(filepath.Join(ancestorLink, "receipt")); err == nil {
		t.Fatal("receipt accepted a symlinked intermediate ancestor")
	}
}

func TestM0ReceiptPublicSurfaceHasNoArbitraryAppendAuthority(t *testing.T) {
	for _, tc := range []struct {
		name    string
		typ     reflect.Type
		allowed map[string]bool
	}{
		{name: "CausalReceipt", typ: reflect.TypeOf((*CausalReceipt)(nil)), allowed: map[string]bool{
			"AcquireRunLease": true, "Close": true, "RunID": true,
		}},
		{name: "CausalReceiptLease", typ: reflect.TypeOf((*CausalReceiptLease)(nil)), allowed: map[string]bool{
			"RecordAttempt": true, "RecordCausal": true, "Release": true, "RunID": true,
		}},
	} {
		seen := map[string]bool{}
		for i := 0; i < tc.typ.NumMethod(); i++ {
			name := tc.typ.Method(i).Name
			seen[name] = true
			if !tc.allowed[name] {
				t.Errorf("%s exports unexpected method %s", tc.name, name)
			}
		}
		for name := range tc.allowed {
			if !seen[name] {
				t.Errorf("%s exact public API is missing %s", tc.name, name)
			}
		}
	}
}

func TestM0ReceiptTypedWritesRequireTheActiveMatchingLease(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	r, err := OpenCausalReceipt(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	stale := acquireReceiptLease(t, r)
	stale.Release()
	active := acquireReceiptLease(t, r)
	defer active.Release()
	if _, err := stale.RecordCausal("stale-authority", m0CausalFieldsV1{}); err == nil {
		t.Fatal("released lease retained receipt write authority")
	}
	if _, err := active.RecordCausal("active-authority", m0CausalFieldsV1{}); err != nil {
		t.Fatalf("active lease typed write: %v", err)
	}
}

func TestM0ReceiptRunSequencesNeverMerge(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	first, err := OpenCausalReceipt(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer first.Close()
	firstLease := acquireReceiptLease(t, first)
	defer firstLease.Release()
	second, err := OpenCausalReceipt(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer second.Close()
	secondLease := acquireReceiptLease(t, second)
	defer secondLease.Release()
	a, err := firstLease.RecordCausal("first", m0CausalFieldsV1{})
	if err != nil {
		t.Fatal(err)
	}
	b, err := secondLease.RecordCausal("second", m0CausalFieldsV1{})
	if err != nil {
		t.Fatal(err)
	}
	if a.RunID == b.RunID || a.Sequence != 1 || b.Sequence != 1 {
		t.Fatalf("receipt runs merged: first=%+v second=%+v", a, b)
	}
}

func TestM0ReceiptFileAndDirectorySyncFailuresLeaveNoUsableReceipt(t *testing.T) {
	secured := func(t *testing.T) string {
		t.Helper()
		dir := t.TempDir()
		if err := os.Chmod(dir, 0o700); err != nil {
			t.Fatal(err)
		}
		return dir
	}
	fixedRandom := func(b []byte) (int, error) {
		for i := range b {
			b[i] = 0
		}
		return len(b), nil
	}
	for _, tc := range []struct {
		name  string
		hooks receiptHooks
	}{
		{name: "header-file-sync", hooks: receiptHooks{random: fixedRandom, syncFile: func(*os.File) error { return errors.New("file sync kill") }}},
		{name: "header-directory-sync", hooks: receiptHooks{random: fixedRandom, syncDir: func(*os.File) error { return errors.New("dir sync kill") }}},
		{name: "final-wrong-mode", hooks: receiptHooks{random: fixedRandom, afterCreate: func(path string) error { return os.Chmod(path, 0o640) }}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if receipt, err := openCausalReceipt(secured(t), tc.hooks); err == nil || receipt != nil {
				t.Fatalf("receipt=%+v err=%v, want unusable failure", receipt, err)
			}
		})
	}
	t.Run("exclusive-collision", func(t *testing.T) {
		dir := secured(t)
		path := filepath.Join(dir, "m0-00000000000000000000000000000000.jsonl")
		if err := os.WriteFile(path, []byte("existing"), 0o600); err != nil {
			t.Fatal(err)
		}
		if receipt, err := openCausalReceipt(dir, receiptHooks{random: fixedRandom}); err == nil || receipt != nil {
			t.Fatalf("O_EXCL collision receipt=%+v err=%v, want failure", receipt, err)
		}
	})
}

func TestM0ReceiptDirFDPreventsParentSwapRedirect(t *testing.T) {
	root := t.TempDir()
	secured := filepath.Join(root, "secured")
	attacker := filepath.Join(root, "attacker")
	parked := filepath.Join(root, "parked")
	for _, dir := range []string{secured, attacker} {
		if err := os.Mkdir(dir, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	fixedRandom := func(b []byte) (int, error) {
		for i := range b {
			b[i] = 0
		}
		return len(b), nil
	}
	r, err := openCausalReceipt(secured, receiptHooks{
		random: fixedRandom,
		afterDirOpen: func(path string) error {
			if err := os.Rename(path, parked); err != nil {
				return err
			}
			return os.Symlink(attacker, path)
		},
	})
	if err != nil {
		t.Fatalf("open after parent swap: %v", err)
	}
	defer r.Close()
	lease := acquireReceiptLease(t, r)
	defer lease.Release()
	stamp, err := lease.RecordCausal("after-parent-swap", m0CausalFieldsV1{})
	if err != nil {
		t.Fatalf("record after parent swap: %v", err)
	}
	if stamp.RunID != r.RunID() {
		t.Fatalf("event run ID = %q, header run ID = %q", stamp.RunID, r.RunID())
	}
	name := r.RunID() + ".jsonl"
	body, err := os.ReadFile(filepath.Join(parked, name))
	if err != nil {
		t.Fatalf("receipt was not created in opened directory: %v", err)
	}
	if !strings.Contains(string(body), `"run_id":"`+r.RunID()+`"`) || !strings.Contains(string(body), `"after-parent-swap"`) {
		t.Fatalf("header and event not retained by original directory: %s", body)
	}
	if _, err := os.Stat(filepath.Join(attacker, name)); !os.IsNotExist(err) {
		t.Fatalf("parent swap redirected receipt to attacker directory: %v", err)
	}
}

func TestM0ReceiptLeaseRejectsCloseUntilRunReleasesIt(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	r, err := OpenCausalReceipt(dir)
	if err != nil {
		t.Fatal(err)
	}
	lease, err := r.AcquireRunLease()
	if err != nil {
		t.Fatalf("AcquireRunLease: %v", err)
	}
	if lease.RunID() != r.RunID() {
		t.Fatalf("lease run ID = %q, receipt run ID = %q", lease.RunID(), r.RunID())
	}
	if second, err := r.AcquireRunLease(); !errors.Is(err, ErrCausalReceiptLeased) || second != nil {
		t.Fatalf("second AcquireRunLease = lease=%v err=%v, want ErrCausalReceiptLeased", second, err)
	}
	if err := r.Close(); !errors.Is(err, ErrCausalReceiptLeased) {
		t.Fatalf("Close during lease = %v, want ErrCausalReceiptLeased", err)
	}
	if _, err := lease.RecordCausal("while-leased", m0CausalFieldsV1{}); err != nil {
		t.Fatalf("record while leased: %v", err)
	}
	lease.Release()
	if err := r.Close(); err != nil {
		t.Fatalf("Close after lease release: %v", err)
	}
}

func TestM0ReceiptLeaseIsExclusiveUnderConcurrentAcquire(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	r, err := OpenCausalReceipt(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	first, err := r.AcquireRunLease()
	if err != nil {
		t.Fatal(err)
	}
	defer first.Release()

	const contenders = 16
	errs := make(chan error, contenders)
	var wg sync.WaitGroup
	for range contenders {
		wg.Add(1)
		go func() {
			defer wg.Done()
			lease, err := r.AcquireRunLease()
			if lease != nil {
				lease.Release()
				errs <- errors.New("concurrent lease unexpectedly acquired")
				return
			}
			errs <- err
		}()
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		if !errors.Is(err, ErrCausalReceiptLeased) {
			t.Fatalf("concurrent AcquireRunLease error = %v, want ErrCausalReceiptLeased", err)
		}
	}
}

func TestM0ReceiptCloseBeforeLeaseAndDuringAppendAreSafe(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	r, err := OpenCausalReceipt(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := r.AcquireRunLease(); err == nil {
		t.Fatal("AcquireRunLease accepted a closed receipt")
	}

	r, err = OpenCausalReceipt(dir)
	if err != nil {
		t.Fatal(err)
	}
	lease, err := r.AcquireRunLease()
	if err != nil {
		t.Fatal(err)
	}
	entered := make(chan struct{})
	releaseSync := make(chan struct{})
	r.sync = func(*os.File) error {
		close(entered)
		<-releaseSync
		return nil
	}
	appendDone := make(chan error, 1)
	go func() {
		_, err := lease.RecordCausal("concurrent-append", m0CausalFieldsV1{})
		appendDone <- err
	}()
	<-entered
	closeDone := make(chan error, 1)
	go func() { closeDone <- r.Close() }()
	close(releaseSync)
	if err := <-appendDone; err != nil {
		t.Fatalf("append during close: %v", err)
	}
	if err := <-closeDone; !errors.Is(err, ErrCausalReceiptLeased) {
		t.Fatalf("Close during active lease = %v, want ErrCausalReceiptLeased", err)
	}
	lease.Release()
	if err := r.Close(); err != nil {
		t.Fatalf("final Close: %v", err)
	}
}

func TestM0ReceiptWriteAndSyncFailuresPoisonAllLaterContinuity(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name string
		fail func(*CausalReceipt)
	}{
		{name: "write", fail: func(r *CausalReceipt) { _ = r.f.Close() }},
		{name: "sync", fail: func(r *CausalReceipt) {
			r.sync = func(*os.File) error { return errors.New("append sync failure") }
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			r, err := OpenCausalReceipt(dir)
			if err != nil {
				t.Fatal(err)
			}
			lease, err := r.AcquireRunLease()
			if err != nil {
				t.Fatal(err)
			}
			tc.fail(r)
			if _, err := lease.RecordCausal("must-not-be-durable", m0CausalFieldsV1{}); !errors.Is(err, ErrCausalReceiptPoisoned) {
				t.Fatalf("failed persistence error = %v, want ErrCausalReceiptPoisoned", err)
			}
			if _, err := lease.RecordCausal("must-never-follow-gap", m0CausalFieldsV1{}); !errors.Is(err, ErrCausalReceiptPoisoned) {
				t.Fatalf("post-gap write error = %v, want latched ErrCausalReceiptPoisoned", err)
			}
			lease.Release()
			if next, err := r.AcquireRunLease(); !errors.Is(err, ErrCausalReceiptPoisoned) || next != nil {
				t.Fatalf("poisoned receipt reacquired: lease=%v err=%v, want ErrCausalReceiptPoisoned", next, err)
			}
			_ = r.Close()
		})
	}
}

func TestM0ReceiptHashesSuccessfullyReadEmptyBody(t *testing.T) {
	dir := t.TempDir()
	if err := os.Chmod(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	r, err := OpenCausalReceipt(dir)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	lease := acquireReceiptLease(t, r)
	defer lease.Release()
	now := time.Now()
	if _, err := lease.RecordAttempt("empty-response", official.AttemptTrace{
		RequestStart: now, BodyReadComplete: now.Add(time.Millisecond), StatusCode: 204, Body: []byte{},
	}); err == nil || !strings.Contains(err.Error(), "invalid 2xx envelope") {
		t.Fatalf("empty 2xx response error = %v, want recorded invalid-envelope evidence", err)
	}
	body, err := os.ReadFile(filepath.Join(dir, r.RunID()+".jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	empty := sha256.Sum256(nil)
	want := `"sha256_raw_bytes_v1":"` + hex.EncodeToString(empty[:]) + `"`
	if !strings.Contains(string(body), want) {
		t.Fatalf("successfully read empty body lacks digest %s: %s", want, body)
	}
}

func TestM0ReceiptOrderingUsesSequenceAndElapsedNotSerializedWallTime(t *testing.T) {
	source, err := os.ReadFile("receipt.go")
	if err != nil {
		t.Fatal(err)
	}
	text := string(source)
	for _, forbidden := range []string{"`json:\"at\"`", "time.Time `json:"} {
		if strings.Contains(text, forbidden) {
			t.Fatalf("receipt serializes wall-time authority: %s", forbidden)
		}
	}
	for _, required := range []string{"Sequence  uint64", "ElapsedNS int64", "r.seq++", "time.Since(r.start)", "trace.RequestStart.Sub(r.start)"} {
		if !strings.Contains(text, required) {
			t.Fatalf("receipt ordering evidence missing %q", required)
		}
	}
	linuxSource, err := os.ReadFile("receipt_linux.go")
	if err != nil {
		t.Fatal(err)
	}
	linuxText := string(linuxSource)
	for _, ownerEvidence := range []string{"unix.Openat", "unix.O_NOFOLLOW", "unix.Fstat", "unix.S_IFREG", "exactPrivateMode(st.Mode, 0o600)", "currentReceiptUID()"} {
		if !strings.Contains(linuxText, ownerEvidence) {
			t.Fatalf("receipt final fstat owner/mode validation missing %q", ownerEvidence)
		}
	}
}
