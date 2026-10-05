package verifylive

// reconcile_g2_test.go — a121 tasks 2.2·2.2.2: design G2(사건의 신원·선택·멱등·동시성·기록 엄격 해독·근거 지문).

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// --- 선택 -------------------------------------------------------------------------

// TestReconcileSelectsTheA063ShapedArtifactAsTheOneCandidate 는 tasks 2.2.2 의 양성 시험이다: 실패한
// conditional-cancel 이 보유를 푼 a063 형 artifact 가 유일한 후보다. 호출자는 id 를 주지 않는다.
func TestReconcileSelectsTheA063ShapedArtifactAsTheOneCandidate(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	res, _, _ := rcAccept(t, h)
	if res.Artifact.ID != rcTargetID {
		t.Fatalf("selected %q, want %q", res.Artifact.ID, rcTargetID)
	}
	if len(h.approvals) != 1 || h.approvals[0].ID != rcTargetID || h.approvals[0].Symbol != rcSymbol ||
		h.approvals[0].Kind != KindConditional {
		t.Fatalf("approval does not name the selected artifact: %+v", h.approvals)
	}
}

// TestReconcileRefusesZeroCandidates 는 후보 0 의 네 모양이다 — 모두 목록 읽기 전에 거절한다(spec).
func TestReconcileRefusesZeroCandidates(t *testing.T) {
	cases := map[string]func() []Entry{
		// 이미 종결(취소 확인)된 artifact 뿐.
		"nothing-outstanding": func() []Entry {
			e := rcA063Entries()
			e = append(e, Entry{Kind: KindCleanup, StepID: StepCleanup, AccountRef: rcMask(), Verdict: VerdictPass,
				Artifacts: []Artifact{{Kind: KindConditional, ID: rcTargetID, Symbol: rcSymbol, Cancelled: true}}})
			return e
		},
		// 보유가 풀리지 않은 조건주문 — conditional-cancel 판정이 그 줄 뒤에 없다.
		"hold-not-released": func() []Entry {
			e := rcA063Entries()
			return e[:3]
		},
		// 객체 유형 사례(재검 P2-d): outstanding 이 일반 주문 artifact 뿐.
		"order-kind-only": func() []Entry {
			e := rcA063Entries()
			e[1].Artifacts[0].Kind = KindOrder
			e[1].Artifacts[0].Deliberate = false
			return e
		},
		// M0 가 가리키는 artifact — 사람 대사 대상이라 자동 경로가 고르지 않는다.
		"m0-named": func() []Entry {
			e := rcA063Entries()
			return append(e, Entry{Kind: KindM0Checkpoint, AccountRef: rcMask(), M0Checkpoint: &M0Checkpoint{
				Kind: "parent-created", ClientOrderID: "client-1", ParentConditionalID: rcTargetID,
				Symbol: rcSymbol, Market: MarketKR}})
		},
	}
	for name, build := range cases {
		t.Run(name, func(t *testing.T) {
			h := newRCHarness(t, build())
			rcRefuse(t, h, RefuseNoCandidate)
			rcRequireNoListRead(t, h)
		})
	}
}

// TestReconcileRefusesSeveralCandidates — 보유가 풀린 조건주문이 둘이다.
func TestReconcileRefusesSeveralCandidates(t *testing.T) {
	e := rcA063Entries()
	second := rcTarget()
	second.ID = "CO-A063-SECOND"
	e[1].Artifacts = append(e[1].Artifacts, second)
	h := newRCHarness(t, e)
	rcRefuse(t, h, RefuseMultipleCandidates)
	rcRequireNoListRead(t, h)
}

// TestReconcileRefusesAnUnresolvedM0Owner — M0Unsettled 가 오류(복수 owner)면 무엇을 빼야 할지 알 수 없으므로 거절.
// (RED 로트 선택: design 은 "M0Unsettled 가 가리키는 artifact 제외" 만 적는다 — 그 판정이 실패한 경우는 fail-closed.)
func TestReconcileRefusesAnUnresolvedM0Owner(t *testing.T) {
	e := rcA063Entries()
	for _, client := range []string{"client-a", "client-b"} {
		e = append(e, Entry{Kind: KindM0Checkpoint, AccountRef: rcMask(), M0Checkpoint: &M0Checkpoint{
			Kind: "pending-create", ClientOrderID: client, Symbol: "000660", Market: MarketKR}})
	}
	h := newRCHarness(t, e)
	rcRefuse(t, h, RefuseM0Unresolved)
	rcRequireNoListRead(t, h)
}

// --- 멱등 -------------------------------------------------------------------------

// TestReconcileIsIdempotent 는 spec 「repeated reconciliation」 이다 — 둘째 요청은 선택 단계에서 "대사할 것 없음" 으로
// 거절하고, 줄을 더하지 않으며, 목록을 읽지 않는다.
func TestReconcileIsIdempotent(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	rcAccept(t, h)
	readsAfterFirst := len(h.reader.listCalls())
	before := h.bytes()
	_, err := h.run()
	rcRequireRefusal(t, err, RefuseNoCandidate)
	rcRequireUnchanged(t, h, before)
	if n := len(h.reader.listCalls()); n != readsAfterFirst {
		t.Fatalf("the second request read %d more list page(s)", n-readsAfterFirst)
	}
	entries, err := LoadEntries(h.path)
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, e := range entries {
		if e.Kind == KindReconcile {
			n++
		}
	}
	if n != 1 {
		t.Fatalf("want exactly one reconcile line after two requests, got %d", n)
	}
}

// --- 기록 엄격 해독 · 개행 꼬리 · 동시 변경 (F3 · R2-3 · P2-7) ------------------------

// TestReconcileRefusesATornFinalLineAtSelection 은 codex F3 이다 — LoadEntries 는 해독 불능 마지막 줄을 개행과 무관하게
// 버리지만, 대사는 모든 비공백 줄을 엄격 해독한다. 개행으로 끝나는 찢긴 줄도 거절.
func TestReconcileRefusesATornFinalLineAtSelection(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	rcAppendRaw(t, h.path, `{"format_version":1,"kind":"step","step_id":"conditional-ca`+"\n")
	if _, err := LoadEntries(h.path); err != nil {
		t.Fatalf("precondition: LoadEntries tolerates the torn tail, got %v", err)
	}
	rcRefuse(t, h, RefuseRecordUndecodable)
	rcRequireNoListRead(t, h)
}

// TestReconcileRefusesABrokenInnerLine — 중간 줄 깨짐.
func TestReconcileRefusesABrokenInnerLine(t *testing.T) {
	h := newRCHarness(t, nil)
	e := rcA063Entries()
	rcWriteEntries(t, h.path, e[:2])
	rcAppendRaw(t, h.path, "{not json}\n")
	rcWriteEntries(t, h.path, e[2:])
	rcRefuse(t, h, RefuseRecordUndecodable)
	rcRequireNoListRead(t, h)
}

// TestReconcileRefusesACompleteFinalLineWithoutNewline 은 codex R2-3 이다 — 개행 없는 완전한 JSON 꼬리는 엄격 해독을
// 통과하지만, 그 뒤에 붙이면 `{…}{…}` 로 둘 다 깨진다. 엄격 해독과 개행 검사가 **병존**해야 한다.
func TestReconcileRefusesACompleteFinalLineWithoutNewline(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	b, err := json.Marshal(Entry{FormatVersion: 1, Kind: KindStep, StepID: StepCosts, Verdict: VerdictSkipped, AccountRef: rcMask()})
	if err != nil {
		t.Fatal(err)
	}
	rcAppendRaw(t, h.path, string(b)) // 개행 없음
	rcRefuse(t, h, RefuseRecordTailNoNewline)
	rcRequireNoListRead(t, h)
}

// TestReconcileRefusesATornFinalLineBeforeAppend 는 F3 의 추가 직전 쪽이다 — 선택 뒤에 찢긴 줄이 생기면 추가 직전 엄격
// 해독이 거절한다(추가 직전 검사 순서: 엄격 해독 → 지문 → 개행 꼬리, RED 로트 선택).
func TestReconcileRefusesATornFinalLineBeforeAppend(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	h.reader.instrument = func(n int, symbol string) (string, error) {
		if n == 1 {
			rcAppendRaw(t, h.path, `{"format_version":1,"kind":"st`+"\n")
		}
		return symbol, nil
	}
	_, err := h.run()
	rcRequireRefusal(t, err, RefuseRecordUndecodable)
	if strings.Contains(string(h.bytes()), `"`+KindReconcile+`"`) {
		t.Fatal("a reconcile line was appended behind a torn line")
	}
}

// TestReconcileRefusesARecordThatChangedBeforeAppend 는 spec 「concurrent change to the record」 다.
func TestReconcileRefusesARecordThatChangedBeforeAppend(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	h.reader.instrument = func(n int, symbol string) (string, error) {
		if n == 1 {
			rcWriteEntries(t, h.path, []Entry{{Kind: KindStep, StepID: StepCosts, Verdict: VerdictSkipped, AccountRef: rcMask()}})
		}
		return symbol, nil
	}
	_, err := h.run()
	rcRequireRefusal(t, err, RefuseRecordChanged)
	if strings.Contains(string(h.bytes()), `"`+KindReconcile+`"`) {
		t.Fatal("a reconcile line was appended to a record that changed under it")
	}
}

// TestReconcileFingerprintIsTheFileBytes 는 freeze P2-7 이다 — 지문은 파일 원문 바이트의 sha256. 해독 결과가 같은
// 변경(빈 줄 추가)도 거절해야 한다. 해독한 Entry 의 지문은 이것을 못 본다.
func TestReconcileFingerprintIsTheFileBytes(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	h.reader.instrument = func(n int, symbol string) (string, error) {
		if n == 1 {
			rcAppendRaw(t, h.path, "\n")
		}
		return symbol, nil
	}
	_, err := h.run()
	rcRequireRefusal(t, err, RefuseRecordChanged)
}

// --- 줄의 모양 · 근거 지문 ---------------------------------------------------------

// TestReconcileLineMatchesTheDesignShape 는 design G2 「결정(골격)」·freeze P2-6·재검 P2-c 의 필드 고정이고, 투영 시험이
// 쓰는 손으로 만든 줄(rcReconcileLineJSON)이 실제 줄과 같은 모양임을 묶는다.
func TestReconcileLineMatchesTheDesignShape(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	_, line, raw := rcAccept(t, h)

	if v, _ := line["format_version"].(float64); int(v) != RecordFormatVersion {
		t.Fatalf("format_version %v, want %d (field addition, no version bump)", line["format_version"], RecordFormatVersion)
	}
	if line["step_id"] != string(StepReconcile) {
		t.Fatalf("step_id %v, want %q", line["step_id"], StepReconcile)
	}
	if line["account_ref"] != rcMask() {
		t.Fatalf("account_ref %v, want the record's mask %q", line["account_ref"], rcMask())
	}
	if v, ok := line["verdict"]; ok && v != "" {
		t.Fatalf("verdict %v: a reconcile line is not a verdict", v)
	}
	if line["mutating"] != false {
		t.Fatalf("mutating %v: the line records no broker mutation", line["mutating"])
	}
	if _, ok := line["calls"]; ok {
		t.Fatal("the reconcile line carries calls; its GETs would become successful-endpoint evidence")
	}
	obs, _ := line["observations"].([]any)
	if len(obs) != 1 {
		t.Fatalf("want exactly one observation (%s), got %v", ObservationReconcileBasis, obs)
	}
	o, _ := obs[0].(map[string]any)
	if o["key"] != ObservationReconcileBasis || !strings.HasPrefix(o["value"].(string), ReconcileBasisDomain) {
		t.Fatalf("basis observation %v, want key %q and a value tagged %q", o, ObservationReconcileBasis, ReconcileBasisDomain)
	}
	// 근거는 실제로 읽은 행 multiset 의 지문이다(접두사만이 아니라 값). 수락 픽스처는 세 그룹이 모두 빈 페이지라 읽은
	// 행이 0 개 — 그래서 이 경로의 근거는 빈 multiset 의 지문이다(비공집합 근거는
	// TestReconcileAcceptsAnExpiredRowOfAnotherIdentifierOnceTheAllowlistHasIt 가 잰다).
	if o["value"] != ReconcileBasisDigest(nil) {
		t.Fatalf("basis %v, want the digest of the rows actually read (empty) %q", o["value"], ReconcileBasisDigest(nil))
	}
	arts, _ := line["artifacts"].([]any)
	if len(arts) != 1 {
		t.Fatalf("want exactly one artifact, got %d", len(arts))
	}
	a, _ := arts[0].(map[string]any)
	want := rcTarget()
	for field, value := range map[string]any{"kind": want.Kind, "id": want.ID, "symbol": want.Symbol, "chain_id": want.ChainID} {
		if a[field] != value {
			t.Fatalf("artifact %s %v, want the outstanding line's %v", field, a[field], value)
		}
	}
	if a["reconciled_absent"] != true {
		t.Fatalf("reconciled_absent %v, want true", a["reconciled_absent"])
	}
	if c, _ := time.Parse(time.RFC3339Nano, a["created_at"].(string)); !c.Equal(want.CreatedAt) {
		t.Fatalf("created_at %v, want the outstanding line's %v (재검 P2-c)", a["created_at"], want.CreatedAt)
	}
	if r, _ := time.Parse(time.RFC3339Nano, a["reconciled_at"].(string)); !r.Equal(rcNow) {
		t.Fatalf("reconciled_at %v, want the append time %v", a["reconciled_at"], rcNow)
	}
	for _, forbidden := range []string{"cancelled_at", "filled", "filled_at", "held_until"} {
		if _, ok := a[forbidden]; ok {
			t.Fatalf("artifact carries %q: reconciled absence is neither a cancel, a fill nor a hold", forbidden)
		}
	}
	if a["cancelled"] == true {
		t.Fatal("reconciled absence written as a cancellation")
	}
	// 설계 모양과 실제 줄의 구조 동형: 같은 키 집합(값 무관).
	var hand map[string]any
	if err := json.Unmarshal([]byte(rcReconcileLineJSON(rcTarget(), rcNow, "x")), &hand); err != nil {
		t.Fatal(err)
	}
	for k := range hand {
		if k == "run_id" || k == "process" || k == "title" || k == "started_at" || k == "finished_at" {
			continue
		}
		if _, ok := line[k]; !ok && k != "verdict" {
			t.Fatalf("real line lacks %q that the projection tests assume", k)
		}
	}
	_ = raw
}

// TestReconcileLineAddsNoAccountIdentityBeyondTheMask 는 R2-5(digest 철회)·Q2 (a) 다 — 대사 줄은 마스크 외 계좌 신원을
// 싣지 않고, 대상 줄이 이미 담은 id 외의 브로커 식별자를 더하지 않는다.
func TestReconcileLineAddsNoAccountIdentityBeyondTheMask(t *testing.T) {
	h := newRCHarness(t, rcA063Entries())
	_, _, raw := rcAccept(t, h)
	digits := strings.NewReplacer("-", "").Replace(rcAccount)
	for _, leak := range []string{rcAccount, digits, digits[:len(digits)-4]} {
		if strings.Contains(raw, leak) {
			t.Fatalf("reconcile line carries account identity %q beyond the mask: %s", leak, raw)
		}
	}
	if strings.Contains(raw, "ORD-") {
		t.Fatalf("reconcile line carries a broker identifier the target line does not: %s", raw)
	}
}

// TestReconcileBasisDigestIsVersionedDomainTaggedAndOrderIndependent 은 tasks 2.2.2 마지막 요구다.
func TestReconcileBasisDigestIsVersionedDomainTaggedAndOrderIndependent(t *testing.T) {
	rows := []ReconcileBasisRow{
		{Group: ReconcileGroupConditionalClosed, ID: "CO-1", Status: "EXPIRED"},
		{Group: ReconcileGroupConditionalClosed, ID: "CO-2", Status: "EXPIRED"},
		{Group: ReconcileGroupPlainOpen, ID: "ORD-1", Status: "PENDING"},
	}
	d := ReconcileBasisDigest(rows)
	if !strings.HasPrefix(d, ReconcileBasisDomain) {
		t.Fatalf("basis %q does not carry its version/domain tag %q", d, ReconcileBasisDomain)
	}
	reversed := []ReconcileBasisRow{rows[2], rows[1], rows[0]}
	if ReconcileBasisDigest(reversed) != d {
		t.Fatal("basis digest depends on read order")
	}
	// multiset: 중복은 다른 근거다(집합으로 접지 않는다).
	if ReconcileBasisDigest(append(rows, rows[0])) == d {
		t.Fatal("basis digest folds a duplicate row (set, not multiset)")
	}
	// 같은 행이라도 그룹이 다르면 다른 근거.
	moved := append([]ReconcileBasisRow(nil), rows...)
	moved[0].Group = ReconcileGroupConditionalOpen
	if ReconcileBasisDigest(moved) == d {
		t.Fatal("basis digest ignores the group")
	}
	// 도메인 분리: 태그 없는 정규 직렬화의 지문과 같으면 안 된다.
	if strings.HasSuffix(d, Digest(rows)) || strings.HasSuffix(d, Digest(reversed)) {
		t.Fatal("basis digest is the untagged record digest (no domain separation)")
	}
	if ReconcileBasisDigest(nil) == "" {
		t.Fatal("an empty basis must still be a tagged digest")
	}
}

// --- 이름만 있던 거절 모양(A-RED 리뷰 P1-4) ------------------------------------------

// TestReconcileRefusesANewerRecordFormat 는 spec 「unsupported record schema SHALL fail closed」 다 — format_version 2
// 줄이 꼬리든 중간이든 해독 가능한 JSON 이어도 이 빌드가 이해하는 형식이 아니므로 거절(LoadEntries 의 errUnknownFormat 과 같은 판정).
func TestReconcileRefusesANewerRecordFormat(t *testing.T) {
	newer := `{"format_version":2,"kind":"step","step_id":"costs","verdict":"skipped"}` + "\n"
	for name, write := range map[string]func(h *rcHarness){
		"tail": func(h *rcHarness) {
			rcWriteEntries(t, h.path, rcA063Entries())
			rcAppendRaw(t, h.path, newer)
		},
		"middle": func(h *rcHarness) {
			e := rcA063Entries()
			rcWriteEntries(t, h.path, e[:2])
			rcAppendRaw(t, h.path, newer)
			rcWriteEntries(t, h.path, e[2:])
		},
	} {
		t.Run(name, func(t *testing.T) {
			h := newRCHarness(t, nil)
			write(h)
			rcRefuse(t, h, RefuseRecordFormat)
			rcRequireNoListRead(t, h)
		})
	}
}

// TestReconcileRefusesAnUnreadableRecord — 기록을 열 수 없으면(경로가 디렉터리) 부재 판정의 입력이 없으므로 거절.
func TestReconcileRefusesAnUnreadableRecord(t *testing.T) {
	h := newRCHarness(t, nil)
	h.path = filepath.Join(t.TempDir(), "record-is-a-directory")
	if err := os.Mkdir(h.path, 0o700); err != nil {
		t.Fatal(err)
	}
	_, err := h.run()
	rcRequireRefusal(t, err, RefuseRecordUnreadable)
	rcRequireNoListRead(t, h)
	if fi, statErr := os.Stat(h.path); statErr != nil || !fi.IsDir() {
		t.Fatal("the refusal touched the record path")
	}
}

// TestReconcileAcceptsAnExpiredRowOfAnotherIdentifierOnceTheAllowlistHasIt 는 A-RED 리뷰 P2-4 다 — 허용 목록이 (전사로)
// 채워지면 다른 id 의 EXPIRED 행(발동 흔적 없음)은 거절 사유가 아니다. 허용 목록 조회와 비공집합 multiset·근거 경로를
// 실제로 돌린다. 허용 목록 값은 시험 주입이다(출처는 여전히 부재 — 생산 값은 빈 목록).
func TestReconcileAcceptsAnExpiredRowOfAnotherIdentifierOnceTheAllowlistHasIt(t *testing.T) {
	prev := reconcileClosedStatusAllowlist
	reconcileClosedStatusAllowlist = map[string]bool{"EXPIRED": true}
	t.Cleanup(func() { reconcileClosedStatusAllowlist = prev })

	h := newRCHarness(t, rcA063Entries())
	h.reader.set(-1, ReconcileGroupConditionalClosed, rcCondPage("", false, rcCondRow("CO-OTHER-EXPIRED", "EXPIRED")))
	_, line, _ := rcAccept(t, h)
	obs, _ := line["observations"].([]any)
	if len(obs) != 1 {
		t.Fatalf("observations %v", obs)
	}
	got := obs[0].(map[string]any)["value"]
	want := ReconcileBasisDigest([]ReconcileBasisRow{
		{Group: ReconcileGroupConditionalClosed, ID: "CO-OTHER-EXPIRED", Status: "EXPIRED"},
	})
	if got != want {
		t.Fatalf("basis %v, want the digest of the one CLOSED row read %q", got, want)
	}
	if got == ReconcileBasisDigest(nil) {
		t.Fatal("a non-empty read produced the empty-basis digest")
	}
}
